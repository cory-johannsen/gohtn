package gohtn

import (
	"fmt"
	"strings"
)

// Sensor are represented by a generic 64 bit floating point Value.
type Sensor interface {
	Value() (float64, error)
	String() string
}

// SimpleSensor stores a single float64 Value and allows it to be set
type SimpleSensor struct {
	value float64
}

func NewSimpleSensor(value float64) *SimpleSensor {
	return &SimpleSensor{value}
}

func (s *SimpleSensor) Value() (float64, error) {
	return s.value, nil
}

func (s *SimpleSensor) Set(value float64) {
	s.value = value
}

func (s *SimpleSensor) String() string {
	return fmt.Sprintf("%f", s.value)
}

// State is represented as an array of Sensors.
// For simplicity each Property in the state currently corresponds to exactly one sensor.
type State struct {
	sensors    []Sensor
	properties map[string]Sensor
}

func NewState(sensors []Sensor, properties map[string]Sensor) *State {
	return &State{sensors: sensors, properties: properties}
}

func (s *State) Property(name string) (float64, error) {
	sensor, ok := s.properties[name]
	if !ok {
		return 0, fmt.Errorf("no Property with name %s", name)
	}
	return sensor.Value()
}

func (s *State) Sensor(name string) (Sensor, error) {
	sensor, ok := s.properties[name]
	if !ok {
		return nil, fmt.Errorf("no Property with name %s", name)
	}
	return sensor, nil
}

func (s *State) String() string {
	sensors := make([]string, 0)
	for i := range s.sensors {
		sensors = append(sensors, s.sensors[i].String())
	}
	properties := make([]string, 0)
	for k, v := range s.properties {
		properties = append(properties, fmt.Sprintf("%s=%s", k, v))
	}
	return fmt.Sprintf("sensors: %s, properties: %s", strings.Join(sensors, ","), strings.Join(properties, ","))
}
