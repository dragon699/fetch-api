package telemetry

import (
	"log/slog"

	"common/telemetry"
	"connector-slack/internal/config"
)

var Log *slog.Logger

func init() {
	Log = telemetry.NewLogger(
		config.Config.LogLevel,
		config.Config.LogFormat,
	)
}
