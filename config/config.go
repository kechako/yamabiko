// Package config provides functions to load and manage configuration settings
package config

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/kechako/skkdic"
)

func defaultConfig() *Config {
	return &Config{
		Host:           "",
		Port:           1178,
		MaxCompletions: 10,
		Logging: &Logging{
			Path:  "", // Default to stdout
			Level: slog.LevelInfo,
			JSON:  false,
		},
		Dictionaries: nil,
	}
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

type Config struct {
	Host           string        `toml:"host"`
	Port           int           `toml:"port"`
	SendEncoding   Encoding      `toml:"send_encoding"`
	RecvEncoding   Encoding      `toml:"recv_encoding"`
	MaxCompletions int           `toml:"max_completions"`
	Logging        *Logging      `toml:"logging"`
	Dictionaries   []*Dictionary `toml:"dictionaries"`
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
	var buf bytes.Buffer
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Text()
		line = os.ExpandEnv(line)
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("config.load: failed to read config: %w", err)
	}

	cfg := defaultConfig()
	_, err := toml.NewDecoder(&buf).Decode(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func FindConfigFile() (string, error) {
	paths, err := getConfigFilePaths()
	if err != nil {
		return "", fmt.Errorf("config.FindConfigFile: %w", err)
	}

	for _, path := range paths {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, err
		}
	}

	return "", nil
}

func getConfigFilePaths() ([]string, error) {
	paths := []string{
		"/etc/yamabiko/config.toml",
		"/etc/yamabiko.toml",
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	paths = append(paths, filepath.Join(home, ".config/yamabiko/config.toml"))
	paths = append(paths, filepath.Join(home, ".yamabiko.toml"))

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	paths = append(paths, filepath.Join(wd, "config.toml"))

	return paths, nil
}
