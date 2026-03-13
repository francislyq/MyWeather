package weather

import (
	"context"
	"myweather/internal/config"
	"myweather/internal/model"
	"sync"

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

func (s *Service) CollectWeatherData(ctx context.Context) []*model.Weather {
	cities := s.cfg.Cities
	results := make([]*model.Weather, len(cities))

	var wg sync.WaitGroup
	wg.Add(len(cities))

	for i, city := range cities {
		go func(i int, city model.City) {
			defer wg.Done()
			weather, err := s.GetWeather(ctx, city.ID)
			if err != nil {
				s.log.Errorf("Failed to get weather for city %d: %v", city.ID, err)
				return
			}
			results[i] = weather
		}(i, city)
	}

	wg.Wait()
	return results
}
