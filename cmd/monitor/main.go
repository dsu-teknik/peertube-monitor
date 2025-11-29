package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "sort"

    "github.com/dsu-teknik/peertube-monitor/pkg/config"
    "github.com/dsu-teknik/peertube-monitor/pkg/peertube"
    "github.com/dsu-teknik/peertube-monitor/pkg/watcher"
)

const version = "1.0.0"

func main() {
    configPath := flag.String("config", "config.json", "Path to configuration file")
    logFile := flag.String("log", "", "Path to log file (default: stdout)")
    verbose := flag.Bool("verbose", false, "Enable verbose logging")
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
    logger := setupLogger(*logFile, *verbose)
    logger.Printf("PeerTube Monitor v%s starting...", version)
    logger.Printf("Configuration loaded from: %s", *configPath)
    logger.Printf("Credentials loaded from: %s", cfg.GetCredentialSource())

    // Create PeerTube client
    client := peertube.NewClient(
        cfg.PeerTube.URL,
        cfg.PeerTube.Username,
        cfg.PeerTube.Password,
    )

    // Validate credentials are configured
    if cfg.PeerTube.URL == "" || cfg.PeerTube.Username == "" || cfg.PeerTube.Password == "" {
        logger.Printf("WARNING: PeerTube credentials not configured!")
        logger.Printf("Please edit config file and restart service")
        logger.Printf("Config location: %s", *configPath)
    } else {
        // Test authentication only if credentials are provided
        logger.Printf("Authenticating with PeerTube server: %s", cfg.PeerTube.URL)
        if err := client.Authenticate(); err != nil {
            logger.Printf("WARNING: Authentication failed: %v", err)
            logger.Printf("Service will start but uploads will fail until credentials are fixed")
        } else {
            logger.Printf("Authentication successful")

            // Fetch metadata from PeerTube
            logger.Printf("Fetching video metadata from PeerTube server...")
            metadata, err := client.FetchMetadata()
            if err != nil {
                logger.Printf("WARNING: Failed to fetch metadata: %v", err)
                logger.Printf("Will use default values from config")
            } else {
                logger.Printf("Available categories: %d options", len(metadata.Categories))
                if *verbose {
                    logSortedMetadata(logger, metadata.Categories)
                }

                logger.Printf("Available licences: %d options", len(metadata.Licences))
                if *verbose {
                    logSortedMetadata(logger, metadata.Licences)
                }

                logger.Printf("Available privacy levels: %d options", len(metadata.Privacies))
                if *verbose {
                    logSortedMetadata(logger, metadata.Privacies)
                }

                // Resolve metadata in config
                if err := cfg.ResolveMetadata(metadata.Categories, metadata.Licences, metadata.Privacies); err != nil {
                    logger.Printf("WARNING: Invalid configuration: %v", err)
                    logger.Printf("Will use raw values from config")
                } else {
                    logger.Printf("Video defaults: category=%q, licence=%q, privacy=%q",
                        metadata.Categories[fmt.Sprintf("%d", cfg.PeerTube.Defaults.Category)],
                        metadata.Licences[fmt.Sprintf("%d", cfg.PeerTube.Defaults.Licence)],
                        metadata.Privacies[fmt.Sprintf("%d", cfg.PeerTube.Defaults.Privacy)])
                }
            }
        }
    }

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

    // Run service (platform-specific implementation)
    if err := runService(w, logger); err != nil {
        log.Fatalf("Service error: %v", err)
    }
}

func setupLogger(logFile string, verbose bool) *log.Logger {
    var output *os.File

    if logFile != "" {
        var err error
        output, err = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
        if err != nil {
            log.Fatalf("Failed to open log file: %v", err)
        }
        // Note: We don't close this file as it's used throughout the program's lifetime
    } else {
        output = os.Stdout
    }

    logger := log.New(output, "", log.LstdFlags)

    if verbose {
        logger.SetFlags(log.LstdFlags | log.Lshortfile)
    }

    return logger
}

func logSortedMetadata(logger *log.Logger, metadata map[string]string) {
    // Create slice of names for sorting
    var names []string
    nameToID := make(map[string]string)
    for id, name := range metadata {
        names = append(names, name)
        nameToID[name] = id
    }

    // Sort alphabetically
    sort.Strings(names)

    // Log in sorted order
    for _, name := range names {
        logger.Printf("  - %s: %q", nameToID[name], name)
    }
}
