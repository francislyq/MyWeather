package config

import (
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"config",
	fx.Provide(
		New,
		NewLogger,
	),
)

func NewLogger() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(logrus.InfoLevel)
	return log
}
