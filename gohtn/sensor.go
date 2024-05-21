package gohtn

import (
	"fmt"
	"github.com/cory-johannsen/gohtn/actor"
	"time"
)

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

type TickSensor struct {
	StartedAt    time.Time
	TickDuration time.Duration
}

func (s *TickSensor) Get() (float64, error) {
	now := time.Now()
	elapsed := now.Sub(s.StartedAt)
	ticks := elapsed.Nanoseconds() / s.TickDuration.Nanoseconds()
	return float64(ticks), nil
}

func (s *TickSensor) Name() string {
	return "TimeOfDay"
}

func (s *TickSensor) String() string {
	value, _ := s.Get()
	return fmt.Sprintf("%s: %f", s.Name(), value)
}

type HourOfDaySensor struct {
	TickSensor
}

func (s *HourOfDaySensor) Get() (float64, error) {
	now := time.Now()
	elapsed := now.Sub(s.StartedAt)
	ticks := elapsed.Nanoseconds() / s.TickDuration.Nanoseconds()
	hour := ticks % 24
	return float64(hour), nil
}

func (s *HourOfDaySensor) Name() string {
	return "HourOfDay"
}

type CustomersEngagedSensor struct {
	Vendor *actor.Vendor
}

func (s *CustomersEngagedSensor) Get() (float64, error) {
	return 0, nil
}

func (s *CustomersEngagedSensor) Name() string {
	return "CustomersEngaged"
}

func (s *CustomersEngagedSensor) String() string {
	value, _ := s.Get()
	return fmt.Sprintf("CustomersEngaged: %d", int64(value))
}
