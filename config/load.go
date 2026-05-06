package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads YAML from path, expands ${VAR} references against the environment,
// and validates the result.
func Load(path string) (Config, error) {
	cfg := defaults()

	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %q: %w", path, err)
	}

	// Expand ${VAR} / $VAR before parsing so secrets can live in env.
	expanded := os.ExpandEnv(string(raw))

	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %q: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	var errs []error
	if c.Server.Address == "" {
		errs = append(errs, errors.New("server.address is required"))
	}
	if c.Database.Path == "" {
		errs = append(errs, errors.New("database.path is required"))
	}
	if c.Auth.JWTSecret == "" {
		errs = append(errs, errors.New("auth.jwt_secret is required (set ${JWT_SECRET} in env)"))
	}
	if c.Auth.AccessTokenTTL <= 0 {
		errs = append(errs, errors.New("auth.access_token_ttl must be positive"))
	}
	if c.Auth.RefreshTokenTTL <= 0 {
		errs = append(errs, errors.New("auth.refresh_token_ttl must be positive"))
	}
	return errors.Join(errs...)
}
