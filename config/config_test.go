package config

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var loadFileTests = map[string]struct {
	want *Config
}{
	"testdata/test01.toml": {
		want: &Config{
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
	yaml string
	want *Config
}{
	"test01": {
		yaml: `dictionaries = [
  {path = "dict1"},
  {path = "dict2", encoding = "utf-8"},
  {path = "dict3", encoding = "shift_jis"},
  {path = "dict4", encoding = "iso-2022-jp"},
  {path = "dict5", encoding = "euc-jp"},
]`,
		want: &Config{
			Dictionaries: []*Dictionary{
				{Path: "dict1", Encoding: Auto},
				{Path: "dict2", Encoding: UTF8},
				{Path: "dict3", Encoding: ShiftJIS},
				{Path: "dict4", Encoding: ISO2022JP},
				{Path: "dict5", Encoding: EUCJP},
			},
		},
	},
	"test02": {
		yaml: "dictionaries = []",
		want: &Config{
			Dictionaries: []*Dictionary{},
		},
	},
	"test03": {
		yaml: "",
		want: &Config{
			Dictionaries: nil,
		},
	},
}

func TestLoad(t *testing.T) {
	for name, tt := range loadTests {
		t.Run(name, func(t *testing.T) {
			cfg, err := Load(strings.NewReader(tt.yaml))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tt.want, cfg); diff != "" {
				t.Errorf("Load(%q) mismatch (-want +got):\n%s", tt.yaml, diff)
			}
		})
	}
}
