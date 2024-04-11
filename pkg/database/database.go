package database

import (
	"context"
	"log"
	"log/slog"

	"github.com/gocql/gocql"
)

type ToScyllaDBConfig interface {
	ToScyllaClusterConfig(*log.Logger) (*gocql.ClusterConfig, error)
}

func getLogLevelFromLogger(logger *slog.Logger) slog.Level {
	levels := []slog.Level{
		slog.LevelDebug,
		slog.LevelInfo,
		slog.LevelWarn,
		slog.LevelError,
	}

	for _, level := range levels {
		if logger.Handler().Enabled(context.Background(), level) {
			return level
		}
	}

	return slog.LevelInfo
}

func NewScyllaDBConnection(cfg ToScyllaDBConfig, logger *slog.Logger) (*gocql.Session, func(), error) {
	clusterConfig, err := cfg.ToScyllaClusterConfig(
		slog.NewLogLogger(logger.Handler().
			WithAttrs(
				[]slog.Attr{
					slog.String("database", "scylladb"),
				},
			),
			getLogLevelFromLogger(logger),
		),
	)
	if err != nil {
		return nil, func() {}, err
	}

	session, err := clusterConfig.CreateSession()
	if err != nil {
		return nil, func() {}, err
	}

	return session, func() { session.Close() }, nil
}
