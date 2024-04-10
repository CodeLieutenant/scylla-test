package config

import (
	"time"

	"github.com/gocql/gocql"
)

type ScyllaDB struct {
	Consistency       string        `json:"consistency"`
	Keyspace          string        `json:"keyspace"`
	DC                string        `json:"dc"`
	Hosts             []string      `json:"hosts"`
	ReconnectInterval time.Duration `json:"reconnect_interval"`
	ConnectionTimeout time.Duration `json:"connection_timeout"`
}

var defaultScyllaDBConfig = ScyllaDB{
	Consistency:       "LOCAL_QUORUM",
	Keyspace:          "test",
	ReconnectInterval: 10 * time.Second,
	ConnectionTimeout: 5 * time.Second,
	DC:                "DC1",
	Hosts:             []string{"127.0.0.1:9042", "127.0.0.1:9043"},
}

func (cfg *ScyllaDB) ToScyllaClusterConfig() (*gocql.ClusterConfig, error) {
	cluster := gocql.NewCluster(cfg.Hosts...)

	consistency := gocql.LocalQuorum

	if err := consistency.UnmarshalText([]byte(cfg.Consistency)); err != nil {
		return nil, err
	}

	cluster.Consistency = consistency
	cluster.ReconnectionPolicy = &gocql.ExponentialReconnectionPolicy{
		InitialInterval: 1 * time.Second,
		MaxInterval:     30 * time.Second,
		MaxRetries:      10,
	}
	cluster.ConnectTimeout = cfg.ConnectionTimeout
	cluster.Keyspace = cfg.Keyspace
	cluster.Authenticator = nil

	cluster.DisableSkipMetadata = false
	cluster.CQLVersion = "3.0.0"
	cluster.ReconnectInterval = cfg.ReconnectInterval
	cluster.Compressor = gocql.SnappyCompressor{}
	cluster.DefaultTimestamp = true

	fallback := gocql.RoundRobinHostPolicy()

	if cfg.DC != "" {
		fallback = gocql.DCAwareRoundRobinPolicy(cfg.DC)
	}

	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(fallback)
	cluster.ProtoVersion = 4

	return cluster, nil
}
