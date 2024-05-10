package gohtn

import (
	"fmt"
	"log"
	"strings"
)

type Task interface {
	Execute(state *State) (*State, error)
	IsComplete() bool
	Name() string
	String() string
}

type Condition interface {
	IsMet(state *State) bool
	String() string
}

// FlagCondition is a simple condition that is gated by a boolean Value that can be set
type FlagCondition struct {
	Value bool
}

func (f *FlagCondition) IsMet(state *State) bool {
	return f.Value
}

func (f *FlagCondition) Set(value bool) {
	f.Value = value
}

func (f *FlagCondition) String() string {
	return fmt.Sprintf("FlagCondition: %t", f.Value)
}

// GTECondition is a condition that is met if the given Property is GTE the specified Value
type GTECondition struct {
	Value    float64
	Property string
}

func (g *GTECondition) IsMet(state *State) bool {
	value, err := state.Property(g.Property)
	if err != nil {
		return false
	}
	return g.Value >= value
}

func (g *GTECondition) String() string {
	return fmt.Sprintf("GTECondition: property %s, value %f", g.Property, g.Value)
}

// TaskCondition is a condition that is met when the given Task is complete
type TaskCondition struct {
	Task Task
}

func (t *TaskCondition) IsMet(state *State) bool {
	return t.Task.IsComplete()
}

func (t *TaskCondition) String() string {
	return fmt.Sprintf("TaskCondition: %s, complete: %t", t.Task.Name(), t.Task.IsComplete())
}

// Action is an action applied by a Task.
type Action func(state *State) error

// PrimitiveTask implements the HTN primitive Task.   It contains a set of preconditions that must be met
// before it will execute.  Once the preconditions are met, the Action is applied, then the completion flag is set.
type PrimitiveTask struct {
	preconditions []Condition
	complete      bool
	action        Action
	name          string
}

func NewPrimitiveTask(name string, preconditions []Condition, action Action) *PrimitiveTask {
	return &PrimitiveTask{
		preconditions: preconditions,
		action:        action,
		complete:      false,
		name:          name,
	}
}

func (t *PrimitiveTask) Execute(state *State) (*State, error) {
	preconditions := make([]string, 0)
	for _, condition := range t.preconditions {
		preconditions = append(preconditions, condition.String())
	}
	log.Printf("executing Task %s, preconditions %s", t.name, strings.Join(preconditions, ","))
	// Determine if the Task preconditions have been met
	var ready = true
	for _, condition := range t.preconditions {
		if !condition.IsMet(state) {
			ready = false
			break
		}
	}
	if ready {
		log.Printf("Task %s preconditions met, applying Task action", t.name)
		// Apply the Task action and update the state
		err := t.action(state)
		if err != nil {
			return nil, err
		}
		// Set the Task to 'complete' so it doesn't execute again
		t.complete = true
	}
	return state, nil
}

func (t *PrimitiveTask) IsComplete() bool {
	return t.complete
}

func (t *PrimitiveTask) Name() string {
	return t.name
}

func (t *PrimitiveTask) String() string {
	preconditions := make([]string, 0)
	for _, condition := range t.preconditions {
		preconditions = append(preconditions, condition.String())
	}
	return fmt.Sprintf("[%s] preconditions: [%s], complete: %t", t.name, strings.Join(preconditions, ","), t.complete)
}

// GoalTask implements the HTN goal Task, composed of preconditions that are other Tasks.  The goal Task is considered
// complete when all condition Tasks are themselves complete.
type GoalTask struct {
	preconditions []TaskCondition
	complete      bool
}

func NewGoalTask(preconditions []TaskCondition) *GoalTask {
	return &GoalTask{
		preconditions: preconditions,
		complete:      false,
	}
}

func (g *GoalTask) Execute(state *State) (*State, error) {
	log.Println("executing goal Task")
	if !g.complete {
		log.Println("goal Task is not complete checking preconditions")
		for _, condition := range g.preconditions {
			if !condition.IsMet(state) {
				log.Println("goal precondition not met, exiting")
				return state, nil
			}
		}
		log.Println("goal conditions met, goal Task is complete.")
		g.complete = true
	}
	return state, nil
}

func (g *GoalTask) IsComplete() bool {
	return g.complete
}

func (g *GoalTask) Name() string {
	return "goal"
}

func (g *GoalTask) String() string {
	preconditions := make([]string, 0)
	for _, condition := range g.preconditions {
		preconditions = append(preconditions, condition.String())
	}
	return fmt.Sprintf("goal: preconditions: [%s], complete: %t", strings.Join(preconditions, ","), g.complete)
}
