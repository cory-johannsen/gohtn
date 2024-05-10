package loader

import (
	"encoding/json"
	"fmt"
	"github.com/cory-johannsen/gohtn/config"
	"github.com/cory-johannsen/gohtn/gohtn"
	"os"
	"path/filepath"
)

func LoadSensors(cfg *config.Config) ([]gohtn.Sensor, error) {
	sensors := make([]gohtn.Sensor, 0)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		sensor, err := loadSensor(path)
		if err != nil {
			return err
		}
		sensors = append(sensors, sensor)
		return nil
	}
	sensorPath := fmt.Sprintf("%s/%s", cfg.AssetRoot, cfg.SensorPath)
	err := filepath.Walk(sensorPath, walkFn)
	if err != nil {
		return nil, fmt.Errorf("error walking the path %q: %v", sensorPath, err)
	}
	return sensors, nil
}

func loadSensor(path string) (gohtn.Sensor, error) {
	sensor := &gohtn.SimpleSensor{}
	buffer, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buffer, sensor)
	if err != nil {
		return nil, err
	}
	return sensor, nil
}
