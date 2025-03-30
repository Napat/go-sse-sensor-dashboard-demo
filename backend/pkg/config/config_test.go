package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Napat/go-sse-sensor-dashboard-demo/backend/pkg/config"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment value and restore it after test completion
	originalEnv := os.Getenv("APP_ENV")
	defer func() {
		os.Setenv("APP_ENV", originalEnv)
	}()

	tests := []struct {
		name          string
		envVars       map[string]string
		expectedPort  int
		expectedError bool
	}{
		{
			name: "valid_config_from_env",
			envVars: map[string]string{
				"APP_ENV":             "dev",
				"APP_PORT":            "9000",
				"APP_STATIC_PATH":     "./testdata",
				"APP_MAX_CONNECTIONS": "5000",
				"APP_READ_TIMEOUT":    "3m",
				"APP_WRITE_TIMEOUT":   "5m",
				"APP_IDLE_TIMEOUT":    "1m",
				"APP_LOG_LEVEL":       "debug",
			},
			expectedPort:  9000,
			expectedError: false,
		},
		{
			name: "missing_required_env",
			envVars: map[string]string{
				"APP_ENV": "dev",
				// APP_STATIC_PATH is required but not set
			},
			expectedError: true,
		},
		{
			name: "invalid_duration_format",
			envVars: map[string]string{
				"APP_ENV":          "dev",
				"APP_STATIC_PATH":  "./testdata",
				"APP_READ_TIMEOUT": "invalid",
			},
			expectedPort:  config.DefaultPort,
			expectedError: false, // No error expected because we use default value
		},
	}

	// Create testdata directory for testing
	err := os.MkdirAll("./testdata", 0755)
	require.NoError(t, err, "Failed to create test directory")
	defer os.RemoveAll("./testdata")

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variables for testing
			for k, v := range tc.envVars {
				os.Setenv(k, v)
			}

			// Call the function being tested
			cfg, err := config.LoadConfig()

			// Check results
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				if cfg != nil {
					assert.Equal(t, tc.expectedPort, cfg.Port)
					assert.Equal(t, os.Getenv("APP_ENV"), string(cfg.Env))
				}
			}

			// Clean up environment after each test
			for k := range tc.envVars {
				os.Unsetenv(k)
			}
		})
	}
}

func TestConfigMethods(t *testing.T) {
	// Test IsProduction, IsUAT, IsDevelopment methods
	tests := []struct {
		name          string
		env           config.Environment
		isProduction  bool
		isUAT         bool
		isDevelopment bool
	}{
		{
			name:          "production_environment",
			env:           config.Prod,
			isProduction:  true,
			isUAT:         false,
			isDevelopment: false,
		},
		{
			name:          "uat_environment",
			env:           config.UAT,
			isProduction:  false,
			isUAT:         true,
			isDevelopment: false,
		},
		{
			name:          "development_environment",
			env:           config.Dev,
			isProduction:  false,
			isUAT:         false,
			isDevelopment: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				Env: tc.env,
			}

			assert.Equal(t, tc.isProduction, cfg.IsProduction())
			assert.Equal(t, tc.isUAT, cfg.IsUAT())
			assert.Equal(t, tc.isDevelopment, cfg.IsDevelopment())
		})
	}
}

func TestString(t *testing.T) {
	// Test String method
	cfg := &config.Config{
		Env:            config.Dev,
		Port:           8080,
		StaticPath:     "/static",
		MaxConnections: 1000,
		ReadTimeout:    5 * time.Minute,
		WriteTimeout:   10 * time.Minute,
		IdleTimeout:    2 * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}

	str := cfg.String()
	assert.Contains(t, str, "Env: dev")
	assert.Contains(t, str, "Port: 8080")
	assert.Contains(t, str, "StaticPath: /static")
	assert.Contains(t, str, "MaxConnections: 1000")
}

func TestConfigDefaults(t *testing.T) {
	// Save original environment value and restore it after test completion
	originalEnv := os.Getenv("APP_ENV")
	defer func() {
		os.Setenv("APP_ENV", originalEnv)
	}()

	// Create testdata directory for testing
	err := os.MkdirAll("./testdata", 0755)
	require.NoError(t, err, "Failed to create test directory")
	defer os.RemoveAll("./testdata")

	// Set only required environment variables
	os.Setenv("APP_ENV", "dev")
	os.Setenv("APP_STATIC_PATH", "./testdata")

	// Call the function being tested
	cfg, err := config.LoadConfig()

	// Check results
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check default values
	assert.Equal(t, config.DefaultPort, cfg.Port)
	assert.Equal(t, config.DefaultMaxConnections, cfg.MaxConnections)
	assert.Equal(t, config.DefaultReadTimeout, cfg.ReadTimeout)
	assert.Equal(t, config.DefaultWriteTimeout, cfg.WriteTimeout)
	assert.Equal(t, config.DefaultIdleTimeout, cfg.IdleTimeout)
	assert.Equal(t, config.DefaultMaxHeaderBytes, cfg.MaxHeaderBytes)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "*", cfg.CORSHosts)
}
