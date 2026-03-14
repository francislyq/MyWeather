package cache

import (
	"context"
	"myweather/internal/config"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

var Module = fx.Module("cache",
	fx.Provide(func() Cache {
		return NewMemoryCache()
	}),
	fx.Invoke(registerCleanUp),
)

func registerCleanUp(lc fx.Lifecycle, cache Cache, cfg *config.Config, log *logrus.Logger) {
	stopChan := make(chan struct{})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.WithField("Interval", cfg.CleanupInterval).Info("Starting cache cleanup goroutine")

			go cache.StartCleanup(cfg.CleanupInterval, stopChan)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping cache cleanup goroutine")
			close(stopChan)
			return nil
		},
	})
}
