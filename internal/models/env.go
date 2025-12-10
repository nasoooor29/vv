package models

import "github.com/kelseyhightower/envconfig"

type EnvVars struct {
	Port        string `envconfig:"PORT" default:"9999"`
	AppEnv      string `envconfig:"APP_ENV" default:"test"`
	DBPath      string `envconfig:"BLUEPRINT_DB_DATABASE" default:"visory.db"`
	APP_VERSION string `envconfig:"APP_VERSION" default:"0.0.2"`
}

var ENV_VARS EnvVars

func LoadEnv() EnvVars {
	var cfg EnvVars
	envconfig.MustProcess("", &cfg)
	ENV_VARS = cfg
	return cfg
}

func init() {
	LoadEnv()
}
