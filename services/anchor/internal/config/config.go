package config

import (
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	GRPCPort int           `koanf:"grpc_port"`
	SQLite   SQLiteConfig  `koanf:"sqlite"`
	Git      GitConfig     `koanf:"git"`
	Logging  LoggingConfig `koanf:"logging"`
}

type SQLiteConfig struct {
	DBPath string `koanf:"db_path"`
}

type GitConfig struct {
	RepoPath    string `koanf:"repo_path"`
	AuthorName  string `koanf:"author_name"`
	AuthorEmail string `koanf:"author_email"`
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
	if err := k.Unmarshal("anchor_service", &cfg); err != nil {
		return nil, err
	}

	setDefaults(&cfg)
	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.GRPCPort == 0 {
		cfg.GRPCPort = 50055
	}
	if cfg.SQLite.DBPath == "" {
		cfg.SQLite.DBPath = "/var/lib/open-nucleus/anchor-queue.db"
	}
	if cfg.Git.AuthorName == "" {
		cfg.Git.AuthorName = "open-nucleus-anchor"
	}
	if cfg.Git.AuthorEmail == "" {
		cfg.Git.AuthorEmail = "anchor@open-nucleus.local"
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}
}
