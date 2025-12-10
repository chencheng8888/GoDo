package api

import (
	"github.com/chencheng8888/GoDo/auth"
	"github.com/chencheng8888/GoDo/controller"
	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

func NewGinEngine(authService *auth.AuthService, authController *controller.AuthController, taskController *controller.TaskController, logger *zap.SugaredLogger) *gin.Engine {
	r := gin.New()
	r.MaxMultipartMemory = 100 << 20
	r.Use(ginzap.Ginzap(logger.Desugar(), time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger.Desugar(), true))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://task.ccqianchengsijin.icu/"}, // 修改为你的前端域名
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Swagger文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	InitRoutes(r, InitAuthRoute(authController), InitTaskRoute(authService, taskController))
	return r
}

func InitRoutes(r *gin.Engine, initer ...RouteIniter) {
	for _, rt := range initer {
		rt.InitRoute(r)
	}
}

func InitAuthRoute(authController *controller.AuthController) RouteIniter {
	return RouteInitFunc(func(r *gin.Engine) {
		g := r.Group("/api/v1/auth")

		{
			g.POST("/login", authController.Login)
		}
	})
}

func InitTaskRoute(authService *auth.AuthService, taskController *controller.TaskController) RouteIniter {
	return RouteInitFunc(func(r *gin.Engine) {
		g := r.Group("/api/v1/tasks")
		// need auth
		g.Use(auth.AuthMiddleware(authService))
		{
			g.GET("/list", taskController.ListTasks)
			g.POST("/upload_file", taskController.UploadFile)
			g.DELETE("/delete_file", taskController.DeleteFile)
			g.GET("/list_files", taskController.ListFiles)
			g.POST("/add_shell_task", taskController.AddShellTask)
			g.DELETE("/delete", taskController.DeleteTask)
			g.GET("/logs", taskController.ListTaskLog)
			g.POST("/run", taskController.RunTask)
		}
	})
}
