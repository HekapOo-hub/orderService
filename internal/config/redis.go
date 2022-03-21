package config

import (
	"fmt"
	"github.com/caarlos0/env/v6"
)

const (
	// RedisStream is a name of stream for RedisHumanCacheRepository to listen
	RedisStream     = "prices"
	RedisCashStream = "cash"
)

// RedisConfig is used for connecting with redis db
type RedisConfig struct {
	Addr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	Password string `env:"REDIS_PASSWORD" envDefault:""`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
}

// NewRedisConfig returns instance of RedisConfig which is used for connecting to redis db
func NewRedisConfig() (*RedisConfig, error) {
	cfg := RedisConfig{}
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("error with parsing env variables in redis config %w", err)
	}
	return &cfg, nil
}
