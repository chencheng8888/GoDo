// Package main GoDo任务调度系统
// @title GoDo任务调度系统API
// @version 1.0
// @description 这是一个基于Go语言开发的任务调度系统API文档
// @termsOfService http://swagger.io/terms/
//
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
//
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
//
// @host localhost:8080
// @BasePath /
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chencheng8888/GoDo/api"
	"github.com/chencheng8888/GoDo/config"
	"github.com/chencheng8888/GoDo/scheduler"
	_ "github.com/chencheng8888/GoDo/docs" // 导入swagger文档
)

var (
	flagConfig string
)

func init() {
	flag.StringVar(&flagConfig, "conf", "config/config.yaml", "define which configuration file to read")
}

type App struct {
	a *api.API
	s *scheduler.Scheduler
}

func NewApp(a *api.API, s *scheduler.Scheduler) *App {
	return &App{
		a: a,
		s: s,
	}
}

func main() {
	flag.Parse()

	cf := config.LoadConfig(flagConfig)

	app, err := WireNewApp(cf)
	if err != nil {
		panic("wire new app failed: " + err.Error())
	}

	go app.a.Run()
	go app.s.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	const timeout = 5 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	app.a.Close(ctx)
	app.s.Stop()
}
