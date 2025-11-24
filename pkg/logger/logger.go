package logger

import (
    "sync"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

var (
    Logger   *zap.Logger
    once     sync.Once
    initErr  error
)

// Init initializes the global logger with thread safety
func Init(level string, format string, output string) error {
    var err error
    once.Do(func() {
        err = initLoggerUnsafe(level, format, output)
        initErr = err
    })
    return initErr
}

// initLoggerUnsafe performs the actual logger initialization
func initLoggerUnsafe(level string, format string, output string) error {
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

// Sugar returns a sugared logger with thread safety
func Sugar() *zap.SugaredLogger {
    if Logger == nil {
        // Initialize with defaults if not already initialized
        _ = Init("info", "console", "stdout")
    }
    return Logger.Sugar()
}

// IsInitialized returns true if logger has been initialized
func IsInitialized() bool {
    return Logger != nil
}

// GetInitError returns any error that occurred during initialization
func GetInitError() error {
    return initErr
}

// Reset resets the logger for testing purposes
func Reset() {
    once = sync.Once{}
    Logger = nil
    initErr = nil
}

// Close flushes any buffered log entries
func Close() error {
    if Logger != nil {
        return Logger.Sync()
    }
    return nil
}

// DefaultLogger creates a default logger if none exists (thread safe)
func DefaultLogger() {
    once.Do(func() {
        if Logger == nil {
            logger, _ := zap.NewProduction()
            Logger = logger
        }
    })
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