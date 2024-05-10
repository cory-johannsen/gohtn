package loader

import (
	"encoding/json"
	"github.com/cory-johannsen/gohtn/config"
	"os"
)

func LoadConfig(configFile string) (*config.Config, error) {
	cfg := &config.Config{}
	buffer, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buffer, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
