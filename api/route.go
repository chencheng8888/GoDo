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

func NewGinEngine(taskController *controller.TaskController) *gin.Engine {
	r := gin.Default()
	InitRoutes(r, InitTaskRoute(taskController))
	return r
}

func InitRoutes(r *gin.Engine, initer ...RouteIniter) {
	for _, rt := range initer {
		rt.InitRoute(r)
	}
}

func InitTaskRoute(taskController *controller.TaskController) RouteIniter {
	return RouteInitFunc(func(r *gin.Engine) {
		g := r.Group("/api/v1/tasks")
		{
			g.GET("/list/:name", taskController.ListTasks)
			g.POST("/upload_script", taskController.UploadScript)
			g.POST("/add_shell_task", taskController.AddShellTask)
		}
	})
}
