package controller

import (
	"errors"
	"fmt"
	"github.com/chencheng8888/GoDo/auth"
	"github.com/chencheng8888/GoDo/dao"
	"github.com/chencheng8888/GoDo/pkg/id_generator"
	"github.com/chencheng8888/GoDo/scheduler"
	"github.com/chencheng8888/GoDo/scheduler/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/chencheng8888/GoDo/config"
	"github.com/chencheng8888/GoDo/pkg"
	"github.com/chencheng8888/GoDo/pkg/response"
	"github.com/chencheng8888/GoDo/scheduler/job"
	"github.com/gin-gonic/gin"
)

const (
	TaskIDPrefix = "task_"
)

type TaskController struct {
	scheduler scheduler.Scheduler

	generator id_generator.TaskIDGenerator

	userDao *dao.UserDao

	userFileDao *dao.UserFileDao

	workDir string

	log *zap.SugaredLogger

	fileNumberLimit     int
	singleFileSizeLimit int
}

func NewTaskController(s scheduler.Scheduler, generator id_generator.TaskIDGenerator, cf *config.ScheduleConfig, fileConf *config.FileConfig,
	userDao *dao.UserDao, userFileDao *dao.UserFileDao, log *zap.SugaredLogger) (*TaskController, error) {
	err := pkg.CreateDirIfNotExist(cf.WorkDir)
	if err != nil {
		return nil, err
	}

	return &TaskController{
		scheduler:           s,
		workDir:             cf.WorkDir,
		generator:           generator,
		fileNumberLimit:     fileConf.NumberLimit,
		singleFileSizeLimit: fileConf.SingleFileSizeLimit,
		userDao:             userDao,
		userFileDao:         userFileDao,
		log:                 log,
	}, nil
}

// TaskResponse 用于API响应的任务结构体
// @Description 任务信息响应结构
type TaskResponse struct {
	ID            string `json:"id" example:"12345"`                                                 // 任务ID
	TaskName      string `json:"task_name" example:"daily-backup"`                                   // 任务名称
	ScheduledTime string `json:"scheduled_time" example:"0 2 * * * *"`                               // Cron表达式
	OwnerName     string `json:"owner_name" example:"admin"`                                         // 任务拥有者
	Description   string `json:"description" example:"每日数据备份任务"`                                     // 任务描述
	JobType       string `json:"job_type" example:"shell"`                                           // 任务类型
	Job           string `json:"job" example:"{\"command\":\"/bin/bash\",\"args\":[\"backup.sh\"]}"` // 任务详情(JSON格式)
}

// TaskToResponse 将scheduler.Task转换为TaskResponse
func TaskToResponse(task domain.Task) TaskResponse {
	return TaskResponse{
		ID:            task.GetID(),
		TaskName:      task.GetTaskName(),
		ScheduledTime: task.GetScheduledTime(),
		OwnerName:     task.GetOwnerName(),
		Description:   task.GetDescription(),
		JobType:       task.GetJob().Type(),
		Job:           task.GetJob().Content(),
	}
}

// ListTaskResponseData 任务列表响应数据
// @Description 任务列表响应数据结构
type ListTaskResponseData struct {
	Tasks []TaskResponse `json:"tasks"` // 任务列表
}

// ListTasks 获取任务列表
// @Summary 获取用户任务列表
// @Description 根据 JWT token 中的用户名获取该用户的所有任务
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=ListTaskResponseData} "获取成功"
// @Failure 401 {object} response.Response "Unauthorized:
// - your request may be unauthorized
// - Authorization header required
// - Authorization header format must be Bearer <token>
// - Invalid or expired token"
// @Router /api/v1/tasks/list [get]
func (tc *TaskController) ListTasks(c *gin.Context) {

	name, ok := auth.GetUsernameFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(response.InvalidRequestCode, "your request may be unauthorized"))
		return
	}

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

// UploadFile 上传文件
// @Summary 上传文件
// @Description 上传文件到服务器，用于后续任务执行
// @Tags 任务管理
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "文件"
// @Success 200 {object} response.Response{data=UploadScriptResponseData} "上传成功"
// @Failure 400 {object} response.Response "Bad Request: file not uploaded; file too large; file number limit exceeded"
// @Failure 401 {object} response.Response "Unauthorized: Authorization header required; wrong format (must be Bearer <token>); invalid or expired token; your request may be unauthorized"
// @Failure 500 {object} response.Response "Server Error: file save failed; search failed"
// @Router /api/v1/tasks/upload_file [post]
func (tc *TaskController) UploadFile(c *gin.Context) {
	name, ok := auth.GetUsernameFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(response.InvalidRequestCode, "your request may be unauthorized"))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.FileNotUploadedCode, response.FileNotUploadedMsg))
		return
	}

	if file.Size > int64(tc.singleFileSizeLimit)*1024*1024 {
		c.JSON(http.StatusBadRequest, response.Error(response.FileTooLargeCode, response.FileTooLargeMsg))
		return
	}

	cnt, err := tc.userFileDao.CountFiles(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(response.SearchFailedCode, response.SearchFailedMsg))
		return
	}

	if cnt > int64(tc.fileNumberLimit) {
		c.JSON(http.StatusBadRequest, response.Error(response.FileNumberLimitCode, response.FileNumberLimitMsg))
		return
	}

	fileName := fmt.Sprintf("%d-%s", time.Now().UnixMilli(), filepath.Base(file.Filename))
	savePath := filepath.Join(tc.workDir, fileName)

	err = tc.userFileDao.AddUserFileRecord(name, fileName, file.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(response.FileSaveFailedCode, response.FileSaveFailedMsg))
		return
	}

	err = c.SaveUploadedFile(file, savePath, 0755)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(response.FileSaveFailedCode, response.FileSaveFailedMsg))
		return
	}

	c.JSON(http.StatusOK, response.Success(UploadScriptResponseData{FileName: fileName}))
}

type DeleteFileRequest struct {
	FileName string `json:"file_name" example:"1699123456789-script.sh"` // 文件名
}

// DeleteFile 删除对应的文件
// @Summary 删除对应的文件
// @Description 删除对应的文件
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DeleteFileRequest true "文件删除参数"
// @Success 200 {object} response.Response{data=nil} "success"
// @Failure 400 {object} response.Response "Bad Request: invalid request; file not found"
// @Failure 401 {object} response.Response "Unauthorized: your request may be unauthorized; Authorization header required; Authorization header format must be Bearer <token>; Invalid or expired token; your user account may have been deleted"
// @Failure 500 {object} response.Response "Internal Server Error: delete file failed"
// @Router /api/v1/tasks/delete_file [delete]
func (tc *TaskController) DeleteFile(c *gin.Context) {
	name, ok := auth.GetUsernameFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(response.InvalidRequestCode, "your request may be unauthorized"))
		return
	}

	var req DeleteFileRequest

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.InvalidRequestCode, fmt.Sprintf("%s:%s", response.InvalidRequestMsg, err.Error())))
		return
	}

	_, err := tc.userDao.GetUser(name)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusUnauthorized, response.Error(response.InvalidRequestCode, "your user account may have been deleted"))
		return
	}

	err = tc.userFileDao.DeleteUserFileRecord(name, req.FileName)
	if errors.Is(err, dao.UserFileNotFoundErr) {
		c.JSON(http.StatusBadRequest, response.Error(response.FileNotFoundCode, response.FileNotFoundMsg))
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(response.DeleteFileFailedCode, response.DeleteFileFailedMsg))
		return
	}

	fullPath := filepath.Join(tc.workDir, req.FileName)

	err = os.Remove(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			tc.log.Errorf("the file [%v] to be deleted does not exist,user:%v", req.FileName, name)
		} else {
			tc.log.Errorf("failed to delete file [%v],user:%v,error:%v", req.FileName, name, err)
		}

		c.JSON(http.StatusInternalServerError, response.Error(response.DeleteFileFailedCode, response.DeleteFileFailedMsg))
		return
	}
	c.JSON(http.StatusOK, response.Success(nil))
}

type ListFilesResponseData struct {
	Files []string `json:"files"`
}

// ListFiles 查询已有文件
// @Summary 查询已有文件
// @Description 查询已有文件
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=ListFilesResponseData} "success"
// @Failure 401 {object} response.Response "Unauthorized: your request may be unauthorized; Authorization header required; Authorization header must be Bearer <token>; Invalid or expired token"
// @Failure 500 {object} response.Response "Internal Server Error: search failed"
// @Router /api/v1/tasks/list_files [get]
func (tc *TaskController) ListFiles(c *gin.Context) {
	name, ok := auth.GetUsernameFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(response.InvalidRequestCode, "your request may be unauthorized"))
		return
	}

	files, err := tc.userFileDao.ListUserFiles(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(response.SearchFailedCode, response.SearchFailedMsg))
		return
	}

	c.JSON(http.StatusOK, response.Success(ListFilesResponseData{Files: files}))
}

// AddShellTaskRequest 添加Shell任务请求
// @Description 添加Shell任务的请求参数
type AddShellTaskRequest struct {
	TaskName      string   `json:"task_name" binding:"required" example:"daily-backup"`          // 任务名称
	Description   string   `json:"description" binding:"required" example:"每日数据备份任务"`            // 任务描述
	ScheduledTime string   `json:"scheduled_time" binding:"required,cron" example:"0 2 * * * *"` // Cron表达式(支持秒级)
	Command       string   `json:"command" binding:"required" example:"./backup.sh"`             // 执行命令
	Args          []string `json:"args" binding:"omitempty" example:"--full"`                    // 命令参数
	UseShell      bool     `json:"use_shell" binding:"omitempty" example:"true"`                 // 是否使用Shell
	Timeout       int      `json:"timeout" binding:"required,max=7200,gt=0" example:"1800"`      // 超时时间(秒)，最大2小时
}

// AddShellTaskResponseData 添加Shell任务响应数据
// @Description 添加任务成功响应数据
type AddShellTaskResponseData struct {
	TaskId string `json:"task_id" example:"12345"` // 新创建的任务ID
}

// AddShellTask 添加Shell任务
// @Summary 添加Shell任务
// @Description 创建一个新的Shell任务，支持定时执行，任务所有者从JWT token中获取
// @Tags 任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AddShellTaskRequest true "任务创建参数"
// @Success 200 {object} response.Response{data=AddShellTaskResponseData} "success"
// @Failure 400 {object} response.Response "Bad request: invalid request"
// @Failure 401 {object} response.Response "Unauthorized: your request may be unauthorized; Authorization header required; Authorization header must be Bearer <token>; Invalid or expired token"
// @Router /api/v1/tasks/add_shell_task [post]
func (tc *TaskController) AddShellTask(c *gin.Context) {
	name, ok := auth.GetUsernameFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error(response.InvalidRequestCode, "your request may be unauthorized"))
		return
	}

	user, err := tc.userDao.GetUser(name)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Error(response.InvalidRequestCode, "your request may be unauthorized"))
		return
	}

	var req AddShellTaskRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.InvalidRequestCode, fmt.Sprintf("%s:%s", response.InvalidRequestMsg, err.Error())))
		return
	}

	if !user.UseShell && req.UseShell {
		c.JSON(http.StatusBadRequest, response.Error(response.InvalidRequestCode, fmt.Sprintf("%s:%s", response.InvalidRequestMsg, "the user is not allowed to use shell to run commands")))
		return
	}

	shellJob := job.NewShellJob(req.UseShell, time.Duration(req.Timeout)*time.Second, tc.workDir, req.Command, req.Args...)

	taskID := tc.generator.Generate(TaskIDPrefix)

	task := domain.NewTask(taskID, req.TaskName, name, req.ScheduledTime, req.Description, shellJob)
	err = tc.scheduler.AddTask(task)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.InvalidRequestCode, fmt.Sprintf("%s:%s", response.InvalidRequestMsg, err.Error())))
		return
	}
	c.JSON(http.StatusOK, response.Success(AddShellTaskResponseData{TaskId: task.GetID()}))
}

// DeleteTaskRequest 删除任务请求
// @Description 删除任务的请求参数
type DeleteTaskRequest struct {
	UserName string `json:"user_name" binding:"required" example:"admin"` // 任务拥有者
	TaskID   string `json:"task_id"  binding:"required" example:"12345"`  // 任务ID
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
// @Failure 400 {object} response.Response "invalid request"
// @Failure 401 {object} response.Response "Authorization header required / Authorization header format must be Bearer <token> / Invalid or expired token"
// @Failure 500 {object} response.Response "删除任务失败"
// @Router /api/v1/tasks/delete [delete]
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
