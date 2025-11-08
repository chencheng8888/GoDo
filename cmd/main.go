package main

import (
	"context"
	"flag"
	"github.com/chencheng8888/GoDo/api"
	"github.com/chencheng8888/GoDo/config"
	"github.com/chencheng8888/GoDo/scheduler"
	"os"
	"os/signal"
	"syscall"
	"time"
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
