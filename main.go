package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/CodeLieutenant/scylladbtest/pkg/config"
	"github.com/CodeLieutenant/scylladbtest/pkg/database"
	"github.com/CodeLieutenant/scylladbtest/pkg/pool"
	"github.com/CodeLieutenant/scylladbtest/pkg/services/metrics"
	"github.com/CodeLieutenant/scylladbtest/pkg/services/random"
	"github.com/CodeLieutenant/scylladbtest/pkg/services/ratelimit"
	"github.com/CodeLieutenant/scylladbtest/pkg/utils"
)

var (
	parallelism int
	requests    int
	timeFrame   time.Duration
	configFile  string
	logLevel    string
)

func parseConfig() (*config.Config, error) {
	var reader io.Reader

	if _, err := os.Stat(configFile); err == nil {
		file, err := os.OpenFile(configFile, os.O_RDONLY, 0o644)
		if err != nil {
			return nil, err
		}

		defer func() {
			// we can here ignore the close error
			// file close, never fails, even if it does
			// only config file is affected, after
			// program exits, system will do cleanup
			_ = file.Close()
		}()
		reader = file
	}

	return config.New(reader)
}

func main() {
	flag.IntVar(&parallelism, "parallelism", utils.Parallelism(), "Maximum parallelism")
	flag.IntVar(&requests, "requests", 1, "Max Requests per second to ScyllaDB")
	flag.StringVar(&configFile, "config", "config.json", "JSON Configuration file")
	flag.StringVar(&logLevel, "log-level", "info", "Logging level (debug, info, warn, error)")
	flag.DurationVar(&timeFrame, "time-frame", 1*time.Second, "Time frame between requests")

	flag.Parse()

	parallelism = utils.Parallelism(parallelism)

	var level slog.Level

	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		log.Printf("Failed to parse logging level: %s using Info as default", logLevel)
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}))

	cfg, err := parseConfig()
	if err != nil {
		log.Panicf("Failed to parse config: %v", err)
	}

	pid := os.Getpid()
	cwd, _ := os.Getwd()

	logger.Info("Stating the application",
		slog.Int("pid", pid),
		slog.String("cwd", cwd),
		slog.String("config_file", configFile),
		slog.Int("cpu_cores", runtime.NumCPU()),
		slog.Int("go_max_procs", runtime.GOMAXPROCS(0)),
		slog.String("go_compiler", runtime.Version()),
	)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	defer cancel()

	session, cleanup, err := database.NewScyllaDBConnection(&cfg.ScyllaDB, logger)
	defer cleanup()

	if err != nil {
		log.Panicf("Failed to create ScyllaDB cluster config: %v", err)
	}

	limiter := ratelimit.NewLeakyBucket(int64(requests), timeFrame)
	m := metrics.New(10_000)
	inserter := random.New(session, m, limiter, logger)

	wp := pool.New(parallelism)
	go wp.Start(ctx, inserter)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_, _ = fmt.Fprintln(os.Stdout, m.Collect().String())
			}
		}
	}()

	logger.Info("Waiting for SIGTERM or SIGINT to exit...")
	defer logger.Info("Exiting program...")
	<-ctx.Done()

	if err := wp.Close(); err != nil {
		log.Panicf("Failed to close worker pool: %v", err)
	}
}
