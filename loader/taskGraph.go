package loader

import (
	"encoding/json"
	"fmt"
	"github.com/cory-johannsen/gohtn/config"
	"github.com/cory-johannsen/gohtn/gohtn"
	"os"
)

func LoadTaskGraph(cfg *config.Config) (*gohtn.TaskGraph, error) {
	taskGraphPath := fmt.Sprintf("%s/%s", cfg.AssetRoot, cfg.TaskGraphPath)
	taskGraph := &gohtn.TaskGraph{}
	buffer, err := os.ReadFile(taskGraphPath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buffer, taskGraph)
	if err != nil {
		return nil, err
	}
	return taskGraph, nil
}
