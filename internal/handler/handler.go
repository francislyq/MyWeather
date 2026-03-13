package handler

import (
	"encoding/json"
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
}

func NewHandler(cfg *config.Config, log *logrus.Logger, weatherService *weather.Service) *Handler {
	return &Handler{
		cfg:            cfg,
		log:            log,
		weatherService: weatherService,
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

	// Cache the results (not implemented yet)

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
