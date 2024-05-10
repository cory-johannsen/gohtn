package gohtn

import "fmt"

// Sensor are represented by a generic 64 bit floating point Value.
type Sensor interface {
	Get() (float64, error)
	String() string
	Name() string
}

// SimpleSensor stores a single float64 Value and allows it to be set
type SimpleSensor struct {
	Value      float64 `json:"value"`
	SensorName string  `json:"name"`
}

func NewSimpleSensor(name string, value float64) *SimpleSensor {
	return &SimpleSensor{SensorName: name, Value: value}
}

func (s *SimpleSensor) Get() (float64, error) {
	return s.Value, nil
}

func (s *SimpleSensor) Name() string {
	return s.SensorName
}

func (s *SimpleSensor) Set(value float64) {
	s.Value = value
}

func (s *SimpleSensor) String() string {
	return fmt.Sprintf("%s: %f", s.SensorName, s.Value)
}
