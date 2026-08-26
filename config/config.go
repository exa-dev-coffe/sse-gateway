package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

type appConfig struct {
	SecretJwt     string
	RabbitmqUrl   string
	AllowedOrigin string
	Port          string
	Env           string
	LogLevel      string
}

var Config appConfig

func init() {
	// Load env
	slog.Info("Loading .env file")
	viper.SetConfigFile(".env") // atau bisa juga pakai viper.SetConfigName("app") + viper.AddConfigPath(".")
	viper.AutomaticEnv()        // override dengan ENV OS kalau ada

	if err := viper.ReadInConfig(); err != nil {
		slog.Warn("No .env file found, fallback to system environment")
	}

	Config = appConfig{
		SecretJwt:     viper.GetString("SECRET_JWT"),
		Port:          viper.GetString("APP_PORT"),
		Env:           viper.GetString("APP_ENV"),
		LogLevel:      viper.GetString("APP_LOG_LEVEL"),
		RabbitmqUrl:   viper.GetString("RABBITMQ_URL"),
		AllowedOrigin: viper.GetString("APP_ALLOW_ORIGINS"),
	}
}
