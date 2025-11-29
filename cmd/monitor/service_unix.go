//go:build !windows

package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/dsu-teknik/peertube-monitor/pkg/watcher"
)

func runService(w *watcher.Watcher, logger *log.Logger) error {
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
    return w.Start()
}
