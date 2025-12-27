package logger

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/m0rjc/goconfig"
)

type LogFormat string

var (
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"
)

// LogConfig holds logging configuration
type LogConfig struct {
	Level      slog.Level `key:"LOG_LEVEL" default:"INFO"`
	Format     LogFormat  `key:"LOG_FORMAT" default:"text"`
	FilePath   string     `key:"LOG_FILE"`
	MaxSize    int        `key:"LOG_MAX_SIZE" default:"100" min:"1"`
	MaxBackups int        `key:"LOG_MAX_BACKUPS" default:"0" min:"0"`
	MaxAge     int        `key:"LOG_MAX_AGE" default:"0" min:"0"`
	Compress   bool       `key:"LOG_COMPRESS" default:"false"`
}

var ErrUnableToLoadConfig = fmt.Errorf("unable to load logger configuration")

var typeLogFormat = goconfig.TransformCustomType[string, LogFormat](goconfig.DefaultStringType[string](),
	func(rawValue string) (LogFormat, error) {
		switch strings.ToLower(rawValue) {
		case "text":
			return LogFormatText, nil
		case "json":
			return LogFormatJSON, nil
		default:
			return LogFormatText, fmt.Errorf("invalid log format: %s", rawValue)
		}
	})

var typeLogLevel = goconfig.TransformCustomType[string, slog.Level](goconfig.DefaultStringType[string](),
	func(rawValue string) (slog.Level, error) {
		switch strings.ToUpper(rawValue) {
		case "DEBUG":
			return slog.LevelDebug, nil
		case "INFO":
			return slog.LevelInfo, nil
		case "WARN", "WARNING":
			return slog.LevelWarn, nil
		case "ERROR":
			return slog.LevelError, nil
		default:
			return slog.LevelInfo, fmt.Errorf("invalid log level: %s", rawValue)
		}
	})

func loadConfig() (*LogConfig, error) {
	config := LogConfig{
		Level:  slog.LevelInfo,
		Format: LogFormatText,
	}

	configError := goconfig.Load(context.Background(), &config,
		goconfig.WithCustomType(typeLogFormat),
		goconfig.WithCustomType(typeLogLevel))

	return &config, configError
}
