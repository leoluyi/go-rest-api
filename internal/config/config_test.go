package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validKey is exactly 32 characters, satisfying the JWT signing key minimum.
const validKey = "test-signing-key-exactly-32-char"

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				DSN:           "postgres://localhost/test",
				JWTSigningKey: validKey,
				AuthUsername:  "admin",
				AuthPassword:  "secret",
			},
			wantErr: false,
		},
		{
			name: "missing DSN",
			cfg: Config{
				JWTSigningKey: validKey,
				AuthUsername:  "admin",
				AuthPassword:  "secret",
			},
			wantErr: true,
		},
		{
			name: "JWT signing key too short",
			cfg: Config{
				DSN:           "postgres://localhost/test",
				JWTSigningKey: "tooshort",
				AuthUsername:  "admin",
				AuthPassword:  "secret",
			},
			wantErr: true,
		},
		{
			name: "missing auth username",
			cfg: Config{
				DSN:           "postgres://localhost/test",
				JWTSigningKey: validKey,
				AuthPassword:  "secret",
			},
			wantErr: true,
		},
		{
			name: "missing auth password",
			cfg: Config{
				DSN:           "postgres://localhost/test",
				JWTSigningKey: validKey,
				AuthUsername:  "admin",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestLoad(t *testing.T) {
	t.Run("valid config file with defaults", func(t *testing.T) {
		f := writeTempConfig(t, `
dsn: "postgres://localhost/test"
jwt_signing_key: "`+validKey+`"
auth_username: "admin"
auth_password: "secret"
`)
		cfg, err := Load(f)
		require.NoError(t, err)
		assert.Equal(t, defaultServerPort, cfg.ServerPort)
		assert.Equal(t, defaultJWTExpirationHours, cfg.JWTExpiration)
		assert.Equal(t, "postgres://localhost/test", cfg.DSN)
		assert.Equal(t, "admin", cfg.AuthUsername)
	})

	t.Run("explicit values override defaults", func(t *testing.T) {
		f := writeTempConfig(t, `
dsn: "postgres://localhost/test"
jwt_signing_key: "`+validKey+`"
auth_username: "admin"
auth_password: "secret"
server_port: 9090
jwt_expiration: 24
`)
		cfg, err := Load(f)
		require.NoError(t, err)
		assert.Equal(t, 9090, cfg.ServerPort)
		assert.Equal(t, 24, cfg.JWTExpiration)
	})

	t.Run("env var overrides YAML value", func(t *testing.T) {
		f := writeTempConfig(t, `
dsn: "postgres://localhost/test"
jwt_signing_key: "`+validKey+`"
auth_username: "admin"
auth_password: "secret"
server_port: 8080
`)
		t.Setenv("APP_SERVER_PORT", "7070")
		cfg, err := Load(f)
		require.NoError(t, err)
		assert.Equal(t, 7070, cfg.ServerPort)
	})

	t.Run("missing file returns error", func(t *testing.T) {
		_, err := Load("/nonexistent/path/config.yml")
		assert.Error(t, err)
	})

	t.Run("invalid YAML returns error", func(t *testing.T) {
		f := writeTempConfig(t, `{invalid yaml: [`)
		_, err := Load(f)
		assert.Error(t, err)
	})

	t.Run("validation error on load", func(t *testing.T) {
		f := writeTempConfig(t, `
dsn: "postgres://localhost/test"
jwt_signing_key: "short"
auth_username: "admin"
auth_password: "secret"
`)
		_, err := Load(f)
		assert.Error(t, err)
	})
}

// writeTempConfig writes content to a temp YAML file and returns its path.
// The file is automatically removed at test cleanup.
func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "config-*.yml")
	require.NoError(t, err)
	_, err = f.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}
