package config

import (
	"encoding/json"
	"io"
)

type Config struct {
	ScyllaDB ScyllaDB `json:"scylladb"`
}

var defaultConfig = Config{
	ScyllaDB: defaultScyllaDBConfig,
}

func New(input io.Reader) (*Config, error) {
	cfg := defaultConfig

	if input == nil {
		return &cfg, nil
	}

	if err := json.NewDecoder(input).Decode(&cfg); err != nil {
		return nil, err
	}

	// let it escape to heap (large object)
	return &cfg, nil
}
