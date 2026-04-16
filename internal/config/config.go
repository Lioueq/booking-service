package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Env     string `mapstructure:"env"`
	Address string `mapstructure:"address"`
}

func MustLoad() *Config {
	cfgPath := os.Getenv("CONFIG_PATH")

	if cfgPath == "" {
		log.Fatal("Invalid config path")
	}

	if _, err := os.Stat(cfgPath); err != nil {
		log.Fatal("Empty config")
	}

	var cfg Config

	viper.SetConfigFile(cfgPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Config read error: %v", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}
	return &cfg
}
