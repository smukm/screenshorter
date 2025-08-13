package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
)

const Version = "1.0.0"

type Config struct {
	Port        string `required:"true" default:"8033"`
	AccessToken string `required:"true" default:"12345"`

	LogLevel  int    `default:"1"`
	LogFormat string `default:"json"`

	LogTarget string `default:"local"`
	SentryDsn string `required:"false"`

	MaxWorkers int `default:"5"`
}

// получить кофигурацию из переменных окружения
func ReadConfig() (*Config, error) {
	config := &Config{}

	err := envconfig.Process("SS", config)
	if err != nil {
		return nil, fmt.Errorf("configs error: %w", err)
	}

	return config, nil
}
