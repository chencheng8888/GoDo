package controller

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/chencheng8888/GoDo/config"
	"github.com/chencheng8888/GoDo/pkg"
	"github.com/chencheng8888/GoDo/pkg/response"
	"github.com/chencheng8888/GoDo/scheduler"
	"github.com/chencheng8888/GoDo/scheduler/job"
	"github.com/gin-gonic/gin"
)

type TaskController struct {
	scheduler *scheduler.Scheduler

	workDir string
}

func NewTaskController(s *scheduler.Scheduler, cf *config.ScheduleConfig) (*TaskController, error) {
	err := pkg.CreateDirIfNotExist(cf.WorkDir)
	if err != nil {
		return nil, err
	}

	return &TaskController{
		scheduler: s,
		workDir:   cf.WorkDir,
	}, nil
}

// TaskResponse 用于API响应的任务结构体
// @Description 任务信息响应结构
type TaskResponse struct {
	ID            int    `json:"id" example:"12345"`                    // 任务ID
	TaskName      string `json:"task_name" example:"daily-backup"`      // 任务名称
	ScheduledTime string `json:"scheduled_time" example:"0 2 * * * *"`  // Cron表达式
	OwnerName     string `json:"owner_name" example:"admin"`            // 任务拥有者
	Description   string `json:"description" example:"每日数据备份任务"`      // 任务描述
	JobType       string `json:"job_type" example:"shell"`              // 任务类型
	Job           string `json:"job" example:"{\"command\":\"/bin/bash\",\"args\":[\"backup.sh\"]}"`  // 任务详情(JSON格式)
}

// TaskToResponse 将scheduler.Task转换为TaskResponse
func TaskToResponse(task scheduler.Task) TaskResponse {
	return TaskResponse{
		ID:            task.GetID(),
		TaskName:      task.GetTaskName(),
		ScheduledTime: task.GetScheduledTime(),
		OwnerName:     task.GetOwnerName(),
		Description:   task.GetDescription(),
		JobType:       task.GetJobType(),
		Job:           task.GetJobJson(),
	}
}

// ListTaskResponseData 任务列表响应数据
// @Description 任务列表响应数据结构
type ListTaskResponseData struct {
	Tasks []TaskResponse `json:"tasks"` // 任务列表
}

// ListTasks 获取任务列表
// @Summary 获取用户任务列表
// @Description 根据JWT token中的用户名获取该用户的所有任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=ListTaskResponseData} "获取成功"
// @Failure 401 {object} response.Response "未授权"
// @Router /api/v1/tasks/list [get]
func (tc *TaskController) ListTasks(c *gin.Context) {
	// 从JWT中间件设置的上下文中获取用户名
	userName, exists := c.Get("userName")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Error(response.UnauthorizedCode, response.UnauthorizedMsg))
		return
	}

	name := userName.(string)
	tasks := tc.scheduler.ListTasks(name)
	
	// 转换为响应结构体
	var taskResponses []TaskResponse
	for _, task := range tasks {
		taskResponses = append(taskResponses, TaskToResponse(task))
	}
	
	c.JSON(http.StatusOK, response.Success(ListTaskResponseData{Tasks: taskResponses}))
}

// UploadScriptResponseData 上传脚本响应数据
// @Description 脚本上传成功响应数据
type UploadScriptResponseData struct {
	FileName string `json:"file_name" example:"1699123456789-script.sh"` // 上传后的文件名
}

// UploadScript 上传脚本文件
// @Summary 上传脚本文件
// @Description 上传脚本文件到服务器，用于后续任务执行
// @Tags 任务管理
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "脚本文件"
// @Success 200 {object} response.Response{data=UploadScriptResponseData} "上传成功"
// @Failure 400 {object} response.Response "文件未上传"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "文件保存失败"
// @Router /api/v1/tasks/upload_script [post]
func (tc *TaskController) UploadScript(c *gin.Context) {
	// 检查鉴权
	_, exists := c.Get("userName")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Error(response.UnauthorizedCode, response.UnauthorizedMsg))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.FileNotUploadedCode, response.FileNotUploadedMsg))
		return
	}

	fileName := fmt.Sprintf("%d-%s", time.Now().UnixMilli(), filepath.Base(file.Filename))
	savePath := filepath.Join(tc.workDir, fileName)

	err = c.SaveUploadedFile(file, savePath, 0755)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(response.FileSaveFailedCode, response.FileSaveFailedMsg))
		return
	}

	c.JSON(http.StatusOK, response.Success(UploadScriptResponseData{FileName: fileName}))
}

// AddShellTaskRequest 添加Shell任务请求
// @Description 添加Shell任务的请求参数
type AddShellTaskRequest struct {
	TaskName      string   `json:"task_name" binding:"required" example:"daily-backup"`        // 任务名称
	Description   string   `json:"description" binding:"required" example:"每日数据备份任务"`          // 任务描述
	ScheduledTime string   `json:"scheduled_time" binding:"required,cron" example:"0 2 * * * *"` // Cron表达式(支持秒级)
	Command       string   `json:"command" binding:"required" example:"/bin/bash"`             // 执行命令
	Args          []string `json:"args" binding:"omitempty" example:"backup.sh,--full"`        // 命令参数
	UseShell      bool     `json:"use_shell" binding:"required" example:"true"`                // 是否使用Shell
	Timeout       int      `json:"timeout" binding:"required,max=7200,gt=0" example:"1800"`    // 超时时间(秒)，最大2小时
}

// AddShellTaskResponseData 添加Shell任务响应数据
// @Description 添加任务成功响应数据
type AddShellTaskResponseData struct {
	TaskId int `json:"task_id" example:"12345"` // 新创建的任务ID
}

// AddShellTask 添加Shell任务
// @Summary 添加Shell任务
// @Description 创建一个新的Shell任务，支持定时执行，任务所有者从JWT token中获取
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AddShellTaskRequest true "任务创建参数"
// @Success 200 {object} response.Response{data=AddShellTaskResponseData} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Router /api/v1/tasks/add_shell_task [post]
func (tc *TaskController) AddShellTask(c *gin.Context) {
	// 从JWT中间件设置的上下文中获取用户名
	userName, exists := c.Get("userName")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Error(response.UnauthorizedCode, response.UnauthorizedMsg))
		return
	}

	var req AddShellTaskRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.InvalidRequestCode, fmt.Sprintf("%s:%s", response.InvalidRequestMsg, err.Error())))
		return
	}

	ownerName := userName.(string)
	shellJob := job.NewShellJob(req.UseShell, time.Duration(req.Timeout)*time.Second, tc.workDir, req.Command, req.Args...)
	task := scheduler.NewTask(req.TaskName, ownerName, req.ScheduledTime, req.Description, shellJob)
	taskId, err := tc.scheduler.AddTask(task)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.InvalidRequestCode, fmt.Sprintf("%s:%s", response.InvalidRequestMsg, err.Error())))
		return
	}
	c.JSON(http.StatusOK, response.Success(AddShellTaskResponseData{TaskId: taskId}))
}

// DeleteTaskRequest 删除任务请求
// @Description 删除任务的请求参数
type DeleteTaskRequest struct {
	TaskID int `json:"task_id" example:"12345"` // 任务ID
}

// DeleteTask 删除任务
// @Summary 删除任务
// @Description 根据任务ID和JWT token中的用户名删除指定任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DeleteTaskRequest true "删除任务参数"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "删除任务失败"
// @Router /api/v1/tasks/delete [delete]
func (tc *TaskController) DeleteTask(c *gin.Context) {
	// 从JWT中间件设置的上下文中获取用户名
	userName, exists := c.Get("userName")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Error(response.UnauthorizedCode, response.UnauthorizedMsg))
		return
	}

	var req DeleteTaskRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.InvalidRequestCode, fmt.Sprintf("%s:%s", response.InvalidRequestMsg, err.Error())))
		return
	}

	ownerName := userName.(string)
	err := tc.scheduler.RemoveTask(ownerName, req.TaskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(response.DeleteTaskFailedCode, fmt.Sprintf("%s:%s", response.DeleteTaskFailedMsg, err.Error())))
		return
	}
	c.JSON(http.StatusOK, response.Success(nil))
}
