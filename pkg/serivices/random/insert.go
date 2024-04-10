package random

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/CodeLieutenant/scylladbtest/pkg/serivices/ratelimit"
	"github.com/gocql/gocql"
)

type Insert struct {
	session *gocql.Session
	limiter ratelimit.Limiter
	logger  *slog.Logger
}

func New(session *gocql.Session, limiter ratelimit.Limiter, logger *slog.Logger) *Insert {
	return &Insert{
		session: session,
		limiter: limiter,
		logger:  logger,
	}
}

func (r *Insert) Run(ctx context.Context) error {
	ch := r.limiter.Ready(ctx)
	query := r.session.Query("INSERT INTO randomdata(id, data) VALUES (?, ?)")
	defer query.Release()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("Exiting the random inserter")
			return nil
		case <-ch:
			id := gocql.UUIDFromTime(time.Now().UTC())
			data := rand.Int32N(10_000)

			if err := query.Bind(id, data).Exec(); err != nil {
				r.logger.Error("Failed to execute insert statement", err)
			}

			r.logger.Debug("Inserted random data",
				// Allocation can be avoided by using different logging library
				slog.String("id", id.String()),
				slog.Int("data", int(data)))
		}
	}
}

func (r *Insert) Error(a any) {
	r.logger.Error("There was an error", a)
}
