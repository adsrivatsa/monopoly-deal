package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	BackendPort  string `mapstructure:"BACKEND_PORT"`
	DatabaseURL  string `mapstructure:"DATABASE_URL"`
	MigrationURL string `mapstructure:"MIGRATION_URL"`
}

func Load(envPath string) (Config, error) {
	viper.SetConfigFile(envPath)
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	var cfg Config

	err := viper.ReadInConfig()
	if err != nil {
		return cfg, err
	}

	err = viper.Unmarshal(&cfg)
	return cfg, err
}
