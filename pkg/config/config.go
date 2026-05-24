package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)



type Config struct {
	ENV string `yaml:"env" env:"ENV" env-required:"true" env-default:"production"`
	RUN_SERVER string `yaml:"run_server"`
	ALLOWED_CLUSTERS map[string]string `yaml:"allowed_cluster_domains"`
	CAPACITY int `yaml:"capacity"`
}

func MustLoad() *Config  {
	config_path := os.Getenv("CONFIG_PATH")
	if config_path == "" {
		log.Fatal("Config Path is not configured in the env file")
	}
	var cfg Config
	err := cleanenv.ReadConfig(config_path, &cfg)
	if err != nil {
		log.Fatalf("Cannot read config file: %s", err.Error())
	}

	return &cfg
}