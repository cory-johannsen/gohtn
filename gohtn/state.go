package gohtn

import (
	"fmt"
	"strings"
)

type Value[T any] func(state *State) T

// Property is a named function that accepts the state and returns a generic typed value.
type Property[T any] struct {
	Name  string
	Value Value[T]
}

// State is represented as an array of Sensors and a map of named Properties.
type State struct {
	Sensors    map[string]any
	Properties map[string]any
}

func (s *State) Property(name string) (any, error) {
	property, ok := s.Properties[name]
	if !ok {
		return 0, fmt.Errorf("no Property with name %s", name)
	}
	return property, nil
}

func (s *State) Sensor(name string) (any, error) {
	sensor, ok := s.Sensors[name]
	if !ok {
		return nil, fmt.Errorf("no sensor with name %s", name)
	}
	return sensor, nil
}

func (s *State) String() string {
	sensors := make([]string, 0)
	for sensor := range s.Sensors {
		sensors = append(sensors, fmt.Sprintf("{%s}", sensor))
	}
	properties := make([]string, 0)
	for k := range s.Properties {
		properties = append(properties, fmt.Sprintf("%s", k))
	}
	return fmt.Sprintf("sensors: %s, properties: %s", strings.Join(sensors, ","), strings.Join(properties, ","))
}
