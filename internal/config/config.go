package config

import (
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/qiangxue/go-env"
	"gopkg.in/yaml.v3"
)

const (
	defaultServerPort               = 8080
	defaultJWTExpirationHours       = 24
	defaultDBMaxOpenConns           = 25
	defaultDBMaxIdleConns           = 5
	defaultDBConnMaxLifetimeMinutes = 5
	defaultShutdownTimeoutSeconds   = 10
)

var validate = validator.New()

// Config represents an application configuration.
type Config struct {
	// the server port. Defaults to 8080
	ServerPort int `yaml:"server_port" env:"SERVER_PORT"`
	// the data source name (DSN) for connecting to the database. required.
	DSN string `yaml:"dsn" env:"DSN,secret" validate:"required"`
	// JWT signing key. required. Must be at least 32 characters for HS256.
	JWTSigningKey string `yaml:"jwt_signing_key" env:"JWT_SIGNING_KEY,secret" validate:"required,min=32"`
	// JWT expiration in hours. Defaults to 24 hours.
	JWTExpiration int `yaml:"jwt_expiration" env:"JWT_EXPIRATION"`
	// authentication username. required. Override with APP_AUTH_USERNAME env var in production.
	AuthUsername string `yaml:"auth_username" env:"AUTH_USERNAME,secret" validate:"required"`
	// authentication password. required. Override with APP_AUTH_PASSWORD env var in production.
	AuthPassword string `yaml:"auth_password" env:"AUTH_PASSWORD,secret" validate:"required"`
	// CORSAllowedOrigins is the list of origins allowed for cross-origin requests.
	// Use ["*"] for development only. In production set specific origins, e.g. ["https://example.com"].
	CORSAllowedOrigins []string `yaml:"cors_allowed_origins"`
	// DBMaxOpenConns sets the maximum number of open database connections. Defaults to 25.
	DBMaxOpenConns int `yaml:"db_max_open_conns" env:"DB_MAX_OPEN_CONNS"`
	// DBMaxIdleConns sets the maximum number of idle database connections. Defaults to 5.
	DBMaxIdleConns int `yaml:"db_max_idle_conns" env:"DB_MAX_IDLE_CONNS"`
	// DBConnMaxLifetimeMinutes sets the maximum lifetime of a database connection in minutes. Defaults to 5.
	DBConnMaxLifetimeMinutes int `yaml:"db_conn_max_lifetime_minutes" env:"DB_CONN_MAX_LIFETIME_MINUTES"`
	// ShutdownTimeoutSeconds is the graceful shutdown timeout in seconds. Defaults to 10.
	ShutdownTimeoutSeconds int `yaml:"shutdown_timeout_seconds" env:"SHUTDOWN_TIMEOUT_SECONDS"`
}

// Validate validates the application configuration.
func (c Config) Validate() error {
	return validate.Struct(c)
}

// Load returns an application configuration which is populated from the given configuration file and environment variables.
func Load(file string) (*Config, error) {
	// default config
	c := Config{
		ServerPort:               defaultServerPort,
		JWTExpiration:            defaultJWTExpirationHours,
		DBMaxOpenConns:           defaultDBMaxOpenConns,
		DBMaxIdleConns:           defaultDBMaxIdleConns,
		DBConnMaxLifetimeMinutes: defaultDBConnMaxLifetimeMinutes,
		ShutdownTimeoutSeconds:   defaultShutdownTimeoutSeconds,
	}

	// load from YAML config file
	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(bytes, &c); err != nil {
		return nil, err
	}

	// load from environment variables prefixed with "APP_"
	// Use a no-op log function to prevent secret values from appearing in logs.
	noop := func(string, ...interface{}) {}
	if err = env.New("APP_", noop).Load(&c); err != nil {
		return nil, err
	}

	// validation
	if err = c.Validate(); err != nil {
		return nil, err
	}

	return &c, nil
}
