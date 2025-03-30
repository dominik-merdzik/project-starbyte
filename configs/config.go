package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Music MusicConfig `toml:"music"`
}

type MusicConfig struct {
	Enabled bool `toml:"enabled"`
	Volume  int  `toml:"volume"`
}

const defaultConfigContent = `[music] # Background music configuration
# Set to false to disable background music
enabled = true
# Set the volume level from 0 to 100
volume = 25 

[keybindings] # Keybindings for the game
vim_mode = false # Set to true to enable vim keybindings
`

// InitConfig ensures that the config directory and file exist, writing the default config if missing, and then loads the config file
func InitConfig(configPath string) (Config, error) {

	// ensure the directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return Config{}, fmt.Errorf("error creating config directory: %w", err)
	}

	// if the file doesn't exist, write the default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.WriteFile(configPath, []byte(defaultConfigContent), 0644); err != nil {
			return Config{}, fmt.Errorf("error writing default config file: %w", err)
		}
	}

	// load the config file
	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding config file: %w", err)
	}
	return cfg, nil
}
