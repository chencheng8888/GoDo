package api

import (
	"github.com/chencheng8888/GoDo/task"
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

func InitTaskRoute(r *gin.Engine, scheduler *task.Scheduler) {
	g := r.Group("/api/v1/tasks")
	{
		g.GET("/list/:name")
	}
}
