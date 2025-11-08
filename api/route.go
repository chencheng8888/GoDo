package api

import (
	"github.com/chencheng8888/GoDo/controller"
	"github.com/gin-gonic/gin"
)

type RouteIniter interface {
	InitRoute(r *gin.Engine)
}

type RouteInitFunc func(r *gin.Engine)

func (f RouteInitFunc) InitRoute(r *gin.Engine) {
	f(r)
}

func InitRoutes(r *gin.Engine, initer ...RouteIniter) {
	for _, rt := range initer {
		rt.InitRoute(r)
	}
}

func InitTaskRoute(r *gin.Engine, taskController *controller.TaskController) {
	g := r.Group("/api/v1/tasks")
	{
		g.GET("/list/:name", taskController.ListTasks)
		g.POST("/upload_script", taskController.UploadScript)
		g.POST("/add_shell_task", taskController.AddShellTask)
	}
}
