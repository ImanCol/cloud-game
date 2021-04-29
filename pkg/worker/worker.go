package worker

import (
	"context"
	"net/http"

	"github.com/giongto35/cloud-game/v2/pkg/config/worker"
	"github.com/giongto35/cloud-game/v2/pkg/monitoring"
	"github.com/giongto35/cloud-game/v2/pkg/network/httpx"
	"github.com/giongto35/cloud-game/v2/pkg/server"
)

type Worker struct {
	conf     worker.Config
	ctx      context.Context
	services *server.Services
}

func New(ctx context.Context, conf worker.Config) *Worker {
	services := server.Services{}
	return &Worker{
		ctx:  ctx,
		conf: conf,
		services: services.AddIf(
			conf.Worker.Monitoring.IsEnabled(), monitoring.New(conf.Worker.Monitoring, "worker"),
		),
	}
}

// !to add proper shutdown on app termination

func (wrk *Worker) Run(ctx context.Context) {
	conf := wrk.conf.Worker.Server

	h := NewHandler(wrk.conf, wrk)

	go h.Run(ctx)

	address := conf.Address
	if conf.Https {
		address = conf.Tls.Address
	}

	go httpx.NewServer(
		address,
		func(_ *httpx.Server) http.Handler {
			h := http.NewServeMux()
			h.HandleFunc("/echo", echo)
			return h
		},
		httpx.WithServerConfig(conf),
		// no need just for one route
		httpx.HttpsRedirect(false),
		httpx.WithPortRoll(true),
	).Start()

	wrk.services.Start()
}

func (wrk *Worker) Shutdown() { wrk.services.Shutdown(wrk.ctx) }
