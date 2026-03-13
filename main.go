package main

import (
	"myweather/internal/config"
	"myweather/internal/handler"
	"myweather/internal/server"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

var Modules = []fx.Option{
	fx.WithLogger(func(logger *logrus.Logger) fxevent.Logger {
		return &fxevent.ConsoleLogger{
			W: logger.WithField("component", "fx").Writer(),
		}
	}),

	config.Module,
	//weather.Module,
	handler.Module,
	server.Module,
}

func main() {
	app := fx.New(Modules...)
	app.Run()
	<-app.Done()
}
