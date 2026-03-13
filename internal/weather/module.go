package weather

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"weather",
	fx.Provide(
		fx.Annotate(NewOpenWeatherMap, fx.As(new(Provider))),
	),
	fx.Provide(NewService),
)
