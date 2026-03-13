package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"myweather/internal/config"
	"myweather/internal/model"
	"net/http"

	"github.com/sirupsen/logrus"
)

const baseURL = "https://api.openweathermap.org/data/2.5/weather"

type owmResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`

	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`

	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`

	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Humidity  int     `json:"humidity"`
		Pressure  int     `json:"pressure"`
		SeaLevel  int     `json:"sea_level"`
		GrndLevel int     `json:"grnd_level"`
	} `json:"main"`

	Visibility int `json:"visibility"`

	Wind struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
		Gust  float64 `json:"gust"`
	} `json:"wind"`

	Rain struct {
		OneH   float64 `json:"1h"`
		ThreeH float64 `json:"3h"`
	} `json:"rain"`

	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`

	Dt  int64 `json:"dt"`
	Sys struct {
		ID      int    `json:"id"`
		Country string `json:"country"`
		Sunrise int64  `json:"sunrise"`
		Sunset  int64  `json:"sunset"`
	} `json:"sys"`

	TimeZone int `json:"timezone"`
	Cod      int `json:"cod"`
}

type OpenWeatherMap struct {
	cfg    *config.Config
	log    *logrus.Logger
	client *http.Client
}

func NewOpenWeatherMap(cfg *config.Config, log *logrus.Logger) Provider {
	return &OpenWeatherMap{
		cfg: cfg,
		log: log,
		client: &http.Client{
			Timeout: cfg.HTTPTimeout,
		},
	}
}

func (o *OpenWeatherMap) FetchWeather(ctx context.Context, cityID int) (*model.Weather, error) {
	url := baseURL + "?id=" + fmt.Sprintf("%d", cityID) + "&appid=" + o.cfg.APIKey + "&units=metric"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		o.log.Errorf("Failed to create request: %v", err)
		return nil, err
	}

	resp, err := o.client.Do(req)
	if err != nil {
		o.log.Errorf("Failed to fetch weather: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		o.log.Errorf("Unexpected status code: %v", resp.Status)
		return nil, fmt.Errorf("Unexpected status code: %v", resp.Status)
	}

	var weatherResp owmResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		o.log.Errorf("Error in decode response: %v", err)
		return nil, err
	}

	weather := &model.Weather{
		CityID:   weatherResp.ID,
		CityName: weatherResp.Name,

		Temperature: weatherResp.Main.Temp,
		FeelsLike:   weatherResp.Main.FeelsLike,
		TempMin:     weatherResp.Main.TempMin,
		TempHigh:    weatherResp.Main.TempMax,
		Pressure:    weatherResp.Main.Pressure,
		Humidity:    weatherResp.Main.Humidity,
		SeaLevel:    weatherResp.Main.SeaLevel,
		GrndLevel:   weatherResp.Main.GrndLevel,

		Visibility: weatherResp.Visibility,
		WindSpeed:  weatherResp.Wind.Speed,
		WindDeg:    weatherResp.Wind.Deg,
		WindGust:   weatherResp.Wind.Gust,

		Rain1h:    weatherResp.Rain.OneH,
		CloudsAll: weatherResp.Clouds.All,
	}

	if len(weatherResp.Weather) > 0 {
		weather.WeatherMain = weatherResp.Weather[0].Main
		weather.Description = weatherResp.Weather[0].Description
		weather.Icon = weatherResp.Weather[0].Icon
	}

	return weather, nil
}
