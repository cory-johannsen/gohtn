package engine

import (
	"github.com/cory-johannsen/gohtn/gohtn"
)

type Sensors map[string]gohtn.Sensor
type Actions map[string]gohtn.Action
type Conditions map[string]gohtn.Condition

type Methods map[string]*gohtn.Method

type Engine struct {
	Sensors       Sensors
	Actions       Actions
	Conditions    Conditions
	TaskResolvers gohtn.TaskResolvers
	Tasks         gohtn.Tasks
	Methods       Methods
	Planner       *gohtn.Planner
	Domain        *gohtn.TaskGraph
}
