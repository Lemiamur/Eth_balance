package configs

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type DefaultConfig struct {
	HTTPClientTimeout   time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
	CacheSize           int
	BlocksToAnalyze     int64
	BatchSize           int64
}

var Defaults = DefaultConfig{
	HTTPClientTimeout:   30 * time.Second,
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 100,
	IdleConnTimeout:     90 * time.Second,
	CacheSize:           100,
	BlocksToAnalyze:     100,
	BatchSize:           10,
}

type Config struct {
	GETBLOCK_API_KEY    string        `yaml:"Getblock_Api_Key" env:"GETBLOCK_API_KEY" env-required:"true"`
	HTTPClientTimeout   time.Duration `yaml:"http_client_timeout" env:"HTTP_CLIENT_TIMEOUT"`
	MaxIdleConns        int           `yaml:"max_idle_conns" env:"MAX_IDLE_CONNS"`
	MaxIdleConnsPerHost int           `yaml:"max_idle_conns_per_host" env:"MAX_IDLE_CONNS_PER_HOST"`
	IdleConnTimeout     time.Duration `yaml:"idle_conn_timeout" env:"IDLE_CONN_TIMEOUT"`
	CacheSize           int           `yaml:"cache_size" env:"CACHE_SIZE"`
	BlocksToAnalyze     int64         `yaml:"blocks_to_analyze" env:"BLOCKS_TO_ANALYZE"`
	BatchSize           int64         `yaml:"batch_size" env:"BATCH_SIZE"`
}

func LoadConfig(path string) (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	cfg := &Config{
		HTTPClientTimeout:   Defaults.HTTPClientTimeout,
		MaxIdleConns:        Defaults.MaxIdleConns,
		MaxIdleConnsPerHost: Defaults.MaxIdleConnsPerHost,
		IdleConnTimeout:     Defaults.IdleConnTimeout,
		CacheSize:           Defaults.CacheSize,
		BlocksToAnalyze:     Defaults.BlocksToAnalyze,
		BatchSize:           Defaults.BatchSize,
	}

	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
