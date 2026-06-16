//go:build windows

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const DefaultFileName = "tickerfile.json"

type Config struct {
	Service ServiceConfig `json:"service"`
	Log     LogConfig     `json:"log"`
	Ticker  TickerConfig  `json:"ticker"`
}

type ServiceConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type LogConfig struct {
	Dir       string `json:"dir"`
	TextFile  string `json:"textFile"`
	Win32File string `json:"win32File"`
}

type TickerConfig struct {
	IntervalSeconds int `json:"intervalSeconds"`
}

func Default(exeDir string) Config {
	return Config{
		Service: ServiceConfig{
			Name:        "tickerfile",
			Description: "Writes timestamps to log files every 2 seconds",
		},
		Log: LogConfig{
			Dir:       exeDir,
			TextFile:  "text.log",
			Win32File: "win32.log",
		},
		Ticker: TickerConfig{
			IntervalSeconds: 2,
		},
	}
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %q: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %q: %w", path, err)
	}
	return cfg, nil
}

func (c *Config) ApplyDefaults(exeDir string) {
	def := Default(exeDir)
	if c.Service.Name == "" {
		c.Service.Name = def.Service.Name
	}
	if c.Service.Description == "" {
		c.Service.Description = def.Service.Description
	}
	if c.Log.Dir == "" {
		c.Log.Dir = def.Log.Dir
	}
	if c.Log.TextFile == "" {
		c.Log.TextFile = def.Log.TextFile
	}
	if c.Log.Win32File == "" {
		c.Log.Win32File = def.Log.Win32File
	}
	if c.Ticker.IntervalSeconds <= 0 {
		c.Ticker.IntervalSeconds = def.Ticker.IntervalSeconds
	}
}

func (c Config) Interval() time.Duration {
	return time.Duration(c.Ticker.IntervalSeconds) * time.Second
}

func ResolvePath(exePath, configPath string) string {
	if configPath != "" {
		return configPath
	}
	return filepath.Join(filepath.Dir(exePath), DefaultFileName)
}

func WriteDefault(path string, exeDir string) error {
	cfg := Default(exeDir)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config %q: %w", path, err)
	}
	return nil
}
