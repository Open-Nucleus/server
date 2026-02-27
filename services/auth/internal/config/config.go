package config

import (
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	GRPCPort int            `koanf:"grpc_port"`
	JWT      JWTConfig      `koanf:"jwt"`
	Git      GitConfig      `koanf:"git"`
	Node     NodeConfig     `koanf:"node"`
	Devices  DevicesConfig  `koanf:"devices"`
	Roles    RolesConfig    `koanf:"roles"`
	Security SecurityConfig `koanf:"security"`
	KeyStore KeyStoreConfig `koanf:"keystore"`
	SQLite   SQLiteConfig   `koanf:"sqlite"`
	Logging  LoggingConfig  `koanf:"logging"`
}

type JWTConfig struct {
	Issuer          string        `koanf:"issuer"`
	AccessLifetime  time.Duration `koanf:"access_lifetime"`
	RefreshLifetime time.Duration `koanf:"refresh_lifetime"`
	ClockSkew       time.Duration `koanf:"clock_skew"`
}

type GitConfig struct {
	RepoPath    string `koanf:"repo_path"`
	AuthorName  string `koanf:"author_name"`
	AuthorEmail string `koanf:"author_email"`
}

type NodeConfig struct {
	KeyPath string `koanf:"key_path"`
	IDPath  string `koanf:"id_path"`
}

type DevicesConfig struct {
	Path string `koanf:"path"`
}

type RolesConfig struct {
	Path string `koanf:"path"`
}

type SecurityConfig struct {
	NonceTTL        time.Duration `koanf:"nonce_ttl"`
	MaxFailures     int           `koanf:"max_failures"`
	FailureWindow   time.Duration `koanf:"failure_window"`
	BootstrapSecret string        `koanf:"bootstrap_secret"`
}

type KeyStoreConfig struct {
	Type string `koanf:"type"` // "memory" or "file"
}

type SQLiteConfig struct {
	DBPath string `koanf:"db_path"`
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
	if err := k.Unmarshal("auth_service", &cfg); err != nil {
		return nil, err
	}

	setDefaults(&cfg)
	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.GRPCPort == 0 {
		cfg.GRPCPort = 50053
	}
	if cfg.JWT.Issuer == "" {
		cfg.JWT.Issuer = "open-nucleus-auth"
	}
	if cfg.JWT.AccessLifetime == 0 {
		cfg.JWT.AccessLifetime = 24 * time.Hour
	}
	if cfg.JWT.RefreshLifetime == 0 {
		cfg.JWT.RefreshLifetime = 7 * 24 * time.Hour
	}
	if cfg.JWT.ClockSkew == 0 {
		cfg.JWT.ClockSkew = 2 * time.Hour
	}
	if cfg.Git.RepoPath == "" {
		cfg.Git.RepoPath = "/var/lib/open-nucleus/data"
	}
	if cfg.Git.AuthorName == "" {
		cfg.Git.AuthorName = "open-nucleus-auth"
	}
	if cfg.Git.AuthorEmail == "" {
		cfg.Git.AuthorEmail = "auth@open-nucleus.local"
	}
	if cfg.Devices.Path == "" {
		cfg.Devices.Path = ".nucleus/devices"
	}
	if cfg.Security.NonceTTL == 0 {
		cfg.Security.NonceTTL = 60 * time.Second
	}
	if cfg.Security.MaxFailures == 0 {
		cfg.Security.MaxFailures = 10
	}
	if cfg.Security.FailureWindow == 0 {
		cfg.Security.FailureWindow = 60 * time.Second
	}
	if cfg.KeyStore.Type == "" {
		cfg.KeyStore.Type = "memory"
	}
	if cfg.SQLite.DBPath == "" {
		cfg.SQLite.DBPath = "/var/lib/open-nucleus/auth.db"
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}
}
