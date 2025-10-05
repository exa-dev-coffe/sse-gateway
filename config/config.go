package config

import (
	"log"

	"github.com/spf13/viper"
)

type appConfig struct {
	SecretJwt   string
	RabbitmqUrl string
	Port        string
	Env         string
	LogLevel    string
}

var Config appConfig

func init() {
	// Load env
	log.Println("Loading .env file")
	viper.SetConfigFile(".env") // atau bisa juga pakai viper.SetConfigName("app") + viper.AddConfigPath(".")
	viper.AutomaticEnv()        // override dengan ENV OS kalau ada

	if err := viper.ReadInConfig(); err != nil {
		log.Println("No .env file found, fallback to system environment")
	}

	Config = appConfig{
		SecretJwt:   viper.GetString("SECRET_JWT"),
		Port:        viper.GetString("APP_PORT"),
		Env:         viper.GetString("APP_ENV"),
		LogLevel:    viper.GetString("APP_LOG_LEVEL"),
		RabbitmqUrl: viper.GetString("RABBITMQ_URL"),
	}
}
