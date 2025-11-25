package implement

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"sync"
	"time"

	"github.com/chencheng8888/GoDo/dao"
	"github.com/chencheng8888/GoDo/scheduler/domain"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	RedisPrefix        = "distributed_scheduler"
	RedisZSetKey       = RedisPrefix + ":" + "zset"
	RedisTaskKeyPrefix = RedisPrefix + ":" + "task"
)

// TODO: 加入分布式锁

type DistributedScheduler struct {
	parser      cron.Parser        // 解析cron表达式
	rdb         *redis.Client      // redis客户端
	taskInfoDao *dao.TaskInfoDao   // 任务信息DAO
	log         *zap.SugaredLogger // 日志记录器
	pool        *ants.Pool         // goroutine池
	executor    domain.Executor    // 任务执行器

	//TODO: 这边或许可以改造下，把taskEntries改成map，可以加速 "从redis的zset中获取task的payload后解析成task"的过程
	taskEntries []domain.Task // 在running前记录task
	running     bool          // 是否已经run
	runningMu   sync.Mutex    // 保护running的互斥锁

	scheduleMapping   map[string]cron.Schedule // cron表达式到schedule的映射，加速解析
	scheduleMappingMu sync.RWMutex             // 保护scheduleMapping的互斥锁

	// lua脚本相关
	registerTaskScript *redis.Script
	removeTaskScript   *redis.Script
	getTaskScript      *redis.Script

	addTaskChan    chan domain.Task // 添加任务信号
	removeTaskChan chan string      // 删除任务信号

	schedulerCtx context.Context    // 所有job执行的上下文
	cancelFunc   context.CancelFunc // 取消所有job的函数，与schedulerCtx相对应

	location *time.Location // 时区
}

func (d *DistributedScheduler) registerScripts() {
	d.registerTaskScript = redis.NewScript(registerTaskScriptStr)
	d.removeTaskScript = redis.NewScript(removeTaskScriptStr)
	d.getTaskScript = redis.NewScript(getTaskScriptStr)
}

func (d *DistributedScheduler) AddTask(t domain.Task) error {
	return d.addTask(t, false)
}

func (d *DistributedScheduler) addTask(t domain.Task, addRedisOnly bool) error {
	_, err := d.parseCronSchedule(t.GetScheduledTime())
	if err != nil {
		return fmt.Errorf("failed to parse schedule time: %w", err)
	}

	if !addRedisOnly {
		// 添加数据库失败
		err = d.taskInfoDao.CreateTaskInfo(newModel(t))
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("task id %s already exists in db", t.GetID())
			}
			return err
		}
	}

	if d.running {
		d.addTaskChan <- t
	} else {
		d.taskEntries = append(d.taskEntries, t)
	}

	d.log.Infof("add a task successfully: %+v", t)
	return nil
}

func (d *DistributedScheduler) ListTasks(userName string) []domain.Task {
	var tasks []domain.Task
	taskInfos, err := d.taskInfoDao.GetTaskInfosByOwnerName(userName)
	if err != nil {
		d.log.Errorf("get task info by owner_name=%s error: %s", userName, err)
		return []domain.Task{}
	}

	for _, taskInfo := range taskInfos {
		task, err := domain.NewTaskFromModel(taskInfo)
		if err != nil {
			d.log.Errorf("new task from model failed: %s", err)
			continue
		}
		tasks = append(tasks, task)
	}
	d.log.Infof("📋 Listed tasks for user:%s,res is [%v]", userName, tasks)
	return tasks
}

func (d *DistributedScheduler) RemoveTask(userName string, taskId string) error {
	err := d.taskInfoDao.DeleteTaskInfoByTaskId(userName, taskId)
	if err != nil {
		d.log.Errorf("delete task info by (user_name=%s and task_id=%s) error: %s", userName, taskId, err)
		return err
	}
	d.removeTaskChan <- taskId
	d.log.Infof("remove task[user_name:%v,taskId:%v] successfully", userName, taskId)

	return nil
}

func (d *DistributedScheduler) Start() {
	d.runningMu.Lock()
	if d.running {
		d.runningMu.Unlock()
		return
	}
	d.running = true
	d.runningMu.Unlock()
	d.run()
}

func (d *DistributedScheduler) run() {
	now := d.now()
	for _, t := range d.taskEntries {
		d.addTaskToRedis(t, now)
	}

	for {
		select {
		case <-d.schedulerCtx.Done():
			d.log.Infof("distributed scheduler stopped")
			return
		default:
		}

		res, err := d.rdb.ZRangeWithScores(d.schedulerCtx, RedisZSetKey, 0, 0).Result()
		if err != nil {
			d.log.Errorf("get redis zset failed: %s", err)
			if errors.Is(err, context.Canceled) {
				d.log.Infof("distributed scheduler stopped")
				return
			}
			time.Sleep(time.Second)
			continue
		}
		var timer *time.Timer
		if len(res) == 0 {
			timer = time.NewTimer(100000 * time.Hour)
		} else {
			nextTimeUnix := int64(res[0].Score)
			nextTime := time.Unix(nextTimeUnix, 0).In(d.location)
			dur := nextTime.Sub(now)
			if dur < 0 {
				dur = 0
			}
			timer = time.NewTimer(dur)
		}

		select {
		case now = <-timer.C:
			now = now.In(d.location)
			tasks := d.getTaskFromRedis(now.Unix())
			if len(tasks) > 0 {
				for _, t := range tasks {
					// 执行task，直接提交task到协程池中去
					d.executeTask(t)
					// 执行完task后需要计算下次运行时间，并把task重新加入到redis中
					// 这里选择使用协程池来执行，防止阻塞主调度协程
					d.resubmitTask(t, now)
				}
			}

		case newTask := <-d.addTaskChan:
			timer.Stop()
			now = d.now()
			d.addTaskToRedis(newTask, now)

		case removeTaskId := <-d.removeTaskChan:
			timer.Stop()
			now = d.now()
			d.removeTaskFromRedis(removeTaskId)
		case <-d.schedulerCtx.Done():
			timer.Stop()
			d.log.Infof("distributed scheduler stopped")
			return
		}
	}
}

func (d *DistributedScheduler) addTaskToRedis(task domain.Task, now time.Time) {
	schedule, _ := d.parseCronSchedule(task.GetScheduledTime())
	payload, err := d.taskToJson(task)
	if err != nil {
		d.log.Errorf("failed to marshal task to json: %v", err)
		return
	}

	nextTime := schedule.Next(now)

	hashKey := d.generateHashKey(task.GetID())

	res, err := d.registerTaskScript.Run(d.schedulerCtx, d.rdb,
		[]string{hashKey, RedisZSetKey}, nextTime.Unix(), hashKey, payload).Result()
	if err != nil {
		d.log.Errorf("failed to add task to redis: %v", err)
		return
	}

	result, ok := res.(int64)
	if !ok {
		d.log.Errorf("assert result failed after excute add task lua script")
		return
	}
	if result == 0 {
		d.log.Errorf("task with ID %s already exists", task.GetID())
		return
	}
	d.log.Infof("add task[%+v] to redis successfully ,now:%v,next:%v", task, now, nextTime)
}

func (d *DistributedScheduler) removeTaskFromRedis(taskId string) {
	hashKey := d.generateHashKey(taskId)

	// 使用lua脚本来执行删除task的操作
	res, err := d.removeTaskScript.Run(d.schedulerCtx, d.rdb, []string{hashKey, RedisZSetKey}, hashKey).Result()
	if err != nil {
		d.log.Errorf("failed to remove task from redis: %v", err)
		return
	}

	result, ok := res.(int64)
	if !ok {
		d.log.Errorf("assert result failed after excute remove task lua script")
		return
	}

	if result == 0 {
		d.log.Errorf("taskId is not in redis,remove task failed")
		return
	}
	d.log.Infof("remove taskId[%s] from redis successfully", taskId)
}

func (d *DistributedScheduler) getTaskFromRedis(score int64) []domain.Task {
	// TODO: 这边的limit可以尝试配置化
	payloads, err := d.getTaskScript.Run(d.schedulerCtx, d.rdb, []string{RedisZSetKey}, score, 10).Result()
	if err != nil {
		d.log.Errorf("get task from redis failed: %s", err)
		return nil
	}
	var tasks []domain.Task
	payloadList, ok := payloads.([]string)
	if !ok {
		d.log.Errorf("assert payloads failed after excute get task lua script")
		return nil
	}
	for _, payload := range payloadList {
		task, err := d.jsonToTask(payload)
		if err != nil {
			d.log.Errorf("getTaskFromRedis: unmarshal task from json failed: %s", err)
			continue
		}
		tasks = append(tasks, task)
	}
	return tasks
}

func (d *DistributedScheduler) executeTask(t domain.Task) {
	err := d.pool.Submit(func() {
		d.executor(d.schedulerCtx, t)
	})
	if err != nil {
		d.log.Errorf("submit task to pool failed: %s,task:%v", err, t)
	}
}

func (d *DistributedScheduler) resubmitTask(t domain.Task, now time.Time) {
	err := d.pool.Submit(func() {
		d.addTaskToRedis(t, now)
	})
	if err != nil {
		d.log.Errorf("submit 'add task to redis' failed: %s", err)
	}
}

func (d *DistributedScheduler) Stop() {
	// 取消上下文
	d.cancelFunc()
	// 释放 goroutine 池资源
	d.pool.Release()
	d.log.Info("✔️Task scheduler stopped")
}

func (d *DistributedScheduler) InitializeTasks() {
	d.log.Infof("🤖start initialize tasks from db...")

	taskInfos, err := d.taskInfoDao.ListTaskInfo()
	if err != nil {
		d.log.Errorf("initialize tasks:failed to get task info from db: %v", err)
	}
	for _, taskInfo := range taskInfos {
		task, err := domain.NewTaskFromModel(taskInfo)
		if err != nil {
			d.log.Errorf("initialize tasks:failed to new task from model[%v]: %v", taskInfo, err)
			continue
		}
		err = d.addTask(task, true)
		if err != nil {
			d.log.Errorf("initialize tasks:failed to add task[%v]: %v", taskInfo, err)
			continue
		}
	}
	d.log.Infof("✅initialize tasks from db finished")
}

func (d *DistributedScheduler) generateHashKey(taskId string) string {
	return fmt.Sprintf("%s:%s", RedisTaskKeyPrefix, taskId)
}

func (d *DistributedScheduler) taskToJson(task domain.Task) (string, error) {
	tmp := map[string]string{
		"id":             task.GetID(),
		"task_name":      task.GetTaskName(),
		"scheduled_time": task.GetScheduledTime(),
		"owner_name":     task.GetOwnerName(),
		"description":    task.GetDescription(),
		"job":            task.GetJob().ToJson(),
		"job_type":       task.GetJob().Type(),
	}
	res, err := json.Marshal(tmp)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func (d *DistributedScheduler) jsonToTask(payload string) (domain.Task, error) {
	tmp := make(map[string]string)
	err := json.Unmarshal([]byte(payload), &tmp)
	if err != nil {
		return domain.Task{}, err
	}
	job, err := domain.GetJob(tmp["job_type"])
	if err != nil {
		return domain.Task{}, err
	}
	err = job.UnmarshalFromJson(tmp["job"])
	if err != nil {
		return domain.Task{}, err
	}
	return domain.NewTask(
		tmp["id"],
		tmp["task_name"],
		tmp["scheduled_time"],
		tmp["owner_name"],
		tmp["description"],
		job,
	), nil
}

func (d *DistributedScheduler) parseCronSchedule(c string) (cron.Schedule, error) {
	d.scheduleMappingMu.RLock()
	schedule, exists := d.scheduleMapping[c]
	d.scheduleMappingMu.RUnlock()
	if !exists {
		newSchedule, err := d.parser.Parse(c)
		if err != nil {
			return nil, err
		}
		d.scheduleMappingMu.Lock()
		d.scheduleMapping[c] = newSchedule
		d.scheduleMappingMu.Unlock()
		return newSchedule, nil
	}
	return schedule, nil
}

func (d *DistributedScheduler) now() time.Time {
	return time.Now().In(d.location)
}
