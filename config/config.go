package config

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/kechako/skkdic"
)

var defaultConfig = &Config{
	Host: "",
	Port: 1178,
	Logging: &Logging{
		Path:  "", // Default to stdout
		Level: slog.LevelInfo,
		JSON:  false,
	},
}

type Encoding = skkdic.Encoding

const (
	Auto      = skkdic.Auto
	Default   = skkdic.Auto
	EUCJP     = skkdic.EUCJP
	ShiftJIS  = skkdic.ShiftJIS
	ISO2022JP = skkdic.ISO2022JP
	UTF8      = skkdic.UTF8
)

type Dictionary struct {
	Path     string   `toml:"path"`
	Encoding Encoding `toml:"encoding"`
}

type Logging struct {
	Path  string     `toml:"path"`
	Level slog.Level `toml:"level"`
	JSON  bool       `toml:"json"`
}

func (l *Logging) merge(other *Logging) {
	if other == nil {
		return
	}

	if other.Path != "" {
		l.Path = other.Path
	}
	if other.Level < l.Level {
		l.Level = other.Level
	}
	if other.JSON {
		l.JSON = other.JSON
	}
}

type Config struct {
	Host         string        `toml:"host"`
	Port         int           `toml:"port"`
	SendEncoding Encoding      `toml:"send_encoding"`
	RecvEncoding Encoding      `toml:"recv_encoding"`
	Dictionaries []*Dictionary `toml:"dictionaries"`
	Logging      *Logging      `toml:"logging"`
}

func (cfg *Config) merge(other *Config) {
	if other == nil {
		return
	}

	if other.Host != "" {
		cfg.Host = other.Host
	}
	if other.Port > 0 {
		cfg.Port = other.Port
	}
	if other.SendEncoding.IsValid() && other.SendEncoding != Default {
		cfg.SendEncoding = other.SendEncoding
	}
	if other.RecvEncoding.IsValid() && other.RecvEncoding != Default {
		cfg.RecvEncoding = other.RecvEncoding
	}

	if len(other.Dictionaries) > 0 {
		cfg.Dictionaries = append(cfg.Dictionaries, other.Dictionaries...)
	}

	if cfg.Logging == nil {
		cfg.Logging = other.Logging
	} else {
		cfg.Logging.merge(other.Logging)
	}
}

func LoadFile(path string) (*Config, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config.LoadFile: failed to open config file: %w", err)
	}
	defer r.Close()
	cfg, err := load(r)
	if err != nil {
		return nil, fmt.Errorf("config.LoadFile: %w", err)
	}
	return cfg, nil
}

func Load(r io.Reader) (*Config, error) {
	cfg, err := load(r)
	if err != nil {
		return nil, fmt.Errorf("config.Load: %w", err)
	}
	return cfg, nil
}

func load(r io.Reader) (*Config, error) {
	var cfg Config
	_, err := toml.NewDecoder(r).Decode(&cfg)
	if err != nil {
		return nil, err
	}

	cfg.merge(defaultConfig)

	return &cfg, nil
}
