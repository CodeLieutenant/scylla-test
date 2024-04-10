package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CodeLieutenant/scylladbtest/pkg/pool"

	"github.com/CodeLieutenant/scylladbtest/pkg/config"
	"github.com/CodeLieutenant/scylladbtest/pkg/utils"
)

func run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Done")
			return
		default:
			fmt.Println("Hello World")
			time.Sleep(5 * time.Second)
		}
	}
}

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
		log.Fatalf("Failed to parse config: %v", err)
	}

	log.Printf("Config: %s %v", configFile, cfg)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	defer cancel()

	clusterConfig, err := cfg.ScyllaDB.ToScyllaClusterConfig()
	if err != nil {
		log.Fatalf("Failed to create ScyllaDB cluster config: %v", err)
	}

	session, err := clusterConfig.CreateSession()
	if err != nil {
		log.Fatalf("Failed to create ScyllaDB session: %v", err)
	}

	defer session.Close()

	wp := pool.New(utils.Parallelism(parallelism))
	wp.Start(ctx, run, nil)

	<-ctx.Done()

	if err := wp.Close(); err != nil {
		log.Fatalf("Failed to close worker pool: %v", err)
	}
}
