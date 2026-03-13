package main

import (
	"myweather/internal/config"
	"myweather/internal/handler"
	"myweather/internal/server"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

type logrusWriter struct {
	logger *logrus.Logger
}

func (w logrusWriter) Write(p []byte) (n int, err error) {
	w.logger.Info(string(p))
	return len(p), nil
}

func main() {
	fx.New(
		config.Module,
		//weather.Module,
		handler.Module,
		server.Module,
		fx.WithLogger(func(log *logrus.Logger) fxevent.Logger {
			return &fxevent.ConsoleLogger{W: &logrusWriter{logger: log}}
		}),
	).Run()
}
