package config

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "sort"
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
    ChannelID          int             `json:"channelId,omitempty"`
    CategoryRaw        json.RawMessage `json:"category"`
    LicenceRaw         json.RawMessage `json:"licence"`
    Language           string          `json:"language"`
    PrivacyRaw         json.RawMessage `json:"privacy"`
    Description        string          `json:"description"`
    Tags               []string        `json:"tags"`
    DownloadEnabled    bool            `json:"downloadEnabled"`
    CommentsEnabled    bool            `json:"commentsEnabled"`
    WaitTranscoding    bool            `json:"waitTranscoding"`
    NSFW               bool            `json:"nsfw"`

    // Resolved integer values (populated after validation)
    Category int `json:"-"`
    Licence  int `json:"-"`
    Privacy  int `json:"-"`
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

// ResolveMetadata resolves category, licence, and privacy from string or int values
func (c *Config) ResolveMetadata(categories, licences, privacies map[string]string) error {
    var err error

    // Resolve category
    c.PeerTube.Defaults.Category, err = resolveField(
        "category",
        c.PeerTube.Defaults.CategoryRaw,
        categories,
    )
    if err != nil {
        return err
    }

    // Resolve licence
    c.PeerTube.Defaults.Licence, err = resolveField(
        "licence",
        c.PeerTube.Defaults.LicenceRaw,
        licences,
    )
    if err != nil {
        return err
    }

    // Resolve privacy
    c.PeerTube.Defaults.Privacy, err = resolveField(
        "privacy",
        c.PeerTube.Defaults.PrivacyRaw,
        privacies,
    )
    if err != nil {
        return err
    }

    return nil
}

func resolveField(fieldName string, raw json.RawMessage, mapping map[string]string) (int, error) {
    // Try parsing as integer first
    var intVal int
    if err := json.Unmarshal(raw, &intVal); err == nil {
        // Verify the integer ID exists in the mapping
        if _, exists := mapping[fmt.Sprintf("%d", intVal)]; !exists {
            return 0, fmt.Errorf("%s: invalid ID %d. Available options: %s", fieldName, intVal, formatMapping(mapping))
        }
        return intVal, nil
    }

    // Try parsing as string
    var strVal string
    if err := json.Unmarshal(raw, &strVal); err != nil {
        return 0, fmt.Errorf("%s: invalid value format (must be string or integer)", fieldName)
    }

    // Look up the string in the mapping (case-insensitive)
    for id, name := range mapping {
        if equalFold(name, strVal) {
            var idInt int
            fmt.Sscanf(id, "%d", &idInt)
            return idInt, nil
        }
    }

    return 0, fmt.Errorf("%s: unknown value %q. Available options: %s", fieldName, strVal, formatMapping(mapping))
}

func formatMapping(mapping map[string]string) string {
    // Create slice of names for sorting
    var names []string
    nameToID := make(map[string]string)
    for id, name := range mapping {
        names = append(names, name)
        nameToID[name] = id
    }

    // Sort alphabetically
    sort.Strings(names)

    // Build output in sorted order
    var items []string
    for _, name := range names {
        items = append(items, fmt.Sprintf("%s=%q", nameToID[name], name))
    }
    return "[" + joinStrings(items, ", ") + "]"
}

func equalFold(a, b string) bool {
    if len(a) != len(b) {
        return false
    }
    for i := 0; i < len(a); i++ {
        ca, cb := a[i], b[i]
        if ca >= 'A' && ca <= 'Z' {
            ca += 'a' - 'A'
        }
        if cb >= 'A' && cb <= 'Z' {
            cb += 'a' - 'A'
        }
        if ca != cb {
            return false
        }
    }
    return true
}

func joinStrings(items []string, sep string) string {
    if len(items) == 0 {
        return ""
    }
    result := items[0]
    for i := 1; i < len(items); i++ {
        result += sep + items[i]
    }
    return result
}
