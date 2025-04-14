package logger

import (
	"os"

	"github.com/rs/zerolog"

	"avito-pvz/internal/config"
)

func InitLogger(cfg *config.Config) *zerolog.Logger {
	levelStr := cfg.LogLevel
	format := cfg.LogFormat

	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		level = zerolog.InfoLevel
	}

	var logger zerolog.Logger
	if format == "text" {
		output := zerolog.ConsoleWriter{Out: os.Stdout}
		logger = zerolog.New(output).Level(level).With().Timestamp().Logger()
	} else {
		logger = zerolog.New(os.Stdout).Level(level).With().Timestamp().Logger()
	}

	logger.Info().Msg("Logger initialized")

	return &logger
}
