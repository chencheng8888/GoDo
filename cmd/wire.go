//go:build wireinject
// +build wireinject

package main

import (
	"github.com/chencheng8888/GoDo/api"
	"github.com/chencheng8888/GoDo/auth"
	"github.com/chencheng8888/GoDo/config"
	"github.com/chencheng8888/GoDo/controller"
	"github.com/chencheng8888/GoDo/dao"
	"github.com/chencheng8888/GoDo/pkg/id_generator"
	"github.com/chencheng8888/GoDo/pkg/log"
	"github.com/chencheng8888/GoDo/scheduler"
	"github.com/google/wire"
)

func WireNewApp(*config.Config) (*App, error) {
	panic(wire.Build(
		NewApp,
		api.ProviderSet,
		auth.ProviderSet,
		config.ProviderSet,
		controller.ProviderSet,
		dao.ProviderSet,
		log.ProviderSet,
		scheduler.ProviderSet,
		id_generator.ProviderSet,
	))
}
