package api

import (
	"time"

	"github.com/chencheng8888/GoDo/auth"
	"github.com/chencheng8888/GoDo/controller"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

type RouteIniter interface {
	InitRoute(r *gin.Engine)
}

type RouteInitFunc func(r *gin.Engine)

func (f RouteInitFunc) InitRoute(r *gin.Engine) {
	f(r)
}

func NewGinEngine(taskController *controller.TaskController, userController *controller.UserController, auth *auth.Auth, logger *zap.SugaredLogger) *gin.Engine {
	r := gin.New()
	r.Use(ginzap.Ginzap(logger.Desugar(), time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger.Desugar(), true))
	
	// Swagger文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	
	InitRoutes(r, InitTaskRoute(taskController, auth), InitUserRoute(userController))
	return r
}

func InitRoutes(r *gin.Engine, initer ...RouteIniter) {
	for _, rt := range initer {
		rt.InitRoute(r)
	}
}

func InitTaskRoute(taskController *controller.TaskController, auth *auth.Auth) RouteIniter {
	return RouteInitFunc(func(r *gin.Engine) {
		g := r.Group("/api/v1/tasks")
		// 为所有任务路由添加JWT鉴权中间件
		g.Use(auth.JWTAuthMiddleware())
		{
			g.GET("/list", taskController.ListTasks)
			g.POST("/upload_script", taskController.UploadScript)
			g.POST("/add_shell_task", taskController.AddShellTask)
			g.DELETE("/delete", taskController.DeleteTask)
		}
	})
}

func InitUserRoute(userController *controller.UserController) RouteIniter {
	return RouteInitFunc(func(r *gin.Engine) {
		g := r.Group("/api/v1/auth")
		{
			g.POST("/login", userController.Login)
			g.POST("/register", userController.Register)
		}
	})
}
