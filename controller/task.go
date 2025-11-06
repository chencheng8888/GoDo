package controller

import (
	"github.com/chencheng8888/GoDo/pkg/response"
	"github.com/chencheng8888/GoDo/scheduler"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TaskController struct {
	scheduler *scheduler.Scheduler
}

func (tc *TaskController) ListTasks(c *gin.Context) {
	name := c.Param("name")
	tasks := tc.scheduler.ListTasks(name)
	c.JSON(http.StatusOK, response.Success(tasks))
}
