package models

import "github.com/kelseyhightower/envconfig"

type EnvVars struct {
	Port   string `envconfig:"PORT" default:"8080"`
	AppEnv string `envconfig:"APP_ENV" default:"local"`
	DBPath string `envconfig:"BLUEPRINT_DB_DATABASE" default:"visory.db"`
}

func LoadEnv() EnvVars {
	var cfg EnvVars
	envconfig.MustProcess("", &cfg)
	return cfg
}
