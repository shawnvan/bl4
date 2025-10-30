package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

// Init initializes the global logger
func Init(level string, format string, output string) error {
	var config zap.Config
	var err error

	switch format {
	case "json":
		config = zap.NewProductionConfig()
	case "console":
		config = zap.NewDevelopmentConfig()
	default:
		config = zap.NewProductionConfig()
	}

	// Parse log level
	logLevel := zap.InfoLevel
	switch level {
	case "debug":
		logLevel = zap.DebugLevel
	case "info":
		logLevel = zap.InfoLevel
	case "warn":
		logLevel = zap.WarnLevel
	case "error":
		logLevel = zap.ErrorLevel
	case "fatal":
		logLevel = zap.FatalLevel
	}

	config.Level = zap.NewAtomicLevelAt(logLevel)
	config.OutputPaths = []string{output}

	// Build logger
	Logger, err = config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return err
	}

	Logger.Info("Logger initialized",
		zap.String("level", level),
		zap.String("format", format),
		zap.String("output", output),
	)

	return nil
}

// Sugar returns a sugared logger for convenience
func Sugar() *zap.SugaredLogger {
	return Logger.Sugar()
}

// Close flushes any buffered log entries
func Close() error {
	if Logger != nil {
		return Logger.Sync()
	}
	return nil
}

// DefaultLogger creates a default logger if none exists
func DefaultLogger() {
	if Logger == nil {
		// Create a basic logger for fallback
		logger, _ := zap.NewProduction()
		Logger = logger
	}
}

// WithLogger executes a function with the logger
func WithLogger(fn func(*zap.Logger)) {
	if Logger != nil {
		fn(Logger)
	} else {
		DefaultLogger()
		fn(Logger)
	}
}