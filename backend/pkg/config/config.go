package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"

	"github.com/Napat/go-sse-sensor-dashboard-demo/backend/pkg/apierror"
)

const (
	DefaultPort           = 8080
	DefaultMaxConnections = 10000
	DefaultReadTimeout    = 5 * time.Minute
	DefaultWriteTimeout   = 10 * time.Minute
	DefaultIdleTimeout    = 2 * time.Minute
	DefaultMaxHeaderBytes = 1 << 20 // 1MB
)

type Environment string

const (
	Dev  Environment = "dev"
	UAT  Environment = "uat"
	Prod Environment = "prod"
)

type SecurityConfig struct {
	// XSS Protection header
	XSSProtection string `mapstructure:"APP_XSS_PROTECTION" validate:"required"`

	// Content-Type-Nosniff header
	ContentTypeNosniff string `mapstructure:"APP_CONTENT_TYPE_NOSNIFF" validate:"required"`

	// X-Frame-Options header
	XFrameOptions string `mapstructure:"APP_X_FRAME_OPTIONS" validate:"required,oneof=DENY SAMEORIGIN ALLOW-FROM"`

	// HSTS max age
	HSTSMaxAge int `mapstructure:"APP_HSTS_MAX_AGE" validate:"required,min=0"`

	// Content Security Policy
	CSPPolicy string `mapstructure:"APP_CSP_POLICY" validate:"required"`
}

type Config struct {
	Port           int         `mapstructure:"APP_PORT" validate:"required,min=1024,max=65535"`
	StaticPath     string      `mapstructure:"APP_STATIC_PATH" validate:"required,direxists"`
	Env            Environment `mapstructure:"APP_ENV" validate:"required,oneof=dev uat prod"`
	MaxConnections int         `mapstructure:"APP_MAX_CONNECTIONS" validate:"required,min=10,max=100000"`

	ReadTimeout    time.Duration `mapstructure:"APP_READ_TIMEOUT" validate:"required,min=1s"`
	WriteTimeout   time.Duration `mapstructure:"APP_WRITE_TIMEOUT" validate:"required,min=1s"`
	IdleTimeout    time.Duration `mapstructure:"APP_IDLE_TIMEOUT" validate:"required,min=1s"`
	MaxHeaderBytes int           `mapstructure:"APP_MAX_HEADER_BYTES" validate:"required,min=1024"`

	LogLevel  string         `mapstructure:"APP_LOG_LEVEL"`
	CORSHosts string         `mapstructure:"APP_CORS_HOSTS"`
	Security  SecurityConfig `validate:"required"`
}

func loadEnvFile(env string) error {
	var envType Environment
	switch env {
	case string(Dev):
		envType = Dev
	case string(UAT):
		envType = UAT
	case string(Prod):
		envType = Prod
	default:
		envType = Dev
	}

	configPath := GetEnvFilePath(envType)

	possibleLocations := []string{
		"../" + configPath,
		"/app/configs/backend/.env." + env,
		fmt.Sprintf("configs/.env.%s", env),
		fmt.Sprintf("./configs/.env.%s", env),
		fmt.Sprintf("../configs/.env.%s", env),
		fmt.Sprintf("/app/configs/.env.%s", env),
	}

	v := viper.New()

	v.AutomaticEnv()

	var foundConfig bool
	var loadErrors []string

	for _, location := range possibleLocations {
		absPath, _ := filepath.Abs(location)
		if _, err := os.Stat(location); err == nil {
			v.SetConfigFile(location)
			v.SetConfigType("env")

			if err := v.ReadInConfig(); err != nil {
				return apierror.Wrap(apierror.ErrConfigNotFound, fmt.Sprintf("error loading file %s: %v", location, err))
			}

			fmt.Printf("Loaded configuration from %s (absolute: %s)\n", location, absPath)
			foundConfig = true
			break
		}
		loadErrors = append(loadErrors, fmt.Sprintf("%s (absolute: %s)", location, absPath))
	}

	if !foundConfig {
		return apierror.Wrap(apierror.ErrConfigNotFound,
			fmt.Sprintf("failed to load .env.%s file, tried: %s", env, strings.Join(loadErrors, ", ")))
	}

	v.SetDefault("APP_PORT", DefaultPort)
	v.SetDefault("APP_MAX_CONNECTIONS", DefaultMaxConnections)
	v.SetDefault("APP_READ_TIMEOUT", DefaultReadTimeout.String())
	v.SetDefault("APP_WRITE_TIMEOUT", DefaultWriteTimeout.String())
	v.SetDefault("APP_IDLE_TIMEOUT", DefaultIdleTimeout.String())
	v.SetDefault("APP_MAX_HEADER_BYTES", DefaultMaxHeaderBytes)
	v.SetDefault("APP_LOG_LEVEL", "info")
	v.SetDefault("APP_CORS_HOSTS", "*")

	viper.MergeConfigMap(v.AllSettings())

	return nil
}

func processConfigValue(value string) string {
	value = strings.TrimSpace(value)

	if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
		(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		// Remove quotation marks
		return value[1 : len(value)-1]
	}

	return value
}

func LoadConfig(envOpt ...string) (*Config, error) {
	var env string
	if len(envOpt) > 0 && envOpt[0] != "" {
		env = envOpt[0]
	} else {
		env = os.Getenv("APP_ENV")
		if env == "" {
			return nil, apierror.Wrap(apierror.ErrEnvironmentInvalid, "APP_ENV environment variable not set")
		}
	}

	switch env {
	case string(Dev), string(UAT), string(Prod):
	default:
		return nil, apierror.Wrap(apierror.ErrEnvironmentInvalid,
			fmt.Sprintf("invalid value '%s', must be one of: dev, uat, prod", env))
	}

	if err := loadEnvFile(env); err != nil {
		return nil, err
	}

	viper.SetDefault("APP_PORT", DefaultPort)
	viper.SetDefault("APP_MAX_CONNECTIONS", DefaultMaxConnections)
	viper.SetDefault("APP_READ_TIMEOUT", DefaultReadTimeout.String())
	viper.SetDefault("APP_WRITE_TIMEOUT", DefaultWriteTimeout.String())
	viper.SetDefault("APP_IDLE_TIMEOUT", DefaultIdleTimeout.String())
	viper.SetDefault("APP_MAX_HEADER_BYTES", DefaultMaxHeaderBytes)
	viper.SetDefault("APP_LOG_LEVEL", "info")
	viper.SetDefault("APP_CORS_HOSTS", "*")
	viper.SetDefault("APP_ENV", env)

	var config Config

	if err := viper.Unmarshal(&config); err != nil {
		return nil, apierror.Wrap(apierror.ErrInvalidConfig, fmt.Sprintf("unable to decode into config struct: %v", err))
	}

	config.Env = Environment(env)
	if config.ReadTimeout == 0 {
		readTimeoutStr := viper.GetString("APP_READ_TIMEOUT")
		duration, err := time.ParseDuration(readTimeoutStr)
		if err != nil {
			config.ReadTimeout = DefaultReadTimeout
		} else {
			config.ReadTimeout = duration
		}
	}

	if config.WriteTimeout == 0 {
		writeTimeoutStr := viper.GetString("APP_WRITE_TIMEOUT")
		duration, err := time.ParseDuration(writeTimeoutStr)
		if err != nil {
			config.WriteTimeout = DefaultWriteTimeout
		} else {
			config.WriteTimeout = duration
		}
	}

	if config.IdleTimeout == 0 {
		idleTimeoutStr := viper.GetString("APP_IDLE_TIMEOUT")
		duration, err := time.ParseDuration(idleTimeoutStr)
		if err != nil {
			config.IdleTimeout = DefaultIdleTimeout
		} else {
			config.IdleTimeout = duration
		}
	}

	config.Security = SecurityConfig{
		XSSProtection:      processConfigValue(viper.GetString("APP_XSS_PROTECTION")),
		ContentTypeNosniff: processConfigValue(viper.GetString("APP_CONTENT_TYPE_NOSNIFF")),
		XFrameOptions:      processConfigValue(viper.GetString("APP_X_FRAME_OPTIONS")),
		HSTSMaxAge:         viper.GetInt("APP_HSTS_MAX_AGE"),
		CSPPolicy:          processConfigValue(viper.GetString("APP_CSP_POLICY")),
	}

	if config.StaticPath == "" {
		return nil, apierror.Wrap(apierror.ErrInvalidConfig, "APP_STATIC_PATH required but not set")
	}

	validate := validator.New()

	validate.RegisterValidation("direxists", func(fl validator.FieldLevel) bool {
		if strings.HasPrefix(fl.Field().String(), "/app/") {
			return true
		}

		_, err := os.Stat(fl.Field().String())
		return err == nil
	})

	if err := validate.Struct(config); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				return nil, apierror.Wrap(apierror.ErrInvalidConfig,
					fmt.Sprintf("validation failed for field %s on '%s' tag", e.Field(), e.Tag()))
			}
		}
		return nil, apierror.Wrap(apierror.ErrInvalidConfig, err.Error())
	}

	return &config, nil
}

func (c *Config) IsProduction() bool {
	return c.Env == Prod
}

func (c *Config) IsUAT() bool {
	return c.Env == UAT
}

func (c *Config) IsDevelopment() bool {
	return c.Env == Dev
}

func (c *Config) String() string {
	return fmt.Sprintf("Config{Port: %d, Env: %s, StaticPath: %s, MaxConnections: %d, ...}",
		c.Port, c.Env, c.StaticPath, c.MaxConnections)
}

func GetEnvFilePath(env Environment) string {
	basePath := "./configs/backend"
	switch env {
	case Dev:
		return basePath + "/.env.dev"
	case UAT:
		return basePath + "/.env.uat"
	case Prod:
		return basePath + "/.env.prod"
	default:
		return basePath + "/.env.dev"
	}
}
