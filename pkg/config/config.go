package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	PeerTube PeerTubeConfig `json:"peertube"`
	Watcher  WatcherConfig  `json:"watcher"`
	Logging  LoggingConfig  `json:"logging"`
}

type PeerTubeConfig struct {
	URL      string        `json:"url"`
	Username string        `json:"username"`
	Password string        `json:"password"`
	Defaults VideoDefaults `json:"defaults"`
}

type VideoDefaults struct {
	Category           int      `json:"category"`
	Licence            int      `json:"licence"`
	Language           string   `json:"language"`
	Privacy            int      `json:"privacy"`
	Description        string   `json:"description"`
	Tags               []string `json:"tags"`
	DownloadEnabled    bool     `json:"downloadEnabled"`
	CommentsEnabled    bool     `json:"commentsEnabled"`
	WaitTranscoding    bool     `json:"waitTranscoding"`
	NSFW               bool     `json:"nsfw"`
}

type WatcherConfig struct {
	WatchPath      string   `json:"watchPath"`
	DonePath       string   `json:"donePath"`
	FailedPath     string   `json:"failedPath"`
	VideoExtensions []string `json:"videoExtensions"`
	SettleTime     int      `json:"settleTime"` // seconds to wait for file to stop changing
	MaxRetries     int      `json:"maxRetries"`
}

type LoggingConfig struct {
	LogFile  string `json:"logFile"`
	Verbose  bool   `json:"verbose"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Set defaults
	if cfg.Watcher.SettleTime == 0 {
		cfg.Watcher.SettleTime = 5
	}
	if cfg.Watcher.MaxRetries == 0 {
		cfg.Watcher.MaxRetries = 3
	}
	if len(cfg.Watcher.VideoExtensions) == 0 {
		cfg.Watcher.VideoExtensions = []string{".mp4", ".webm", ".mkv", ".avi", ".mov", ".flv"}
	}

	// Override with environment variables if present
	cfg.loadFromEnv()

	// Ensure paths are absolute
	if cfg.Watcher.WatchPath != "" && !filepath.IsAbs(cfg.Watcher.WatchPath) {
		cfg.Watcher.WatchPath, _ = filepath.Abs(cfg.Watcher.WatchPath)
	}
	if cfg.Watcher.DonePath != "" && !filepath.IsAbs(cfg.Watcher.DonePath) {
		cfg.Watcher.DonePath, _ = filepath.Abs(cfg.Watcher.DonePath)
	}
	if cfg.Watcher.FailedPath != "" && !filepath.IsAbs(cfg.Watcher.FailedPath) {
		cfg.Watcher.FailedPath, _ = filepath.Abs(cfg.Watcher.FailedPath)
	}

	return &cfg, nil
}

func (c *Config) loadFromEnv() {
	// Load PeerTube credentials from environment variables if not set in config
	if username := os.Getenv("PEERTUBE_USERNAME"); username != "" {
		c.PeerTube.Username = username
	}
	if password := os.Getenv("PEERTUBE_PASSWORD"); password != "" {
		c.PeerTube.Password = password
	}
	if url := os.Getenv("PEERTUBE_URL"); url != "" {
		c.PeerTube.URL = url
	}
}

func (c *Config) GetCredentialSource() string {
	username := os.Getenv("PEERTUBE_USERNAME")
	password := os.Getenv("PEERTUBE_PASSWORD")

	if username != "" && password != "" {
		return "environment variables"
	}
	if username != "" || password != "" {
		return "mixed (config file + environment variables)"
	}
	return "config file"
}

func (c *Config) Validate() error {
	if c.PeerTube.URL == "" {
		return fmt.Errorf("peertube.url is required")
	}
	if c.PeerTube.Username == "" {
		return fmt.Errorf("peertube.username is required")
	}
	if c.PeerTube.Password == "" {
		return fmt.Errorf("peertube.password is required")
	}
	if c.Watcher.WatchPath == "" {
		return fmt.Errorf("watcher.watchPath is required")
	}

	// Create directories if they don't exist
	for _, path := range []string{c.Watcher.WatchPath, c.Watcher.DonePath, c.Watcher.FailedPath} {
		if path != "" {
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", path, err)
			}
		}
	}

	return nil
}
