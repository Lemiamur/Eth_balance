package configs

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	App                 App           `yaml:"app"`
	GETBLOCK_API_KEY    string        `yaml:"Getblock_Api_Key" env:"GETBLOCK_API_KEY" env-required:"true"`
	HTTPClientTimeout   time.Duration `yaml:"http_client_timeout"`
	MaxIdleConns        int           `yaml:"max_idle_conns"`
	MaxIdleConnsPerHost int           `yaml:"max_idle_conns_per_host"`
	IdleConnTimeout     time.Duration `yaml:"idle_conn_timeout"`
	CacheSize           int           `yaml:"cache_size"`
	BlocksToAnalyze     int64         `yaml:"blocks_to_analyze"`
	BatchSize           int64         `yaml:"batch_size"`
	HTTP                HTTP          `yaml:"http"`
}

type App struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
}

type HTTP struct {
	Port string `env-required:"true" yaml:"port" env:"HTTP_PORT"`
}

func LoadConfig(path string) (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
