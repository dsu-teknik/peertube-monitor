package watcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dsu-teknik/peertube-monitor/pkg/config"
	"github.com/dsu-teknik/peertube-monitor/pkg/peertube"
)

type UploadHandler struct {
	client     *peertube.Client
	config     *config.Config
	logger     *log.Logger
	retryCount map[string]int
}

func NewUploadHandler(client *peertube.Client, cfg *config.Config, logger *log.Logger) *UploadHandler {
	return &UploadHandler{
		client:     client,
		config:     cfg,
		logger:     logger,
		retryCount: make(map[string]int),
	}
}

func (h *UploadHandler) HandleFile(path string) error {
	h.logger.Printf("Starting upload: %s", path)

	// Extract video name from filename (without extension)
	filename := filepath.Base(path)
	videoName := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Build video attributes from config defaults
	attrs := peertube.VideoAttributes{
		Name:            videoName,
		Category:        h.config.PeerTube.Defaults.Category,
		Licence:         h.config.PeerTube.Defaults.Licence,
		Language:        h.config.PeerTube.Defaults.Language,
		Privacy:         h.config.PeerTube.Defaults.Privacy,
		Description:     h.config.PeerTube.Defaults.Description,
		Tags:            h.config.PeerTube.Defaults.Tags,
		DownloadEnabled: h.config.PeerTube.Defaults.DownloadEnabled,
		CommentsEnabled: h.config.PeerTube.Defaults.CommentsEnabled,
		WaitTranscoding: h.config.PeerTube.Defaults.WaitTranscoding,
		NSFW:            h.config.PeerTube.Defaults.NSFW,
	}

	// Attempt upload
	result, err := h.client.Upload(path, attrs)
	if err != nil {
		return h.handleFailure(path, err)
	}

	h.logger.Printf("Upload successful: %s (UUID: %s)", result.Video.Name, result.Video.UUID)

	// Move to done folder or delete
	return h.handleSuccess(path)
}

func (h *UploadHandler) handleSuccess(path string) error {
	if h.config.Watcher.DonePath != "" {
		// Move to done folder
		destPath := filepath.Join(h.config.Watcher.DonePath, filepath.Base(path))

		// Handle filename collision
		destPath = h.ensureUniqueFilename(destPath)

		if err := os.Rename(path, destPath); err != nil {
			h.logger.Printf("Error moving file to done folder: %v", err)
			// Try copying instead
			if err := h.copyFile(path, destPath); err != nil {
				return fmt.Errorf("copying to done folder: %w", err)
			}
			if err := os.Remove(path); err != nil {
				h.logger.Printf("Warning: could not remove original file: %v", err)
			}
		}
		h.logger.Printf("Moved to done: %s", destPath)
	} else {
		// Delete file
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("deleting file: %w", err)
		}
		h.logger.Printf("Deleted: %s", path)
	}

	// Clear retry count
	delete(h.retryCount, path)
	return nil
}

func (h *UploadHandler) handleFailure(path string, uploadErr error) error {
	h.logger.Printf("Upload failed: %v", uploadErr)

	// Increment retry count
	h.retryCount[path]++
	retries := h.retryCount[path]

	if retries < h.config.Watcher.MaxRetries {
		h.logger.Printf("Will retry (%d/%d)", retries, h.config.Watcher.MaxRetries)
		return uploadErr
	}

	// Max retries reached, move to failed folder
	h.logger.Printf("Max retries reached, moving to failed folder")

	if h.config.Watcher.FailedPath != "" {
		destPath := filepath.Join(h.config.Watcher.FailedPath, filepath.Base(path))
		destPath = h.ensureUniqueFilename(destPath)

		if err := os.Rename(path, destPath); err != nil {
			h.logger.Printf("Error moving file to failed folder: %v", err)
			// Try copying instead
			if err := h.copyFile(path, destPath); err != nil {
				return fmt.Errorf("copying to failed folder: %w", err)
			}
			if err := os.Remove(path); err != nil {
				h.logger.Printf("Warning: could not remove original file: %v", err)
			}
		}
		h.logger.Printf("Moved to failed: %s", destPath)
	} else {
		// Rename with .failed extension
		failedPath := path + ".failed"
		if err := os.Rename(path, failedPath); err != nil {
			return fmt.Errorf("renaming to .failed: %w", err)
		}
		h.logger.Printf("Renamed to: %s", failedPath)
	}

	// Clear retry count
	delete(h.retryCount, path)
	return nil
}

func (h *UploadHandler) ensureUniqueFilename(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	// File exists, add counter
	ext := filepath.Ext(path)
	nameWithoutExt := strings.TrimSuffix(path, ext)

	for i := 1; ; i++ {
		newPath := fmt.Sprintf("%s_%d%s", nameWithoutExt, i, ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}

func (h *UploadHandler) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := destFile.ReadFrom(sourceFile); err != nil {
		return err
	}

	return destFile.Sync()
}
