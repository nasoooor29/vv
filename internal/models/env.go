package models

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type EnvVars struct {
	Port              string `envconfig:"PORT" default:"9999"`
	AppEnv            string `envconfig:"APP_ENV" default:"test"`
	DBPath            string `envconfig:"BLUEPRINT_DB_DATABASE" default:"visory.db"`
	APP_VERSION       string `envconfig:"APP_VERSION" default:"0.0.2"`
	GoogleOAuthKey    string `envconfig:"GOOGLE_OAUTH_KEY" required:"true"`
	GoogleOAuthSecret string `envconfig:"GOOGLE_OAUTH_SECRET" required:"true"`
	GithubOAuthKey    string `envconfig:"GITHUB_OAUTH_KEY" required:"true"`
	GithubOAuthSecret string `envconfig:"GITHUB_OAUTH_SECRET" required:"true"`
	// OAuthCallbackURL  string `envconfig:"OAUTH_CALLBACK_URL" default:"http://localhost:9999/api/auth/oauth/callback" required:"true"`
	BaseUrl         string `envconfig:"BASE_URL" default:"http://localhost"`
	SessionSecret   string `envconfig:"SESSION_SECRET" required:"true"`
	FRONTEND_DASH   string `envconfig:"FRONTEND_DASH_URL" default:"http://localhost:5173/app"`
	BaseUrlWithPort string
}

var ENV_VARS EnvVars

// LoadEnvProduction loads environment variables strictly for production
func LoadEnvProduction() EnvVars {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		slog.Warn("error loading .env file, using environment variables", "err", err)
	}

	var cfg EnvVars
	err = envconfig.Process("", &cfg)
	if err != nil && !IsCiEnvironment() {
		slog.Error("error processing environment variables", "err", err)
		panic(err)
	}

	ENV_VARS = cfg
	ENV_VARS.BaseUrlWithPort = ENV_VARS.BaseUrl + ":" + ENV_VARS.Port
	return cfg
}

func IsCiEnvironment() bool {
	return os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != ""
}
