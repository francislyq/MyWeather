package config

import (
	"myweather/internal/model"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Port                  int
	APIKey                string
	HTTPTimeout           time.Duration
	Cities                []model.City
	CacheTTL              time.Duration
	CleanupInterval       time.Duration
	StaleWhileRevalidate  time.Duration
}

func New(log *logrus.Logger) *Config {
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnvAsInt("SERVER_PORT", 3000),
		APIKey:      getEnvStr("WEATHER_API_KEY", ""),
		HTTPTimeout: getEnvDuration("HTTP_TIMEOUT", 10*time.Second),
		Cities: []model.City{
			{ID: 6167865, Name: "Toronto", Country: "CA"},
			{ID: 6094817, Name: "Ottawa", Country: "CA"},
			{ID: 1850147, Name: "Tokyo", Country: "JP"},
		},
		CacheTTL:             getEnvDuration("CACHE_TTL", 5*time.Minute),
		CleanupInterval:      getEnvDuration("CACHE_CLEANUP_INTERVAL", 1*time.Minute),
		StaleWhileRevalidate: getEnvDuration("CACHE_STALE_WHILE_REVALIDATE", 0),
	}

	if cfg.APIKey == "" {
		log.Error("API key is not set. Please set the API_KEY environment variable.")
	}

	log.WithFields(logrus.Fields{
		"port":         cfg.Port,
		"http_timeout": cfg.HTTPTimeout,
		"cities_count": len(cfg.Cities),
	}).Info("Configuration loaded")

	return cfg
}

func getEnvAsInt(name string, defaultVal int) int {
	valStr := os.Getenv(name)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func getEnvStr(name string, defaultVal string) string {
	val := os.Getenv(name)
	if val == "" {
		return defaultVal
	}
	return val
}

func getEnvDuration(name string, defaultVal time.Duration) time.Duration {
	valStr := os.Getenv(name)
	if valStr == "" {
		return defaultVal
	}
	val, err := time.ParseDuration(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}
