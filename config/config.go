package config

const Version = "1.0.0"

type Config struct {
	LogLevel  int    `default:"1"`
	LogFormat string `default:"json"`

	LogTarget string `default:"local"`
	SentryDsn string `required:"false"`
}

// получить кофигурацию из переменных окружения
func ReadConfig() (*Config, error) {
	config := &Config{}

	return config, nil
}
