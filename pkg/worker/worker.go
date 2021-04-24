package worker

import (
	"context"
	"github.com/giongto35/cloud-game/v2/pkg/config/worker"
	"github.com/giongto35/cloud-game/v2/pkg/monitoring"
	"github.com/giongto35/cloud-game/v2/pkg/network/httpx"
	"github.com/giongto35/cloud-game/v2/pkg/server"
	"log"
	"net/http"
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

func (wrk *Worker) Run() {
	go func() {
		h := NewHandler(wrk.conf, wrk)
		defer func() {
			log.Printf("[worker] Closing handler")
			h.Close()
		}()

		go h.Run()
		h.Prepare()
		wrk.init()
	}()
	wrk.services.Start()
}

func (wrk *Worker) init() {
	conf := wrk.conf.Worker.Server

	address := conf.Address
	if conf.Https {
		address = conf.Tls.Address
	}
	httpx.NewServer(
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
}

func (wrk *Worker) Shutdown() { wrk.services.Shutdown(wrk.ctx) }
