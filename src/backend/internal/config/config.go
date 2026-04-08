package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	BackendDomain         string        `mapstructure:"BACKEND_DOMAIN"`
	DatabaseURL           string        `mapstructure:"DATABASE_URL"`
	MigrationURL          string        `mapstructure:"MIGRATION_URL"`
	CookieSecret          string        `mapstructure:"COOKIE_SECRET"`
	IsProduction          bool          `mapstructure:"IS_PRODUCTION"`
	GoogleClientID        string        `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret    string        `mapstructure:"GOOGLE_CLIENT_SECRET"`
	GoogleClientRedirect  string        `mapstructure:"GOOGLE_CLIENT_REDIRECT"`
	AccessTokenDuration   time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration  time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	FrontendDomain        string        `mapstructure:"FRONTEND_DOMAIN"`
	FrontendLoginRoute    string        `mapstructure:"FRONTEND_LOGIN_ROUTE"`
	FrontendLobbyRoute    string        `mapstructure:"FRONTEND_LOBBY_ROUTE"`
	WebsocketPingInterval time.Duration `mapstructure:"WEBSOCKET_PING_INTERVAL"`
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
