package engine

import (
	"github.com/cory-johannsen/gohtn/actor"
	"github.com/cory-johannsen/gohtn/gohtn"
)

type Actions map[string]gohtn.Action
type Conditions map[string]gohtn.Condition

type Methods map[string]*gohtn.Method

type Engine struct {
	Actors        actor.Actors
	Sensors       gohtn.Sensors
	Actions       Actions
	Conditions    Conditions
	TaskResolvers gohtn.TaskResolvers
	Tasks         gohtn.Tasks
	Methods       Methods
	Planner       *gohtn.Planner
	Domain        *gohtn.TaskGraph
}
