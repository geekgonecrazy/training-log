package config

import "time"

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	Log      LogConfig      `yaml:"log"`
}

type ServerConfig struct {
	Address         string        `yaml:"address"`
	InternalAddress string        `yaml:"internal_address"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type AuthConfig struct {
	JWTSecret        string        `yaml:"jwt_secret"`
	AccessTokenTTL   time.Duration `yaml:"access_token_ttl"`
	RefreshTokenTTL  time.Duration `yaml:"refresh_token_ttl"`
	RegistrationOpen bool          `yaml:"registration_open"`
	CookieDomain     string        `yaml:"cookie_domain"`
	CookieSecure     bool          `yaml:"cookie_secure"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func defaults() Config {
	return Config{
		Server: ServerConfig{
			Address:         ":8080",
			InternalAddress: ":9090",
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    30 * time.Second,
			ShutdownTimeout: 10 * time.Second,
		},
		Database: DatabaseConfig{
			Path: "./data/habit.db",
		},
		Auth: AuthConfig{
			AccessTokenTTL:   15 * time.Minute,
			RefreshTokenTTL:  720 * time.Hour,
			RegistrationOpen: false,
			CookieSecure:     true,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
		},
	}
}
