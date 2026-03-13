package handler

import (
	"encoding/json"
	"myweather/internal/config"
	"myweather/internal/model"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Handler struct {
	// Add any dependencies here, e.g. services, loggers, etc.
	cfg *config.Config
	log *logrus.Logger
}

func NewHandler(cfg *config.Config, log *logrus.Logger) *Handler {
	return &Handler{
		cfg: cfg,
		log: log,
	}
}

func (h *Handler) ListCities(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.cfg.Cities)
}

func (h *Handler) GetCityWeather(w http.ResponseWriter, r *http.Request) {
	// Implement logic to get weather for a specific city
}

func (h *Handler) CollectWeatherData(w http.ResponseWriter, r *http.Request) {
	// Implement logic to collect weather data
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, model.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
	})
}

func (h *Handler) GetCacheStats(w http.ResponseWriter, r *http.Request) {
	// Implement logic to get cache statistics
}

func (h *Handler) ClearCache(w http.ResponseWriter, r *http.Request) {
	// Implement logic to clear cache
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
