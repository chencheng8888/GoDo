package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/wire"
	"go.uber.org/zap"
	"net/http"

	"github.com/chencheng8888/GoDo/config"
	"github.com/gin-gonic/gin"
)

var (
	ProviderSet = wire.NewSet(NewAPI, NewGinEngine)
)

type API struct {
	serverConfig *config.ServerConfig
	server       *http.Server

	log *zap.SugaredLogger
}

func NewAPI(sc *config.ServerConfig, e *gin.Engine, log *zap.SugaredLogger) *API {
	addr := fmt.Sprintf("%s:%d", sc.Host, sc.Port)

	srv := &http.Server{
		Addr:    addr,
		Handler: e,
	}

	return &API{
		serverConfig: sc,
		server:       srv,
		log:          log,
	}
}

func (a *API) Run() {
	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic("server run failed:" + err.Error())
	}
}

func (a *API) Close(ctx context.Context) {
	a.log.Info("👉start shutting down server...")
	_ = a.server.Shutdown(ctx)
	a.log.Info("👌server shut down successfully")
}
