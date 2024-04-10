package database

import (
	"github.com/gocql/gocql"
)

type ToScyllaDBConfig interface {
	ToScyllaClusterConfig() (*gocql.ClusterConfig, error)
}

func NewScyllaDBConnection(cfg ToScyllaDBConfig) (*gocql.Session, func(), error) {
	clusterConfig, err := cfg.ToScyllaClusterConfig()
	if err != nil {
		return nil, func() {}, err
	}

	session, err := clusterConfig.CreateSession()
	if err != nil {
		return nil, func() {}, err
	}

	return session, func() { session.Close() }, nil
}
