package watcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileHandler interface {
	HandleFile(path string) error
}

type Watcher struct {
	watchPath       string
	extensions      []string
	settleTime      time.Duration
	handler         FileHandler
	fsWatcher       *fsnotify.Watcher
	pendingFiles    map[string]*fileState
	logger          *log.Logger
}

type fileState struct {
	path         string
	lastModified time.Time
	size         int64
	timer        *time.Timer
}

func New(watchPath string, extensions []string, settleTime int, handler FileHandler, logger *log.Logger) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating fsnotify watcher: %w", err)
	}

	w := &Watcher{
		watchPath:    watchPath,
		extensions:   extensions,
		settleTime:   time.Duration(settleTime) * time.Second,
		handler:      handler,
		fsWatcher:    fsWatcher,
		pendingFiles: make(map[string]*fileState),
		logger:       logger,
	}

	if err := fsWatcher.Add(watchPath); err != nil {
		return nil, fmt.Errorf("watching path %s: %w", watchPath, err)
	}

	w.logger.Printf("Watching: %s", watchPath)
	return w, nil
}

func (w *Watcher) Start() error {
	// Scan for existing files on startup
	if err := w.scanExisting(); err != nil {
		return fmt.Errorf("scanning existing files: %w", err)
	}

	// Watch for new files
	for {
		select {
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return nil
			}
			w.handleEvent(event)

		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return nil
			}
			w.logger.Printf("Watcher error: %v", err)
		}
	}
}

func (w *Watcher) Stop() {
	w.fsWatcher.Close()
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	// Only process video files
	if !w.isVideoFile(event.Name) {
		return
	}

	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		w.logger.Printf("New file detected: %s", event.Name)
		w.scheduleFileCheck(event.Name)

	case event.Op&fsnotify.Write == fsnotify.Write:
		// File is being written, reschedule check
		w.scheduleFileCheck(event.Name)

	case event.Op&fsnotify.Remove == fsnotify.Remove:
		// File was removed, cancel processing
		if state, exists := w.pendingFiles[event.Name]; exists {
			if state.timer != nil {
				state.timer.Stop()
			}
			delete(w.pendingFiles, event.Name)
			w.logger.Printf("File removed before processing: %s", event.Name)
		}
	}
}

func (w *Watcher) scheduleFileCheck(path string) {
	info, err := os.Stat(path)
	if err != nil {
		w.logger.Printf("Error stating file %s: %v", path, err)
		return
	}

	state, exists := w.pendingFiles[path]
	if exists && state.timer != nil {
		state.timer.Stop()
	}

	if !exists {
		state = &fileState{path: path}
		w.pendingFiles[path] = state
	}

	state.lastModified = info.ModTime()
	state.size = info.Size()

	// Schedule file processing after settle time
	state.timer = time.AfterFunc(w.settleTime, func() {
		w.processFile(path)
	})
}

func (w *Watcher) processFile(path string) {
	state, exists := w.pendingFiles[path]
	if !exists {
		return
	}

	// Verify file hasn't changed
	info, err := os.Stat(path)
	if err != nil {
		w.logger.Printf("Error checking file %s: %v", path, err)
		delete(w.pendingFiles, path)
		return
	}

	if info.ModTime() != state.lastModified || info.Size() != state.size {
		// File is still being modified, reschedule
		w.logger.Printf("File still changing: %s", path)
		w.scheduleFileCheck(path)
		return
	}

	// File is ready, process it
	delete(w.pendingFiles, path)
	w.logger.Printf("Processing file: %s", path)

	if err := w.handler.HandleFile(path); err != nil {
		w.logger.Printf("Error handling file %s: %v", path, err)
	}
}

func (w *Watcher) scanExisting() error {
	w.logger.Printf("Scanning for existing files in %s", w.watchPath)

	entries, err := os.ReadDir(w.watchPath)
	if err != nil {
		return fmt.Errorf("reading watch directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(w.watchPath, entry.Name())
		if w.isVideoFile(path) {
			w.logger.Printf("Found existing file: %s", path)
			w.scheduleFileCheck(path)
		}
	}

	return nil
}

func (w *Watcher) isVideoFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, validExt := range w.extensions {
		if ext == strings.ToLower(validExt) {
			return true
		}
	}
	return false
}
