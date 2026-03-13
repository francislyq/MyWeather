package weather

import (
	"context"
	"myweather/internal/config"
	"myweather/internal/model"

	"github.com/sirupsen/logrus"
)

type Service struct {
	provider Provider

	cfg *config.Config
	log *logrus.Logger
}

func NewService(provider Provider, cfg *config.Config, log *logrus.Logger) *Service {
	return &Service{
		provider: provider,
		cfg:      cfg,
		log:      log,
	}
}

func (s *Service) GetWeather(ctx context.Context, cityID int) (*model.Weather, error) {
	// check cache first (not implemented yet)

	// fetch from provider
	return s.provider.FetchWeather(ctx, cityID)
}
