package weather

import (
	"context"
	"myweather/internal/cache"
	"myweather/internal/config"
	"myweather/internal/model"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/singleflight"
)

type Service struct {
	provider Provider

	cfg *config.Config
	log *logrus.Logger

	cache   cache.Cache
	sfGroup singleflight.Group
}

func NewService(provider Provider, cfg *config.Config, log *logrus.Logger, cache cache.Cache) *Service {
	return &Service{
		provider: provider,
		cfg:      cfg,
		log:      log,
		cache:    cache,
	}
}

func (s *Service) GetWeather(ctx context.Context, cityID int) (*model.WeatherResult, error) {
	key := strconv.Itoa(cityID)

	// check cache first
	if entry, found := s.cache.Get(key); found {
		w, ok := entry.Value.(*model.Weather)
		if !ok {
			s.log.Warnf("Cache entry for city %d has invalid type, ignoring", cityID)
			s.cache.Delete(key)
		} else {
			s.log.Infof("Cache hit for city %d", cityID)

			return &model.WeatherResult{
				Weather:   w,
				CacheHit:  true,
				FetchedAt: entry.FetchedAt,
				ExpiresAt: entry.ExpiresAt,
			}, nil
		}
	}

	// Singleflight to prevent thundering herd
	result, err, _ := s.sfGroup.Do(key, func() (interface{}, error) {
		s.log.WithField("city_id", cityID).Info("Fetching weather from provider")

		w, err := s.provider.FetchWeather(ctx, cityID)
		if err != nil {
			s.log.Errorf("Failed to fetch weather for city %d: %v", cityID, err)
			return nil, err
		}

		s.cache.Set(key, w, s.cfg.CacheTTL)
		return w, nil
	})

	if err != nil {
		s.log.Errorf("Failed to fetch single flight weather for multiple request from city %d: %v", cityID, err)
		return nil, err
	}

	w := result.(*model.Weather)

	// Read back from cache to get accurate timestamps
	entry, _ := s.cache.Get(key)
	fetchedAt := time.Now()
	expiresAt := time.Now().Add(s.cfg.CacheTTL)
	if entry != nil {
		fetchedAt = entry.FetchedAt
		expiresAt = entry.ExpiresAt
	}

	return &model.WeatherResult{
		Weather:   w,
		CacheHit:  false,
		FetchedAt: fetchedAt,
		ExpiresAt: expiresAt,
	}, nil

	// fetch from provider
	//return s.provider.FetchWeather(ctx, cityID)
}

func (s *Service) CollectWeatherData(ctx context.Context) []model.CityWeatherResult {
	cities := s.cfg.Cities
	results := make([]model.CityWeatherResult, len(cities))

	var wg sync.WaitGroup
	wg.Add(len(cities))

	for i, city := range cities {
		go func(i int, city model.City) {
			defer wg.Done()

			start := time.Now()
			wr, err := s.GetWeather(ctx, city.ID)
			latency := time.Since(start).Milliseconds()
			r := model.CityWeatherResult{
				CityID:    city.ID,
				CityName:  city.Name,
				LatencyMs: latency,
			}

			if err != nil {
				r.Status = "error"
				r.Cache = "MISS"
				r.Error = err.Error()
				s.log.Errorf("Failed to get weather for city %d: %v", city.ID, err)
			} else {
				r.Status = "success"
				if wr.CacheHit {
					r.Cache = "HIT"
				} else {
					r.Cache = "MISS"
				}
				r.Weather = wr.Weather
			}

			s.log.WithFields(logrus.Fields{
				"city_id":    city.ID,
				"city_name":  city.Name,
				"status":     r.Status,
				"cache":      r.Cache,
				"latency_ms": r.LatencyMs,
				"error":      r.Error,
			}).Info("Collected weather data")

			results[i] = r
		}(i, city)
	}

	wg.Wait()
	return results
}
