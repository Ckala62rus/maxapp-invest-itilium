// Package config loads application settings from YAML and environment variables.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config keeps all runtime settings in one structure.
type Config struct {
	// App contains generic application metadata and runtime mode.
	App AppConfig `yaml:"app"`
	// HTTP contains server bind settings and HTTP timeouts.
	HTTP HTTPConfig `yaml:"http"`
	// Auth contains MAX mini app validation and backend access token settings.
	Auth AuthConfig `yaml:"auth"`
	// Logging contains log formatting settings.
	Logging LoggingConfig `yaml:"logging"`
	// Itilium contains external API connection settings.
	Itilium ItiliumConfig `yaml:"itilium"`
	// Redis contains cache connection settings.
	Redis RedisConfig `yaml:"redis"`
	// Postgres contains relational database settings.
	Postgres PostgresConfig `yaml:"postgres"`
}

// AppConfig stores basic application metadata.
type AppConfig struct {
	// Name is the service name shown in logs and metrics labels.
	Name string `yaml:"name"`
	// Env describes the runtime profile like development or production.
	Env string `yaml:"env"`
	// Version describes the currently running application build.
	Version string `yaml:"version"`
	// DemoMode keeps demo repositories and demo ITILIUM client enabled.
	DemoMode bool `yaml:"demo_mode"`
}

// HTTPConfig stores HTTP server settings.
type HTTPConfig struct {
	// Host is the bind host for the HTTP server.
	Host string `yaml:"host"`
	// Port is the bind port for the HTTP server.
	Port int `yaml:"port"`
	// ReadTimeout limits time spent reading a request.
	ReadTimeout time.Duration `yaml:"read_timeout"`
	// WriteTimeout limits time spent writing a response.
	WriteTimeout time.Duration `yaml:"write_timeout"`
	// IdleTimeout limits idle keepalive duration.
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

// AuthConfig stores MAX validation and backend access token settings.
type AuthConfig struct {
	// BotToken stores the MAX bot token used to validate WebAppData signatures.
	BotToken string `yaml:"bot_token"`
	// AccessTokenSecret stores the HMAC secret for backend bearer tokens.
	AccessTokenSecret string `yaml:"access_token_secret"`
	// AccessTokenTTL limits the lifetime of issued backend access tokens.
	AccessTokenTTL time.Duration `yaml:"access_token_ttl"`
	// MaxInitDataTTL limits how old MAX initData may be before it is rejected.
	MaxInitDataTTL time.Duration `yaml:"max_init_data_ttl"`
	// AllowDebugIdentityHeaders keeps local development possible outside MAX.
	AllowDebugIdentityHeaders bool `yaml:"allow_debug_identity_headers"`
}

// LoggingConfig stores logging settings.
type LoggingConfig struct {
	// Level controls the minimum log severity.
	Level string `yaml:"level"`
	// Format controls plain or json output.
	Format string `yaml:"format"`
}

// ItiliumConfig stores integration settings for the ITILIUM API.
type ItiliumConfig struct {
	// BaseURL is the public or internal base URL of ITILIUM API.
	BaseURL string `yaml:"base_url"`
	// Login is the ITILIUM API login.
	Login string `yaml:"login"`
	// Password is the ITILIUM API password.
	Password string `yaml:"password"`
	// Timeout limits outbound ITILIUM request latency.
	Timeout time.Duration `yaml:"timeout"`
	// InsecureSkipVerify disables TLS certificate verification for temporary development setups.
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
}

// RedisConfig stores cache settings.
type RedisConfig struct {
	// Address stores host:port of the Redis server.
	Address string `yaml:"address"`
	// Password stores the Redis password when auth is enabled.
	Password string `yaml:"password"`
	// DB selects the logical Redis database number.
	DB int `yaml:"db"`
	// Enabled allows the app to disable Redis in local development.
	Enabled bool `yaml:"enabled"`
}

// PostgresConfig stores database settings.
type PostgresConfig struct {
	// Host is the database host.
	Host string `yaml:"host"`
	// Port is the database port.
	Port int `yaml:"port"`
	// Database is the database name.
	Database string `yaml:"database"`
	// User is the database user.
	User string `yaml:"user"`
	// Password is the database password.
	Password string `yaml:"password"`
	// SSLMode controls the TLS mode used by the driver.
	SSLMode string `yaml:"ssl_mode"`
}

// Load loads configuration from YAML and then applies environment overrides.
func Load(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("config path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("decode config file: %w", err)
	}

	applyEnvOverrides(cfg)

	return cfg, nil
}

// Address formats the full HTTP listen address.
func (c HTTPConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// DSN formats a PostgreSQL DSN used by tools and repositories.
func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.SSLMode,
	)
}

// applyEnvOverrides allows container environments to override mounted YAML files.
func applyEnvOverrides(cfg *Config) {
	setStringFromEnv(&cfg.App.Name, "APP_NAME")
	setStringFromEnv(&cfg.App.Env, "APP_ENV")
	setStringFromEnv(&cfg.App.Version, "APP_VERSION")
	setBoolFromEnv(&cfg.App.DemoMode, "APP_DEMO_MODE")

	setStringFromEnv(&cfg.Auth.BotToken, "MAX_BOT_TOKEN")
	setStringFromEnv(&cfg.Auth.AccessTokenSecret, "AUTH_ACCESS_TOKEN_SECRET")
	setDurationFromEnv(&cfg.Auth.AccessTokenTTL, "AUTH_ACCESS_TOKEN_TTL")
	setDurationFromEnv(&cfg.Auth.MaxInitDataTTL, "AUTH_MAX_INIT_DATA_TTL")
	setBoolFromEnv(&cfg.Auth.AllowDebugIdentityHeaders, "AUTH_ALLOW_DEBUG_IDENTITY_HEADERS")

	setStringFromEnv(&cfg.Logging.Level, "LOG_LEVEL")
	setStringFromEnv(&cfg.Logging.Format, "LOG_FORMAT")

	setStringFromEnv(&cfg.Itilium.BaseURL, "ITILIUM_BASE_URL")
	setStringFromEnv(&cfg.Itilium.Login, "ITILIUM_LOGIN")
	setStringFromEnv(&cfg.Itilium.Password, "ITILIUM_PASSWORD")
	setDurationFromEnv(&cfg.Itilium.Timeout, "ITILIUM_TIMEOUT")
	setBoolFromEnv(&cfg.Itilium.InsecureSkipVerify, "ITILIUM_INSECURE_SKIP_VERIFY")

	setStringFromEnv(&cfg.Redis.Address, "REDIS_ADDRESS")
	setStringFromEnv(&cfg.Redis.Password, "REDIS_PASSWORD")
	setIntFromEnv(&cfg.Redis.DB, "REDIS_DB")
	setBoolFromEnv(&cfg.Redis.Enabled, "REDIS_ENABLED")

	setStringFromEnv(&cfg.Postgres.Host, "POSTGRES_HOST")
	setIntFromEnv(&cfg.Postgres.Port, "POSTGRES_PORT")
	setStringFromEnv(&cfg.Postgres.Database, "POSTGRES_DB")
	setStringFromEnv(&cfg.Postgres.User, "POSTGRES_USER")
	setStringFromEnv(&cfg.Postgres.Password, "POSTGRES_PASSWORD")
	setStringFromEnv(&cfg.Postgres.SSLMode, "POSTGRES_SSLMODE")

	setStringFromEnv(&cfg.HTTP.Host, "HTTP_HOST")
	setIntFromEnv(&cfg.HTTP.Port, "HTTP_PORT")
	setDurationFromEnv(&cfg.HTTP.ReadTimeout, "HTTP_READ_TIMEOUT")
	setDurationFromEnv(&cfg.HTTP.WriteTimeout, "HTTP_WRITE_TIMEOUT")
	setDurationFromEnv(&cfg.HTTP.IdleTimeout, "HTTP_IDLE_TIMEOUT")
}

// setStringFromEnv applies a string environment override when it exists.
func setStringFromEnv(target *string, key string) {
	if value, ok := os.LookupEnv(key); ok {
		*target = value
	}
}

// setIntFromEnv applies an integer environment override when it exists.
func setIntFromEnv(target *int, key string) {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.Atoi(value); err == nil {
			*target = parsed
		}
	}
}

// setBoolFromEnv applies a boolean environment override when it exists.
func setBoolFromEnv(target *bool, key string) {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.ParseBool(value); err == nil {
			*target = parsed
		}
	}
}

// setDurationFromEnv applies a duration environment override when it exists.
func setDurationFromEnv(target *time.Duration, key string) {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := time.ParseDuration(value); err == nil {
			*target = parsed
		}
	}
}
