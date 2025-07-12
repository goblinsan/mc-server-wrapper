package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	ServerDir       string `json:"server_dir"`
	BackupDir       string `json:"backup_dir"`
	WorldDir        string `json:"world_dir"`
	NetworkShare    string `json:"network_share"`
	WikiNavURL      string `json:"wiki_nav_url"`
	LastVersionFile string `json:"last_version_file"`
}

// LoadConfig loads the config from the given JSON file path
func LoadConfig(path string) (Config, error) {
	var cfg Config
	f, err := os.Open(path)
	if err != nil {
		return cfg, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("failed to decode config: %w", err)
	}
	return cfg, nil
}
