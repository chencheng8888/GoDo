package controller

import (
	"github.com/chencheng8888/GoDo/pkg/response"
	"github.com/chencheng8888/GoDo/task"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TaskController struct {
	scheduler *task.Scheduler
}

func (tc *TaskController) ListTasks(c *gin.Context) {
	name := c.Param("name")
	tasks := tc.scheduler.ListTasks(name)
	c.JSON(http.StatusOK, response.Success(tasks))
}
