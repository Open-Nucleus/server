package config

import (
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	GRPCPort   int            `koanf:"grpc_port"`
	Git        GitConfig      `koanf:"git"`
	SQLite     SQLiteConfig   `koanf:"sqlite"`
	Validation ValidationConfig `koanf:"validation"`
	WriteLock  WriteLockConfig  `koanf:"write_lock"`
	Index      IndexConfig    `koanf:"index"`
	Matching   MatchingConfig `koanf:"matching"`
	Logging    LoggingConfig  `koanf:"logging"`
}

type GitConfig struct {
	RepoPath    string `koanf:"repo_path"`
	AuthorName  string `koanf:"author_name"`
	AuthorEmail string `koanf:"author_email"`
}

type SQLiteConfig struct {
	DBPath      string `koanf:"db_path"`
	JournalMode string `koanf:"journal_mode"`
	BusyTimeout int    `koanf:"busy_timeout"`
	CacheSize   int    `koanf:"cache_size"`
}

type ValidationConfig struct {
	StrictMode   bool `koanf:"strict_mode"`
	RequireICD10 bool `koanf:"require_icd10"`
	RequireLOINC bool `koanf:"require_loinc"`
}

type WriteLockConfig struct {
	Timeout time.Duration `koanf:"timeout"`
}

type IndexConfig struct {
	AutoRebuildOnDrift    bool `koanf:"auto_rebuild_on_drift"`
	HealthCheckOnStartup  bool `koanf:"health_check_on_startup"`
}

type MatchingConfig struct {
	DefaultThreshold float64 `koanf:"default_threshold"`
	MaxResults       int     `koanf:"max_results"`
	FuzzyMaxDistance int     `koanf:"fuzzy_max_distance"`
}

type LoggingConfig struct {
	Level  string `koanf:"level"`
	Format string `koanf:"format"`
}

func Load(path string) (*Config, error) {
	k := koanf.New(".")

	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return nil, err
	}

	var cfg Config
	if err := k.Unmarshal("patient_service", &cfg); err != nil {
		return nil, err
	}

	setDefaults(&cfg)
	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.GRPCPort == 0 {
		cfg.GRPCPort = 50051
	}
	if cfg.Git.RepoPath == "" {
		cfg.Git.RepoPath = "/var/lib/open-nucleus/data"
	}
	if cfg.Git.AuthorName == "" {
		cfg.Git.AuthorName = "open-nucleus"
	}
	if cfg.Git.AuthorEmail == "" {
		cfg.Git.AuthorEmail = "system@open-nucleus.local"
	}
	if cfg.SQLite.DBPath == "" {
		cfg.SQLite.DBPath = "/var/lib/open-nucleus/index.db"
	}
	if cfg.WriteLock.Timeout == 0 {
		cfg.WriteLock.Timeout = 5 * time.Second
	}
	if cfg.Matching.DefaultThreshold == 0 {
		cfg.Matching.DefaultThreshold = 0.7
	}
	if cfg.Matching.MaxResults == 0 {
		cfg.Matching.MaxResults = 10
	}
	if cfg.Matching.FuzzyMaxDistance == 0 {
		cfg.Matching.FuzzyMaxDistance = 2
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}
}
