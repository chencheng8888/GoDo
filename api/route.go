package api

import (
	"github.com/chencheng8888/GoDo/controller"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

type RouteIniter interface {
	InitRoute(r *gin.Engine)
}

type RouteInitFunc func(r *gin.Engine)

func (f RouteInitFunc) InitRoute(r *gin.Engine) {
	f(r)
}

func NewGinEngine(taskController *controller.TaskController, logger *zap.SugaredLogger) *gin.Engine {
	r := gin.New()
	r.Use(ginzap.Ginzap(logger.Desugar(), time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger.Desugar(), true))
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
			g.DELETE("/delete", taskController.DeleteTask)
		}
	})
}
