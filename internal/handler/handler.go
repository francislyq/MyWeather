package handler

import (
	"encoding/json"
	"fmt"
	"myweather/internal/cache"
	"myweather/internal/config"
	"myweather/internal/model"
	"myweather/internal/weather"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	cfg *config.Config
	log *logrus.Logger

	weatherService *weather.Service
	cache          cache.Cache
}

func NewHandler(cfg *config.Config, log *logrus.Logger, weatherService *weather.Service, c cache.Cache) *Handler {
	return &Handler{
		cfg:            cfg,
		log:            log,
		weatherService: weatherService,
		cache:          c,
	}
}

func (h *Handler) ListCities(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.cfg.Cities)
}

func (h *Handler) GetCityWeather(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cityIDStr, ok := vars["id"]
	if !ok {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{
			Error: "id is required",
			Code:  http.StatusBadRequest,
		})
		return
	}

	cityID, err := strconv.Atoi(cityIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{
			Error: "Invalid city_id",
			Code:  http.StatusBadRequest,
		})
		return
	}

	// Check if city exists in config
	cityExists := false
	for _, city := range h.cfg.Cities {
		if city.ID == cityID {
			cityExists = true
			break
		}
	}

	if !cityExists {
		writeJSON(w, http.StatusNotFound, model.ErrorResponse{
			Error: "City not found in the configuration cities list",
			Code:  http.StatusNotFound,
		})
		return
	}

	results, err := h.weatherService.GetWeather(r.Context(), cityID)
	if err != nil {
		h.log.Errorf("Failed to get weather for city %d: %v", cityID, err)
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{
			Error: "Failed to get weather data",
			Code:  http.StatusInternalServerError,
		})
		return
	}

	// Cache the results
	setCacheHeaders(w, results)

	writeJSON(w, http.StatusOK, results)
}

func (h *Handler) CollectWeatherData(w http.ResponseWriter, r *http.Request) {
	// Implement logic to collect weather data
	results := h.weatherService.CollectWeatherData(r.Context())
	writeJSON(w, http.StatusOK, results)
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, model.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
	})
}

func (h *Handler) GetCacheStats(w http.ResponseWriter, r *http.Request) {
	stats := h.cache.Stats()
	writeJSON(w, http.StatusOK, model.CacheStatsResponse{
		Hits:   int(stats.Hits),
		Misses: int(stats.Misses),
		Size:   int(stats.Size),
	})
}

func (h *Handler) ClearCache(w http.ResponseWriter, r *http.Request) {
	h.cache.Clear()
	h.log.Info("Cache flushed")
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

const headerXCache = "X-Cache"

func setCacheHeaders(w http.ResponseWriter, result *model.WeatherResult) {
	switch {
	case result.CacheHit && result.IsStale:
		w.Header().Set(headerXCache, "STALE")
	case result.CacheHit:
		w.Header().Set(headerXCache, "HIT")
	default:
		w.Header().Set(headerXCache, "MISS")
	}

	ttlRemaining := time.Until(result.ExpiresAt).Seconds()
	if ttlRemaining < 0 {
		ttlRemaining = 0
	}
	w.Header().Set("X-Cache-TTL-Remaining", fmt.Sprintf("%.0f", ttlRemaining))
	w.Header().Set("X-Cache-Fetched-At", result.FetchedAt.UTC().Format(time.RFC3339))
}
