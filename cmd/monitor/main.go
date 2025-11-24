package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dsu-teknik/peertube-monitor/pkg/config"
	"github.com/dsu-teknik/peertube-monitor/pkg/peertube"
	"github.com/dsu-teknik/peertube-monitor/pkg/watcher"
)

const version = "1.0.0"

func main() {
	configPath := flag.String("config", "config.json", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("PeerTube Monitor v%s\n", version)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Setup logging
	logger := setupLogger(cfg)
	logger.Printf("PeerTube Monitor v%s starting...", version)
	logger.Printf("Configuration loaded from: %s", *configPath)

	// Create PeerTube client
	client := peertube.NewClient(
		cfg.PeerTube.URL,
		cfg.PeerTube.Username,
		cfg.PeerTube.Password,
	)

	// Test authentication
	logger.Printf("Authenticating with PeerTube server: %s", cfg.PeerTube.URL)
	if err := client.Authenticate(); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	logger.Printf("Authentication successful")

	// Create upload handler
	handler := watcher.NewUploadHandler(client, cfg, logger)

	// Create and start watcher
	w, err := watcher.New(
		cfg.Watcher.WatchPath,
		cfg.Watcher.VideoExtensions,
		cfg.Watcher.SettleTime,
		handler,
		logger,
	)
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}
	defer w.Stop()

	logger.Printf("Monitoring folder: %s", cfg.Watcher.WatchPath)
	if cfg.Watcher.DonePath != "" {
		logger.Printf("Success folder: %s", cfg.Watcher.DonePath)
	} else {
		logger.Printf("Success action: Delete files")
	}
	if cfg.Watcher.FailedPath != "" {
		logger.Printf("Failed folder: %s", cfg.Watcher.FailedPath)
	} else {
		logger.Printf("Failed action: Rename with .failed extension")
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Printf("Shutdown signal received, stopping...")
		w.Stop()
		os.Exit(0)
	}()

	// Start watching
	if err := w.Start(); err != nil {
		log.Fatalf("Watcher error: %v", err)
	}
}

func setupLogger(cfg *config.Config) *log.Logger {
	var output *os.File

	if cfg.Logging.LogFile != "" {
		var err error
		output, err = os.OpenFile(cfg.Logging.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		// Note: We don't close this file as it's used throughout the program's lifetime
	} else {
		output = os.Stdout
	}

	logger := log.New(output, "", log.LstdFlags)

	if cfg.Logging.Verbose {
		logger.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	return logger
}
