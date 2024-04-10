package config

import (
	"log"
	"os"
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
	Consistency:       "QUORUM",
	Keyspace:          "scyllatest",
	ReconnectInterval: 5 * time.Second,
	ConnectionTimeout: 10 * time.Second,
	DC:                "DC1",
	Hosts:             []string{"127.0.0.1:9042"},
}

func (cfg *ScyllaDB) ToScyllaClusterConfig() (*gocql.ClusterConfig, error) {
	cluster := gocql.NewCluster(cfg.Hosts...)
	consistency := gocql.Quorum
	if err := consistency.UnmarshalText([]byte(cfg.Consistency)); err != nil {
		return nil, err
	}

	cluster.Consistency = consistency
	cluster.ConnectTimeout = cfg.ConnectionTimeout
	cluster.Authenticator = nil
	cluster.ReconnectInterval = cfg.ReconnectInterval
	cluster.Compressor = gocql.SnappyCompressor{}
	cluster.DefaultTimestamp = true
	cluster.DisableSkipMetadata = false
	cluster.Keyspace = cfg.Keyspace
	cluster.Logger = log.New(os.Stderr, "[ScyllaDB] ", log.LstdFlags|log.Lshortfile|log.LUTC|log.Lmsgprefix)

	fallback := gocql.RoundRobinHostPolicy()

	if cfg.DC != "" {
		fallback = gocql.DCAwareRoundRobinPolicy(cfg.DC)
	}

	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(fallback)
	cluster.ProtoVersion = 4

	return cluster, nil
}
