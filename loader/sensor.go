package loader

import (
	"encoding/json"
	"fmt"
	"github.com/cory-johannsen/gohtn/config"
	"github.com/cory-johannsen/gohtn/gohtn"
	"os"
	"path/filepath"
	"strings"
)

type SensorType string

const (
	Simple SensorType = "simple"
)

func LoadSensors(cfg *config.Config) ([]gohtn.Sensor, error) {
	sensorPath := fmt.Sprintf("%s/%s", cfg.AssetRoot, cfg.SensorPath)
	sensors := make([]gohtn.Sensor, 0)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		sensorSubpath := strings.TrimPrefix(path, fmt.Sprintf("%s/", sensorPath))
		pathComponents := strings.Split(sensorSubpath, "/")
		sensorType := SensorType(pathComponents[0])
		sensor, err := loadSensor(sensorType, path)
		if err != nil {
			return err
		}
		sensors = append(sensors, sensor)
		return nil
	}
	err := filepath.Walk(sensorPath, walkFn)
	if err != nil {
		return nil, fmt.Errorf("error walking the path %q: %v", sensorPath, err)
	}
	return sensors, nil
}

func initSensor(sensorType SensorType) (gohtn.Sensor, error) {
	switch sensorType {
	case Simple:
		return &gohtn.SimpleSensor{}, nil
	}
	return nil, fmt.Errorf("invalid sensor type")
}

func loadSensor(sensorType SensorType, path string) (gohtn.Sensor, error) {
	sensor, err := initSensor(sensorType)
	if err != nil {
		return nil, err
	}
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
