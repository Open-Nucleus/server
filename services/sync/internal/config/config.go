package config

import (
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	GRPCPort   int              `koanf:"grpc_port"`
	Git        GitConfig        `koanf:"git"`
	Transports TransportsConfig `koanf:"transports"`
	Sync       SyncConfig       `koanf:"sync"`
	Priority   PriorityConfig   `koanf:"priority"`
	Merge      MergeConfig      `koanf:"merge"`
	Discovery  DiscoveryConfig  `koanf:"discovery"`
	History    HistoryConfig    `koanf:"history"`
	Events     EventsConfig     `koanf:"events"`
	Node       NodeConfig       `koanf:"node"`
	Logging    LoggingConfig    `koanf:"logging"`
}

type GitConfig struct {
	RepoPath    string `koanf:"repo_path"`
	AuthorName  string `koanf:"author_name"`
	AuthorEmail string `koanf:"author_email"`
}

type TransportsConfig struct {
	LocalNetwork TransportEntry `koanf:"local_network"`
	WiFiDirect   TransportEntry `koanf:"wifi_direct"`
	Bluetooth    TransportEntry `koanf:"bluetooth"`
	USB          TransportEntry `koanf:"usb"`
}

type TransportEntry struct {
	Enabled     bool   `koanf:"enabled"`
	MDNSService string `koanf:"mdns_service"`
	MDNSDomain  string `koanf:"mdns_domain"`
	Port        int    `koanf:"port"`
}

type SyncConfig struct {
	MaxConcurrent    int           `koanf:"max_concurrent"`
	Cooldown         time.Duration `koanf:"cooldown"`
	HandshakeTimeout time.Duration `koanf:"handshake_timeout"`
	TransferTimeout  time.Duration `koanf:"transfer_timeout"`
	ChunkSize        int           `koanf:"chunk_size"`
}

type PriorityConfig struct {
	Enabled              bool `koanf:"enabled"`
	BluetoothConstrained bool `koanf:"bluetooth_constrained"`
}

type MergeConfig struct {
	DrugInteractionCheck bool `koanf:"drug_interaction_check"`
}

type DiscoveryConfig struct {
	ScanInterval time.Duration `koanf:"scan_interval"`
	PeerTTL      time.Duration `koanf:"peer_ttl"`
}

type HistoryConfig struct {
	DBPath     string `koanf:"db_path"`
	MaxEntries int    `koanf:"max_entries"`
}

type EventsConfig struct {
	BufferSize int `koanf:"buffer_size"`
}

type NodeConfig struct {
	ID string `koanf:"id"`
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
	if err := k.Unmarshal("sync_service", &cfg); err != nil {
		return nil, err
	}

	setDefaults(&cfg)
	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.GRPCPort == 0 {
		cfg.GRPCPort = 50052
	}
	if cfg.Git.RepoPath == "" {
		cfg.Git.RepoPath = "/var/lib/open-nucleus/data"
	}
	if cfg.Git.AuthorName == "" {
		cfg.Git.AuthorName = "open-nucleus-sync"
	}
	if cfg.Git.AuthorEmail == "" {
		cfg.Git.AuthorEmail = "sync@open-nucleus.local"
	}
	if cfg.Sync.MaxConcurrent == 0 {
		cfg.Sync.MaxConcurrent = 1
	}
	if cfg.Sync.Cooldown == 0 {
		cfg.Sync.Cooldown = 30 * time.Second
	}
	if cfg.Sync.HandshakeTimeout == 0 {
		cfg.Sync.HandshakeTimeout = 10 * time.Second
	}
	if cfg.Sync.TransferTimeout == 0 {
		cfg.Sync.TransferTimeout = 300 * time.Second
	}
	if cfg.Sync.ChunkSize == 0 {
		cfg.Sync.ChunkSize = 65536
	}
	if cfg.Discovery.ScanInterval == 0 {
		cfg.Discovery.ScanInterval = 30 * time.Second
	}
	if cfg.Discovery.PeerTTL == 0 {
		cfg.Discovery.PeerTTL = 300 * time.Second
	}
	if cfg.History.DBPath == "" {
		cfg.History.DBPath = "/var/lib/open-nucleus/sync.db"
	}
	if cfg.History.MaxEntries == 0 {
		cfg.History.MaxEntries = 10000
	}
	if cfg.Events.BufferSize == 0 {
		cfg.Events.BufferSize = 100
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}
}
