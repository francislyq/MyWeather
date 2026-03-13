package server

import (
	"context"
	"myweather/internal/config"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func New(lc fx.Lifecycle, r *mux.Router, cfg *config.Config, log *logrus.Logger) *http.Server {
	addr := ":" + strconv.Itoa(cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Infof("Starting server on %s", addr)
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.WithError(err).Fatal("HTTP Server failed to start")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Shutting down HTTP server...")
			return srv.Shutdown(ctx)
		},
	})

	return srv
}
