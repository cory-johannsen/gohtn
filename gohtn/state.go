package gohtn

import (
	"fmt"
	"strings"
)

// Property is a function that accepts the state and returns a float64.
type Property func(state *State) float64

// State is represented as an array of Sensors and a map of named Properties.
type State struct {
	Sensors    []Sensor            `json:"sensors"`
	Properties map[string]Property `json:"properties"`
}

func NewState(sensors []Sensor, properties map[string]Property) *State {
	return &State{Sensors: sensors, Properties: properties}
}

func (s *State) Property(name string) (float64, error) {
	property, ok := s.Properties[name]
	if !ok {
		return 0, fmt.Errorf("no Property with name %s", name)
	}
	return property(s), nil
}

func (s *State) Sensor(name string) (Sensor, error) {
	for _, sensor := range s.Sensors {
		if sensor.Name() == name {
			return sensor, nil
		}
	}
	return nil, fmt.Errorf("no sensor with name %s", name)
}

func (s *State) String() string {
	sensors := make([]string, 0)
	for i := range s.Sensors {
		sensors = append(sensors, fmt.Sprintf("{%s}", s.Sensors[i].String()))
	}
	properties := make([]string, 0)
	for k, v := range s.Properties {
		properties = append(properties, fmt.Sprintf("%s=%v", k, v(s)))
	}
	return fmt.Sprintf("sensors: %s, properties: %s", strings.Join(sensors, ","), strings.Join(properties, ","))
}
