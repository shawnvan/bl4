package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shawnvan/bl4/internal/api"
	"github.com/shawnvan/bl4/pkg/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Environment represents the application environment
type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
	Testing     Environment = "testing"
)

// String returns the string representation of the environment
func (e Environment) String() string {
	return string(e)
}

// IsProduction returns true if the environment is production
func (e Environment) IsProduction() bool {
	return e == Production
}

// IsDevelopment returns true if the environment is development
func (e Environment) IsDevelopment() bool {
	return e == Development
}

// IsTesting returns true if the environment is testing
func (e Environment) IsTesting() bool {
	return e == Testing
}

// Config holds the complete application configuration
type Config struct {
	Environment Environment          `mapstructure:"environment"`
	Server      *api.Config          `mapstructure:"server"`
	Logging     *LoggingConfig       `mapstructure:"logging"`
	Features    *FeatureFlags        `mapstructure:"features"`
	Performance *PerformanceConfig   `mapstructure:"performance"`
	Storage     *StorageConfig       `mapstructure:"storage"`
	Security    *SecurityConfig      `mapstructure:"security"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// FeatureFlags holds feature toggle configuration
type FeatureFlags struct {
	EnableBitStream      bool `mapstructure:"enable_bitstream"`
	EnableBatchProcess   bool `mapstructure:"enable_batch_process"`
	EnableAnalysis       bool `mapstructure:"enable_analysis"`
	EnableRandomGenerator bool `mapstructure:"enable_random_generator"`
	EnableWebUI          bool `mapstructure:"enable_web_ui"`
	EnableGUI            bool `mapstructure:"enable_gui"`
	EnableCLI            bool `mapstructure:"enable_cli"`
	EnableTUI            bool `mapstructure:"enable_tui"`
	MaxBatchSize         int  `mapstructure:"max_batch_size"`
	EnableCache          bool `mapstructure:"enable_cache"`
	CacheSize            int  `mapstructure:"cache_size"`
}

// PerformanceConfig holds performance-related configuration
type PerformanceConfig struct {
	MaxConcurrentRequests int           `mapstructure:"max_concurrent_requests"`
	RequestTimeout        string        `mapstructure:"request_timeout"`
	ProcessingTimeout     string        `mapstructure:"processing_timeout"`
	EnableMetrics         bool          `mapstructure:"enable_metrics"`
	MetricsPort           int           `mapstructure:"metrics_port"`
	EnableProfiling       bool          `mapstructure:"enable_profiling"`
	ProfilingPort         int           `mapstructure:"profiling_port"`
	MemoryLimitMB         int           `mapstructure:"memory_limit_mb"`
	GCPercent             int           `mapstructure:"gc_percent"`
	MaxIdleConnections    int           `mapstructure:"max_idle_connections"`
	IdleTimeout           string        `mapstructure:"idle_timeout"`
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	Type         string `mapstructure:"type"`
	DataPath     string `mapstructure:"data_path"`
	BackupPath   string `mapstructure:"backup_path"`
	MaxFileSize  int64  `mapstructure:"max_file_size"`
	AutoBackup   bool   `mapstructure:"auto_backup"`
	BackupInterval string `mapstructure:"backup_interval"`
	Compression  bool   `mapstructure:"compression"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	EnableAuth        bool     `mapstructure:"enable_auth"`
	APIKey            string   `mapstructure:"api_key"`
	AllowedOrigins    []string `mapstructure:"allowed_origins"`
	AllowedMethods    []string `mapstructure:"allowed_methods"`
	AllowedHeaders    []string `mapstructure:"allowed_headers"`
	RateLimitEnabled  bool     `mapstructure:"rate_limit_enabled"`
	RateLimitRequests int      `mapstructure:"rate_limit_requests"`
	RateLimitWindow   string   `mapstructure:"rate_limit_window"`
	EnableHTTPS       bool     `mapstructure:"enable_https"`
	CertFile          string   `mapstructure:"cert_file"`
	KeyFile           string   `mapstructure:"key_file"`
}

// Manager manages application configuration
type Manager struct {
	config *Config
	viper  *viper.Viper
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	v := viper.New()

	// Set default configuration
	setDefaults(v)

	// Configure viper
	v.SetEnvPrefix("BL4")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Support for config files
	v.SetConfigName("config")
	v.SetConfigType("json")
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")
	v.AddConfigPath("/etc/bl4")
	v.AddConfigPath("$HOME/.bl4")

	return &Manager{
		viper: v,
	}
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Environment defaults
	v.SetDefault("environment", "development")

	// Server defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.idle_timeout", "120s")
	v.SetDefault("server.max_header_bytes", 1048576)
	v.SetDefault("server.enable_metrics", true)
	v.SetDefault("server.enable_logging", true)
	v.SetDefault("server.enable_cors", true)
	v.SetDefault("server.trust_proxy", false)
	v.SetDefault("server.shutdown_timeout", "30s")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "console")
	v.SetDefault("logging.output", "stdout")
	v.SetDefault("logging.max_size", 100)
	v.SetDefault("logging.max_backups", 3)
	v.SetDefault("logging.max_age", 28)
	v.SetDefault("logging.compress", true)

	// Feature defaults
	v.SetDefault("features.enable_bitstream", false)
	v.SetDefault("features.enable_batch_process", true)
	v.SetDefault("features.enable_analysis", false)
	v.SetDefault("features.enable_random_generator", false)
	v.SetDefault("features.enable_web_ui", true)
	v.SetDefault("features.enable_gui", true)
	v.SetDefault("features.enable_cli", true)
	v.SetDefault("features.enable_tui", true)
	v.SetDefault("features.max_batch_size", 1000)
	v.SetDefault("features.enable_cache", false)
	v.SetDefault("features.cache_size", 100)

	// Performance defaults
	v.SetDefault("performance.max_concurrent_requests", 1000)
	v.SetDefault("performance.request_timeout", "5s")
	v.SetDefault("performance.processing_timeout", "30s")
	v.SetDefault("performance.enable_metrics", true)
	v.SetDefault("performance.metrics_port", 9090)
	v.SetDefault("performance.enable_profiling", false)
	v.SetDefault("performance.profiling_port", 6060)
	v.SetDefault("performance.memory_limit_mb", 512)
	v.SetDefault("performance.gc_percent", 100)
	v.SetDefault("performance.max_idle_connections", 100)
	v.SetDefault("performance.idle_timeout", "90s")

	// Storage defaults
	v.SetDefault("storage.type", "file")
	v.SetDefault("storage.data_path", "./data")
	v.SetDefault("storage.backup_path", "./backups")
	v.SetDefault("storage.max_file_size", 104857600) // 100MB
	v.SetDefault("storage.auto_backup", false)
	v.SetDefault("storage.backup_interval", "24h")
	v.SetDefault("storage.compression", false)

	// Security defaults
	v.SetDefault("security.enable_auth", false)
	v.SetDefault("security.allowed_origins", []string{"*"})
	v.SetDefault("security.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("security.allowed_headers", []string{"Origin", "Content-Type", "Accept", "Authorization"})
	v.SetDefault("security.rate_limit_enabled", false)
	v.SetDefault("security.rate_limit_requests", 100)
	v.SetDefault("security.rate_limit_window", "1m")
	v.SetDefault("security.enable_https", false)
}

// Load loads configuration from file and environment
func (m *Manager) Load() error {
	// Try to read config file
	if err := m.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, use defaults
			logger.Sugar().Info("Config file not found, using defaults and environment variables")
		} else {
			return fmt.Errorf("error reading config file: %w", err)
		}
	} else {
		logger.Sugar().Infow("Config file loaded", "file", m.viper.ConfigFileUsed())
	}

	// Unmarshal configuration
	var config Config
	if err := m.viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := m.validateConfig(&config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	m.config = &config

	// Initialize logging
	if err := m.initializeLogging(); err != nil {
		return fmt.Errorf("failed to initialize logging: %w", err)
	}

	logger.Sugar().Infow("Configuration loaded successfully",
		"environment", config.Environment,
		"server_port", config.Server.Port,
		"log_level", config.Logging.Level,
	)

	return nil
}

// validateConfig validates the loaded configuration
func (m *Manager) validateConfig(config *Config) error {
	// Validate environment
	switch config.Environment {
	case Development, Production, Testing:
		// Valid environments
	default:
		return fmt.Errorf("invalid environment: %s", config.Environment)
	}

	// Validate server configuration
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// Validate logging configuration
	validLogLevels := []string{"debug", "info", "warn", "error", "fatal"}
	validLevel := false
	for _, level := range validLogLevels {
		if config.Logging.Level == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("invalid log level: %s", config.Logging.Level)
	}

	validLogFormats := []string{"json", "console"}
	validFormat := false
	for _, format := range validLogFormats {
		if config.Logging.Format == format {
			validFormat = true
			break
		}
	}
	if !validFormat {
		return fmt.Errorf("invalid log format: %s", config.Logging.Format)
	}

	// Validate feature flags
	if config.Features.MaxBatchSize < 1 || config.Features.MaxBatchSize > 10000 {
		return fmt.Errorf("invalid max batch size: %d", config.Features.MaxBatchSize)
	}

	// Validate performance configuration
	if config.Performance.MaxConcurrentRequests < 1 {
		return fmt.Errorf("invalid max concurrent requests: %d", config.Performance.MaxConcurrentRequests)
	}

	// Validate storage paths
	if config.Storage.DataPath == "" {
		return fmt.Errorf("storage data path cannot be empty")
	}

	// Create storage directories if they don't exist
	if err := m.ensureDirectories(config); err != nil {
		return fmt.Errorf("failed to create storage directories: %w", err)
	}

	return nil
}

// ensureDirectories creates necessary directories
func (m *Manager) ensureDirectories(config *Config) error {
	dirs := []string{
		config.Storage.DataPath,
		config.Storage.BackupPath,
		"./logs",
		"./temp",
	}

	for _, dir := range dirs {
		if dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}
	}

	return nil
}

// initializeLogging initializes the logger based on configuration
func (m *Manager) initializeLogging() error {
	// Resolve output path
	output := m.config.Logging.Output
	if output == "stdout" {
		// Use stdout
	} else if output == "file" {
		// Create log file path
		logDir := "./logs"
		logFile := filepath.Join(logDir, "bl4.log")
		output = logFile
	} else {
		// Use as-is (should be a file path)
	}

	// Initialize logger
	if err := logger.Init(
		m.config.Logging.Level,
		m.config.Logging.Format,
		output,
	); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	return nil
}

// GetConfig returns the loaded configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// GetServerConfig returns the server configuration
func (m *Manager) GetServerConfig() *api.Config {
	if m.config == nil {
		return api.DefaultConfig()
	}
	return m.config.Server
}

// GetEnvironment returns the current environment
func (m *Manager) GetEnvironment() Environment {
	if m.config == nil {
		return Development
	}
	return m.config.Environment
}

// IsProduction returns true if running in production environment
func (m *Manager) IsProduction() bool {
	return m.GetEnvironment().IsProduction()
}

// IsDevelopment returns true if running in development environment
func (m *Manager) IsDevelopment() bool {
	return m.GetEnvironment().IsDevelopment()
}

// IsFeatureEnabled checks if a feature is enabled
func (m *Manager) IsFeatureEnabled(feature string) bool {
	if m.config == nil || m.config.Features == nil {
		return false
	}

	switch feature {
	case "bitstream":
		return m.config.Features.EnableBitStream
	case "batch_process":
		return m.config.Features.EnableBatchProcess
	case "analysis":
		return m.config.Features.EnableAnalysis
	case "random_generator":
		return m.config.Features.EnableRandomGenerator
	case "web_ui":
		return m.config.Features.EnableWebUI
	case "gui":
		return m.config.Features.EnableGUI
	case "cli":
		return m.config.Features.EnableCLI
	case "tui":
		return m.config.Features.EnableTUI
	case "cache":
		return m.config.Features.EnableCache
	default:
		return false
	}
}

// GetConfigPath returns the path to the currently loaded config file
func (m *Manager) GetConfigPath() string {
	return m.viper.ConfigFileUsed()
}

// Reload reloads the configuration from file and environment
func (m *Manager) Reload() error {
	logger.Logger.Info("Reloading configuration")
	return m.Load()
}

// Save saves the current configuration to a file
func (m *Manager) Save(path string) error {
	if m.config == nil {
		return fmt.Errorf("no configuration to save")
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config to file
	if err := m.viper.WriteConfigAs(path); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	logger.Sugar().Infow("Configuration saved", "path", path)
	return nil
}

// Watch enables watching for configuration file changes
func (m *Manager) Watch(callback func(*Config)) error {
	m.viper.WatchConfig()
	m.viper.OnConfigChange(func(e fsnotify.Event) {
		logger.Sugar().Infow("Configuration file changed", "file", e.Name)

		// Reload configuration
		if err := m.Load(); err != nil {
			logger.Logger.Error("Failed to reload configuration", zap.Error(err))
			return
		}

		// Call callback with new configuration
		if callback != nil {
			callback(m.config)
		}
	})

	return nil
}

// Global configuration manager instance
var globalManager *Manager

// InitializeGlobalConfig initializes the global configuration manager
func InitializeGlobalConfig() error {
	globalManager = NewManager()
	return globalManager.Load()
}

// GetGlobalConfig returns the global configuration
func GetGlobalConfig() *Config {
	if globalManager == nil {
		// Initialize with defaults if not already done
		InitializeGlobalConfig()
	}
	return globalManager.GetConfig()
}

// GetGlobalManager returns the global configuration manager
func GetGlobalManager() *Manager {
	if globalManager == nil {
		globalManager = NewManager()
	}
	return globalManager
}