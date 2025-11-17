package scheduler

import (
	"context"
	"github.com/chencheng8888/GoDo/dao"
	"github.com/chencheng8888/GoDo/model"
	"go.uber.org/zap"
)

type Middleware func(next Executor) Executor

type LogMiddleware struct {
	log *zap.SugaredLogger
}

func NewLogMiddleware(log *zap.SugaredLogger) *LogMiddleware {
	return &LogMiddleware{log: log}
}

func (l *LogMiddleware) Handler(next Executor) Executor {
	return func(ctx context.Context, t Task) TaskResult {
		l.log.Infof("🚩 start task: %+v", t)
		result := next(ctx, t)
		l.log.Infof("✔️ finish task: %+v,duration: %v, result: %+v", t, result.EndTime.Sub(result.StartTime), result)
		return result
	}
}

type TaskLogMiddleware struct {
	log        *zap.SugaredLogger
	taskLogDao *dao.TaskLogDao
}

func NewTaskLogMiddleware(log *zap.SugaredLogger, taskLogDao *dao.TaskLogDao) *TaskLogMiddleware {
	return &TaskLogMiddleware{log: log, taskLogDao: taskLogDao}
}

func (tl *TaskLogMiddleware) Handler(next Executor) Executor {
	return func(ctx context.Context, t Task) TaskResult {
		result := next(ctx, t)
		taskLog := model.TaskLog{
			TaskId:    t.id,
			Name:      t.taskName,
			Content:   t.f.ToJson(),
			Output:    result.Output,
			ErrOutput: result.ErrOutput,
			StartTime: result.StartTime,
			EndTime:   result.EndTime,
		}
		err := tl.taskLogDao.CreateTaskLog(&taskLog)
		if err != nil {
			tl.log.Errorf("failed to create task log for task %+v: %v", taskLog, err)
		}
		return result
	}
}
