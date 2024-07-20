package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	RepoURL       string `json:"repo_url"`
	DefaultBranch string `json:"default_branch"`
	TerraformPath string `json:"terraform_path"`
	LogFile       string `json:"log_file"`
}

var (
	// DefaultConfig holds the default configuration
	DefaultConfig = Config{
		RepoURL:       "https://github.com/jlgore/nsfw-infra",
		DefaultBranch: "main",
		TerraformPath: "terraform",
		LogFile:       "nsfwctl.log",
	}

	// CurrentConfig holds the current active configuration
	CurrentConfig Config
)

// LoadConfig loads the configuration from a file
func LoadConfig(configPath string) error {
	// Start with default config
	CurrentConfig = DefaultConfig

	// If config file exists, load it
	if _, err := os.Stat(configPath); err == nil {
		file, err := os.Open(configPath)
		if err != nil {
			return fmt.Errorf("error opening config file: %v", err)
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&CurrentConfig); err != nil {
			return fmt.Errorf("error decoding config file: %v", err)
		}
	}

	return nil
}

// SaveConfig saves the current configuration to a file
func SaveConfig(configPath string) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %v", err)
	}

	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("error creating config file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(CurrentConfig); err != nil {
		return fmt.Errorf("error encoding config file: %v", err)
	}

	return nil
}

// GetConfigFilePath returns the path to the config file
func GetConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %v", err)
	}

	return filepath.Join(homeDir, ".nsfwctl", "config.json"), nil
}

// Init initializes the configuration
func Init() error {
	configPath, err := GetConfigFilePath()
	if err != nil {
		return err
	}

	if err := LoadConfig(configPath); err != nil {
		return err
	}

	// If config file doesn't exist, create it with default values
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := SaveConfig(configPath); err != nil {
			return err
		}
	}

	return nil
}
