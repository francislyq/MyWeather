package weather

import (
	"context"
	"myweather/internal/model"
)

type Provider interface {
	FetchWeather(ctx context.Context, cityID int) (*model.Weather, error)
}
