package config

import (
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	GRPCPort int          `koanf:"grpc_port"`
	SQLite   SQLiteConfig `koanf:"sqlite"`
	DrugDB   DrugDBConfig `koanf:"drug_db"`
	Logging  LoggingConfig `koanf:"logging"`
}

type SQLiteConfig struct {
	DBPath string `koanf:"db_path"`
}

type DrugDBConfig struct {
	MedicationsDir  string `koanf:"medications_dir"`
	InteractionsFile string `koanf:"interactions_file"`
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
	if err := k.Unmarshal("formulary_service", &cfg); err != nil {
		return nil, err
	}

	setDefaults(&cfg)
	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.GRPCPort == 0 {
		cfg.GRPCPort = 50054
	}
	if cfg.SQLite.DBPath == "" {
		cfg.SQLite.DBPath = "/var/lib/open-nucleus/formulary.db"
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}
}
