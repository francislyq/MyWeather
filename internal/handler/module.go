package handler

import (
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"handler",
	fx.Provide(
		NewHandler,
		NewRouter,
	),
)

func NewRouter(h *Handler, log *logrus.Logger) *mux.Router {
	r := mux.NewRouter()

	// Routes
	log.Info("Setting up routes")

	r.HandleFunc("/cities", h.ListCities).Methods("GET")
	r.HandleFunc("/cities/{id}/weather", h.GetCityWeather).Methods("GET")
	r.HandleFunc("/weather/collect", h.CollectWeatherData).Methods("POST")
	r.HandleFunc("/cache/stats", h.GetCacheStats).Methods("GET")
	r.HandleFunc("/cache", h.ClearCache).Methods("DELETE")
	r.HandleFunc("/health", h.HealthCheck).Methods("GET")

	return r
}
