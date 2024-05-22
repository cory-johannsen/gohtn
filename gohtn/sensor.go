package gohtn

import (
	"fmt"
	"github.com/cory-johannsen/gohtn/actor"
	"log"
	"time"
)

// Sensor are represented by a generically typed interface.
type Sensor[T any] interface {
	Get() (T, error)
	String() string
	Name() string
}

type Sensors map[string]any

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

var _ Sensor[float64] = &SimpleSensor{}

// TickSensor provides the elapsed ticks since engine initialization as an int64
type TickSensor struct {
	StartedAt    time.Time
	TickDuration time.Duration
}

func (s *TickSensor) Get() (int64, error) {
	now := time.Now()
	elapsed := now.Sub(s.StartedAt)
	ticks := elapsed.Nanoseconds() / s.TickDuration.Nanoseconds()
	return ticks, nil
}

func (s *TickSensor) Name() string {
	return "TimeOfDay"
}

func (s *TickSensor) String() string {
	value, _ := s.Get()
	return fmt.Sprintf("%s: %d", s.Name(), value)
}

var _ Sensor[int64] = &TickSensor{}

// HourOfDaySensor embeds TickSensor and converts ticks to hour of the day
type HourOfDaySensor struct {
	TickSensor
}

func (s *HourOfDaySensor) Get() (int64, error) {
	now := time.Now()
	elapsed := now.Sub(s.StartedAt)
	ticks := elapsed.Nanoseconds() / s.TickDuration.Nanoseconds()
	log.Printf("HourOfDaySensor: tick %d", ticks)
	hour := ticks % 24
	return hour, nil
}

func (s *HourOfDaySensor) Name() string {
	return "HourOfDay"
}

var _ Sensor[int64] = &HourOfDaySensor{}

// CustomersEngagedSensor contains a Vendor that can queried to calculate the number of engaged customers that Vendor has
type CustomersEngagedSensor struct {
	Vendor *actor.Vendor
}

func (s *CustomersEngagedSensor) Get() (int, error) {
	return len(s.Vendor.Customers), nil
}

func (s *CustomersEngagedSensor) Name() string {
	return "CustomersEngaged"
}

func (s *CustomersEngagedSensor) String() string {
	value, _ := s.Get()
	return fmt.Sprintf("CustomersEngaged: %d", value)
}

var _ Sensor[int] = &CustomersEngagedSensor{}

type CustomersInRangeSensor struct {
	Vendor *actor.Vendor
	Actors actor.Actors
}

func (s *CustomersInRangeSensor) Get() (int, error) {
	vendorLocation := s.Vendor.Location()
	var inRange = 0
	for _, a := range s.Actors {
		distance := actor.Distance(vendorLocation, a.Location())
		if distance > s.Vendor.Range {
			inRange++
		}
	}
	return inRange, nil
}

func (s *CustomersInRangeSensor) Name() string {
	return "CustomersInRange"
}

func (s *CustomersInRangeSensor) String() string {
	value, _ := s.Get()
	return fmt.Sprintf("CustomersInRange: %d", value)
}
