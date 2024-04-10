package main

import (
	"context"
	"flag"
	"io"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CodeLieutenant/scylladbtest/pkg/config"
	"github.com/CodeLieutenant/scylladbtest/pkg/database"
	"github.com/CodeLieutenant/scylladbtest/pkg/pool"
	"github.com/CodeLieutenant/scylladbtest/pkg/serivices/random"
	"github.com/CodeLieutenant/scylladbtest/pkg/serivices/ratelimit"
	"github.com/CodeLieutenant/scylladbtest/pkg/utils"
)

var (
	parallelism  int
	reqPerSec    int
	scyllaDBHost string
	configFile   string
)

func parseConfig() (*config.Config, error) {
	var reader io.Reader

	if _, err := os.Stat(configFile); err == nil {
		file, err := os.OpenFile(configFile, os.O_RDONLY, 0o644)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		reader = file
	}

	return config.New(reader)
}

func main() {
	pid := os.Getpid()
	cwd, _ := os.Getwd()

	log.Printf("Stating the application: %d", pid)
	log.Printf("Current directory: %s", cwd)

	flag.IntVar(&parallelism, "parallelism", utils.Parallelism(), "Maximum parallelism")
	flag.IntVar(&reqPerSec, "req", 1, "Max Requests per second to ScyllaDB")
	flag.StringVar(&scyllaDBHost, "host", "127.0.0.1:9042", "ScyllaDB host")
	flag.StringVar(&configFile, "config", "config.json", "JSON Configuration file")

	flag.Parse()

	cfg, err := parseConfig()
	if err != nil {
		log.Panicf("Failed to parse config: %v", err)
	}

	log.Printf("Config: %s %v", configFile, cfg)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	defer cancel()

	session, cleanup, err := database.NewScyllaDBConnection(&cfg.ScyllaDB)
	defer cleanup()

	if err != nil {
		log.Panicf("Failed to create ScyllaDB cluster config: %v", err)
	}

	limiter := ratelimit.NewLeakyBucket(1, utils.Parallelism(parallelism), 1*time.Second)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	inserter := random.New(session, limiter, logger)

	wp := pool.New(utils.Parallelism(parallelism))
	wp.Start(ctx, inserter)

	<-ctx.Done()

	if err := wp.Close(); err != nil {
		log.Panicf("Failed to close worker pool: %v", err)
	}
}
