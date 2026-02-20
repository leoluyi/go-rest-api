package config

import (
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/qiangxue/go-env"
	"github.com/leoluyi/go-api-template/pkg/log"
	"gopkg.in/yaml.v2"
)

const (
	defaultServerPort         = 8080
	defaultJWTExpirationHours = 72
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
	// JWT expiration in hours. Defaults to 72 hours (3 days)
	JWTExpiration int `yaml:"jwt_expiration" env:"JWT_EXPIRATION"`
	// authentication username. required. Override with APP_AUTH_USERNAME env var in production.
	AuthUsername string `yaml:"auth_username" env:"AUTH_USERNAME,secret" validate:"required"`
	// authentication password. required. Override with APP_AUTH_PASSWORD env var in production.
	AuthPassword string `yaml:"auth_password" env:"AUTH_PASSWORD,secret" validate:"required"`
	// CORSAllowedOrigins is the list of origins allowed for cross-origin requests.
	// Use ["*"] for development only. In production set specific origins, e.g. ["https://example.com"].
	CORSAllowedOrigins []string `yaml:"cors_allowed_origins"`
}

// Validate validates the application configuration.
func (c Config) Validate() error {
	return validate.Struct(c)
}

// Load returns an application configuration which is populated from the given configuration file and environment variables.
func Load(file string, logger log.Logger) (*Config, error) {
	// default config
	c := Config{
		ServerPort:    defaultServerPort,
		JWTExpiration: defaultJWTExpirationHours,
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
	if err = env.New("APP_", logger.Infof).Load(&c); err != nil {
		return nil, err
	}

	// validation
	if err = c.Validate(); err != nil {
		return nil, err
	}

	return &c, nil
}
