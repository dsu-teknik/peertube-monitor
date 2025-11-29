//go:build windows

package main

import (
    "log"

    "github.com/dsu-teknik/peertube-monitor/pkg/watcher"
    "golang.org/x/sys/windows/svc"
    "golang.org/x/sys/windows/svc/debug"
)

type monitorService struct {
    watcher *watcher.Watcher
    logger  *log.Logger
}

func (m *monitorService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
    const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

    // Tell Windows we're starting
    changes <- svc.Status{State: svc.StartPending}

    // Start the watcher in a goroutine since it blocks
    errChan := make(chan error, 1)
    go func() {
        if err := m.watcher.Start(); err != nil {
            m.logger.Printf("Watcher error: %v", err)
            errChan <- err
        }
    }()

    // Tell Windows we're running
    changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
    m.logger.Printf("Service started successfully")

    // Wait for stop signal or error
loop:
    for {
        select {
        case c := <-r:
            switch c.Cmd {
            case svc.Interrogate:
                changes <- c.CurrentStatus
            case svc.Stop, svc.Shutdown:
                m.logger.Printf("Service stop requested")
                break loop
            default:
                m.logger.Printf("Unexpected service control request #%d", c)
            }
        case err := <-errChan:
            m.logger.Printf("Watcher stopped with error: %v", err)
            break loop
        }
    }

    // Tell Windows we're stopping
    changes <- svc.Status{State: svc.StopPending}
    m.watcher.Stop()
    m.logger.Printf("Service stopped")

    return
}

func runService(w *watcher.Watcher, logger *log.Logger) error {
    // Check if we're running as a service or interactively
    isService, err := svc.IsWindowsService()
    if err != nil {
        return err
    }

    if !isService {
        // Running interactively (e.g., from command line for testing)
        return debug.Run("PeerTubeMonitor", &monitorService{watcher: w, logger: logger})
    }

    // Running as a Windows service
    return svc.Run("PeerTubeMonitor", &monitorService{watcher: w, logger: logger})
}
