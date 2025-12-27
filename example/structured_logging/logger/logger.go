package logger

import (
	"io"
	"log/slog"
	"os"

	"github.com/m0rjc/goconfig"
)

// logWriter holds the opened log writer (file or lumberjack)
var logWriter io.WriteCloser

// Init initializes the global logger based on configuration.
// The configuration may be corrupt if we were unable to load config, so try our best.
func Init() error {
	config, configError := loadConfig()

	// Try to initialize as best we can, even if we have an error.
	// This allows us to log the error before returning.
	loggerError := initialise(config)

	if configError != nil {
		goconfig.LogError(slog.Default(), configError)
		return ErrUnableToLoadConfig
	}
	if loggerError != nil {
		slog.Error("logger_initialisation_failed", "error", loggerError)
		return loggerError
	}
	return nil
}

func initialise(config *LogConfig) error {
	// Create handler options
	opts := &slog.HandlerOptions{
		Level: config.Level,
	}

	// Determine output destination
	var output io.Writer
	if config.FilePath != "" {
		// Use lumberjack for log rotation. Commented out so that this code can build in the goconfig repo.
		//lj := &lumberjack.Logger{
		//	Filename:   config.FilePath,
		//	MaxSize:    config.MaxSize,    // MB
		//	MaxBackups: config.MaxBackups, // Number of backups
		//	MaxAge:     config.MaxAge,     // Days
		//	Compress:   config.Compress,   // Compress old files
		//}
		//logWriter = lj
		//output = lj
	} else {
		// Log to stdout
		output = os.Stdout
	}

	// Choose handler based on format
	var handler slog.Handler
	switch config.Format {
	case LogFormatJSON:
		handler = slog.NewJSONHandler(output, opts)
	case LogFormatText:
		handler = slog.NewTextHandler(output, opts)
	default:
		handler = slog.NewTextHandler(output, opts)
	}

	slog.SetDefault(slog.New(handler))

	// If we opened a file then log the fact
	if config.FilePath != "" {
		slog.Info("logfile_initialised [FAKE - SEE DEMO CODE]",
			"file", config.FilePath,
			"max_size", config.MaxSize,
			"max_backups", config.MaxBackups,
			"max_age", config.MaxAge,
			"compress", config.Compress)
	}
	return nil
}

// Close closes the log writer if one was opened
func Close() {
	if logWriter != nil {
		logWriter.Close()
	}
}
