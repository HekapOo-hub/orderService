package config

import (
	"fmt"
	"github.com/caarlos0/env/v6"
)

type PositionPort struct {
	Value string `env:"POSITION_PORT" envDefault:":50005"`
}

// GetPositionPort returns position server port for grpc calls
func GetPositionPort() (*PositionPort, error) {
	cfg := PositionPort{}
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("error with parsing env variables in position port config %w", err)
	}
	return &cfg, nil
}
