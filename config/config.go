package config

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	Server   Server
	Database Database
}

type Server struct {
	Port string
}
type Database struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func NewConfig() (*Config, error) {

	var config Config

	config.Server.Port = viper.GetString("SERVER_PORT")
	config.Database.Host = viper.GetString("DATABASE_HOST")
	config.Database.Port = viper.GetString("DATABASE_PORT")
	config.Database.User = viper.GetString("DATABASE_USER")
	config.Database.Password = viper.GetString("DATABASE_PASSWORD")
	config.Database.Name = viper.GetString("DATABASE_NAME")

	return &config, nil
}

func InitViper() error {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warn().Msg("No .env config file found, relying on environment variables.")
			return nil
		}
		return fmt.Errorf("error reading config file: %w", err)
	}
	log.Info().Str("configFile", viper.ConfigFileUsed()).Msg("Successfully loaded configuration.")
	return nil
}
