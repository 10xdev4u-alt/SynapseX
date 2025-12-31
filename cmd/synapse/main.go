package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/princetheprogrammer/synapse/internal/config"
	"github.com/princetheprogrammer/synapse/internal/logger"
	"github.com/princetheprogrammer/synapse/pkg/node"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var (
		configPath  string
		showVersion bool
		logLevel    string
		logFormat   string
		port        int
	)

	flag.StringVar(&configPath, "config", "", "path to configuration file")
	flag.BoolVar(&showVersion, "version", false, "show version information")
	flag.StringVar(&logLevel, "log-level", "", "log level (debug, info, warn, error)")
	flag.StringVar(&logFormat, "log-format", "", "log format (json, console)")
	flag.IntVar(&port, "port", 0, "P2P listen port (overrides config)")
	flag.Parse()

	if showVersion {
		fmt.Printf("synapse version %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built: %s\n", date)
		os.Exit(0)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	if logLevel != "" {
		cfg.Logging.Level = logLevel
	}
	if logFormat != "" {
		cfg.Logging.Format = logFormat
	}
	if port > 0 {
		cfg.P2P.ListenPort = port
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "invalid configuration: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.New(cfg.Logging.Level, cfg.Logging.Format, cfg.Logging.OutputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	log.Infof("starting synapse version %s", version)

	n, err := node.New(cfg, log)
	if err != nil {
		log.Fatalf("failed to create node: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := n.Start(ctx); err != nil {
		log.Fatalf("failed to start node: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	log.Info("synapse is running, press Ctrl+C to stop")

	sig := <-sigCh
	log.Infof("received signal: %s, initiating shutdown", sig)

	cancel()

	if err := n.Stop(); err != nil {
		log.Errorf("error during shutdown: %v", err)
		os.Exit(1)
	}

	n.Wait()
	log.Info("synapse stopped successfully")
}

func loadConfig(configPath string) (*config.Config, error) {
	if configPath != "" {
		return config.Load(configPath)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config.Default(), nil
	}

	defaultPath := filepath.Join(homeDir, ".synapse", "config.json")
	return config.Load(defaultPath)
}
