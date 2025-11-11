package controller

import (
	"fmt"
	"github.com/chencheng8888/GoDo/config"
	"github.com/chencheng8888/GoDo/pkg"
	"github.com/chencheng8888/GoDo/pkg/response"
	"github.com/chencheng8888/GoDo/scheduler"
	"github.com/chencheng8888/GoDo/scheduler/job"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"time"
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

type ListTaskResponseData struct {
	Tasks []scheduler.Task `json:"tasks"`
}

func (tc *TaskController) ListTasks(c *gin.Context) {
	name := c.Param("name")
	if len(name) == 0 {
		c.JSON(http.StatusBadRequest, response.Error(response.InvalidRequestCode, response.InvalidRequestMsg))
		return
	}

	tasks := tc.scheduler.ListTasks(name)
	c.JSON(http.StatusOK, response.Success(ListTaskResponseData{Tasks: tasks}))
}

type UploadScriptResponseData struct {
	FileName string `json:"file_name"`
}

func (tc *TaskController) UploadScript(c *gin.Context) {
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

type AddShellTaskRequest struct {
	TaskName      string   `json:"task_name" binding:"required"`
	OwnerName     string   `json:"owner_name" binding:"required"`
	Description   string   `json:"description" binding:"required"`
	ScheduledTime string   `json:"scheduled_time" binding:"required,cron"`
	Command       string   `json:"command" binding:"required"`
	Args          []string `json:"args" binding:"omitempty"`
	UseShell      bool     `json:"use_shell" binding:"required"`
	Timeout       int      `json:"timeout" binding:"required,max=7200,gt=0"` // in seconds, max 2 hours
}

type AddShellTaskResponseData struct {
	TaskId int `json:"task_id"`
}

func (tc *TaskController) AddShellTask(c *gin.Context) {
	var req AddShellTaskRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.InvalidRequestCode, fmt.Sprintf("%s:%s", response.InvalidRequestMsg, err.Error())))
		return
	}

	shellJob := job.NewShellJob(req.UseShell, time.Duration(req.Timeout)*time.Second, tc.workDir, req.Command, req.Args...)
	task := scheduler.NewTask(req.TaskName, req.OwnerName, req.ScheduledTime, req.Description, shellJob)
	taskId, err := tc.scheduler.AddTask(task)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.InvalidRequestCode, fmt.Sprintf("%s:%s", response.InvalidRequestMsg, err.Error())))
		return
	}
	c.JSON(http.StatusOK, response.Success(AddShellTaskResponseData{TaskId: taskId}))
}

type DeleteTaskRequest struct {
	UserName string `json:"user_name"`
	TaskID   int    `json:"task_id"`
}

func (tc *TaskController) DeleteTask(c *gin.Context) {
	var req DeleteTaskRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.InvalidRequestCode, fmt.Sprintf("%s:%s", response.InvalidRequestMsg, err.Error())))
		return
	}

	err := tc.scheduler.RemoveTask(req.UserName, req.TaskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(response.DeleteTaskFailedCode, fmt.Sprintf("%s:%s", response.DeleteTaskFailedMsg, err.Error())))
		return
	}
	c.JSON(http.StatusOK, response.Success(nil))
}
