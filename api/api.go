package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/chencheng8888/GoDo/config"
	"github.com/gin-gonic/gin"
)

type API struct {
	serverConfig *config.ServerConfig
	server       *http.Server
}

func NewAPI(sc *config.ServerConfig, e *gin.Engine) *API {
	addr := fmt.Sprintf("%s:%d", sc.Host, sc.Port)

	srv := &http.Server{
		Addr:    addr,
		Handler: e,
	}

	return &API{
		serverConfig: sc,
		server:       srv,
	}
}

func (a *API) Run() {
	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic("server run failed:" + err.Error())
	}
}

func (a *API) Close(ctx context.Context) {
	_ = a.server.Shutdown(ctx)
}
