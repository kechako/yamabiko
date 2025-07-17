package config

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var loadFileTests = map[string]struct {
	want *Config
}{
	"testdata/test01.toml": {
		want: &Config{
			Host: "",
			Port: 1178,
			Logging: &Logging{
				Path:  "",
				Level: slog.LevelInfo,
				JSON:  false,
			},
			Dictionaries: nil,
		},
	},
	"testdata/test02.toml": {
		want: &Config{
			Host:         "0.0.0.0",
			Port:         12345,
			SendEncoding: EUCJP,
			RecvEncoding: ShiftJIS,
			Logging: &Logging{
				Path:  "/var/log/yamabiko-test.log",
				Level: slog.LevelDebug,
				JSON:  true,
			},
			Dictionaries: []*Dictionary{
				{Path: "dict1", Encoding: Auto},
				{Path: "dict2", Encoding: UTF8},
				{Path: "dict3", Encoding: ShiftJIS},
				{Path: "dict4", Encoding: ISO2022JP},
				{Path: "dict5", Encoding: EUCJP},
			},
		},
	},
}

func TestLoadFile(t *testing.T) {
	for name, tt := range loadFileTests {
		t.Run(name, func(t *testing.T) {
			cfg, err := LoadFile(name)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tt.want, cfg); diff != "" {
				t.Errorf("LoadFile(%q) mismatch (-want +got):\n%s", name, diff)
			}
		})
	}
}

var loadTests = map[string]struct {
	toml string
	want *Config
}{
	"test01": {
		toml: "",
		want: &Config{
			Host: "",
			Port: 1178,
			Logging: &Logging{
				Path:  "",
				Level: slog.LevelInfo,
				JSON:  false,
			},
			Dictionaries: nil,
		},
	},
	"test02": {
		toml: `host = "0.0.0.0"
port = 12345
send_encoding = "euc-jp"
recv_encoding = "shift_jis"
logging = {path = "/var/log/yamabiko-test.log", level = "debug", json = true}
dictionaries = [
  {path = "dict1"},
  {path = "dict2", encoding = "utf-8"},
  {path = "dict3", encoding = "shift_jis"},
  {path = "dict4", encoding = "iso-2022-jp"},
  {path = "dict5", encoding = "euc-jp"},
]`,
		want: &Config{
			Host:         "0.0.0.0",
			Port:         12345,
			SendEncoding: EUCJP,
			RecvEncoding: ShiftJIS,
			Logging: &Logging{
				Path:  "/var/log/yamabiko-test.log",
				Level: slog.LevelDebug,
				JSON:  true,
			},
			Dictionaries: []*Dictionary{
				{Path: "dict1", Encoding: Auto},
				{Path: "dict2", Encoding: UTF8},
				{Path: "dict3", Encoding: ShiftJIS},
				{Path: "dict4", Encoding: ISO2022JP},
				{Path: "dict5", Encoding: EUCJP},
			},
		},
	},
	"test03": {
		toml: "dictionaries = []",
		want: &Config{
			Host: "",
			Port: 1178,
			Logging: &Logging{
				Path:  "",
				Level: slog.LevelInfo,
				JSON:  false,
			},
			Dictionaries: []*Dictionary{},
		},
	},
}

func TestLoad(t *testing.T) {
	for name, tt := range loadTests {
		t.Run(name, func(t *testing.T) {
			cfg, err := Load(strings.NewReader(tt.toml))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tt.want, cfg); diff != "" {
				t.Errorf("Load(%q) mismatch (-want +got):\n%s", tt.toml, diff)
			}
		})
	}
}
