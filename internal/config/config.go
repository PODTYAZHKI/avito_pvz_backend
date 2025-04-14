package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppName         string
	AppEnv          string
	AppHost         string
	AppPort         string
	DBHost          string
	DBUser          string
	DBPass          string
	DBName          string
	DBPort          string
	LogLevel        string
	LogFormat       string
	SecretKey       string
	DBSslMode       string
	ShutdownTimeout time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	SetDefaults()

	if err := ValidateRequired(); err != nil {
		return nil, err
	}

	cfg := &Config{
		AppName:         viper.GetString("APP_NAME"),
		AppEnv:          viper.GetString("APP_ENV"),
		AppHost:         viper.GetString("APP_HOST"),
		AppPort:         viper.GetString("APP_PORT"),
		DBHost:          viper.GetString("DB_HOST"),
		DBUser:          viper.GetString("DB_USER"),
		DBPass:          viper.GetString("DB_PASS"),
		DBName:          viper.GetString("DB_NAME"),
		DBPort:          viper.GetString("DB_PORT"),
		DBSslMode:       viper.GetString("DB_SSL_MODE"),
		LogLevel:        viper.GetString("LOG_LEVEL"),
		LogFormat:       viper.GetString("LOG_FORMAT"),
		SecretKey:       viper.GetString("SECRET_KEY"),
		ShutdownTimeout: viper.GetDuration("SHUTDOWN_TIMEOUT"),
		MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
		MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
		ConnMaxLifetime: viper.GetDuration("DB_CONN_MAX_LIFETIME"),
	}

	return cfg, nil
}

func SetDefaults() {
	viper.SetDefault("APP_NAME", "pvz-service")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_HOST", "0.0.0.0	")
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("DB_HOST", "postgres")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASS", "secret")
	viper.SetDefault("DB_NAME", "pvz")
	viper.SetDefault("DB_SSL_MODE", "disable")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "json")
	viper.SetDefault("SHUTDOWN_TIMEOUT", "10s")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 25)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 25)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", "5m")
	viper.SetDefault("SECRET_KEY", "very-secret-key")
}

func ValidateRequired() error {
	required := []string{
		"APP_NAME",
		"APP_HOST",
		"APP_PORT",
		"DB_HOST",
		"DB_USER",
		"DB_PASS",
		"DB_NAME",
		"SECRET_KEY",
	}

	for _, key := range required {
		if viper.GetString(key) == "" {
			return fmt.Errorf("required configuration %s is missing", key)
		}
	}

	return nil
}
