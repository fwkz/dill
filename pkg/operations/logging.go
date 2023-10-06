package operations

import (
	"log/slog"
	"os"
)

var logLevel = new(slog.LevelVar)

func SetupLogging() {
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})
	slog.SetDefault(slog.New(h))
}
