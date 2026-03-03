package config

import (
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Server    ServerConfig    `koanf:"server"`
	Auth      AuthConfig      `koanf:"auth"`
	GRPC      GRPCConfig      `koanf:"grpc"`
	RateLimit RateLimitConfig `koanf:"rate_limit"`
	CORS      CORSConfig      `koanf:"cors"`
	WebSocket WebSocketConfig `koanf:"websocket"`
	Logging   LoggingConfig   `koanf:"logging"`
	Smart     SmartConfig     `koanf:"smart"`
}

type SmartConfig struct {
	Enabled bool   `koanf:"enabled"`
	BaseURL string `koanf:"base_url"`
}

type ServerConfig struct {
	Port           int           `koanf:"port"`
	ReadTimeout    time.Duration `koanf:"read_timeout"`
	WriteTimeout   time.Duration `koanf:"write_timeout"`
	MaxRequestBody string        `koanf:"max_request_body"`
}

type AuthConfig struct {
	JWTIssuer     string        `koanf:"jwt_issuer"`
	TokenLifetime time.Duration `koanf:"token_lifetime"`
	RefreshWindow time.Duration `koanf:"refresh_window"`
}

type GRPCConfig struct {
	PatientService   string        `koanf:"patient_service"`
	SyncService      string        `koanf:"sync_service"`
	AuthService      string        `koanf:"auth_service"`
	FormularyService string        `koanf:"formulary_service"`
	AnchorService    string        `koanf:"anchor_service"`
	SentinelAgent    string        `koanf:"sentinel_agent"`
	DialTimeout      time.Duration `koanf:"dial_timeout"`
	RequestTimeout   time.Duration `koanf:"request_timeout"`
}

type RateLimitConfig struct {
	ReadRPM    int `koanf:"read_rpm"`
	ReadBurst  int `koanf:"read_burst"`
	WriteRPM   int `koanf:"write_rpm"`
	WriteBurst int `koanf:"write_burst"`
	AuthRPM    int `koanf:"auth_rpm"`
	AuthBurst  int `koanf:"auth_burst"`
}

type CORSConfig struct {
	AllowedOrigins []string `koanf:"allowed_origins"`
}

type WebSocketConfig struct {
	PingInterval   time.Duration `koanf:"ping_interval"`
	MaxConnections int           `koanf:"max_connections"`
}

type LoggingConfig struct {
	Level     string `koanf:"level"`
	AuditFile string `koanf:"audit_file"`
	Format    string `koanf:"format"`
}

func Load(path string) (*Config, error) {
	k := koanf.New(".")

	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return nil, err
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	setDefaults(&cfg)
	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 30 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 30 * time.Second
	}
	if cfg.Auth.JWTIssuer == "" {
		cfg.Auth.JWTIssuer = "open-nucleus-auth"
	}
	if cfg.Auth.TokenLifetime == 0 {
		cfg.Auth.TokenLifetime = 24 * time.Hour
	}
	if cfg.Auth.RefreshWindow == 0 {
		cfg.Auth.RefreshWindow = 2 * time.Hour
	}
	if cfg.GRPC.DialTimeout == 0 {
		cfg.GRPC.DialTimeout = 5 * time.Second
	}
	if cfg.GRPC.RequestTimeout == 0 {
		cfg.GRPC.RequestTimeout = 30 * time.Second
	}
	if cfg.RateLimit.ReadRPM == 0 {
		cfg.RateLimit.ReadRPM = 200
	}
	if cfg.RateLimit.ReadBurst == 0 {
		cfg.RateLimit.ReadBurst = 50
	}
	if cfg.RateLimit.WriteRPM == 0 {
		cfg.RateLimit.WriteRPM = 60
	}
	if cfg.RateLimit.WriteBurst == 0 {
		cfg.RateLimit.WriteBurst = 20
	}
	if cfg.RateLimit.AuthRPM == 0 {
		cfg.RateLimit.AuthRPM = 10
	}
	if cfg.RateLimit.AuthBurst == 0 {
		cfg.RateLimit.AuthBurst = 5
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}
}
