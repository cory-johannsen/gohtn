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

type Tasks map[string]Task
type TaskResolver func() (Task, error)
type TaskResolvers map[string]TaskResolver

// Action is an action applied by a Task.
type Action func(state *State) error

// PrimitiveTask implements the HTN primitive Task.   It contains a set of preconditions that must be met
// before it will execute.  Once the preconditions are met, the Action is applied, then the completion flag is set.
type PrimitiveTask struct {
	Preconditions []Condition `json:"preconditions"`
	Complete      bool        `json:"complete"`
	Action        Action      `json:"action"`
	TaskName      string      `json:"name"`
}

func (t *PrimitiveTask) Execute(state *State) (*State, error) {
	preconditions := make([]string, 0)
	for _, condition := range t.Preconditions {
		preconditions = append(preconditions, condition.String())
	}
	log.Printf("executing Task {%s}, preconditions {%s}", t.Name(), strings.Join(preconditions, ","))
	// Determine if the Task preconditions have been met
	var ready = true
	for _, condition := range t.Preconditions {
		log.Printf("evaluating condition {%s}", condition.String())
		if !condition.IsMet(state) {
			ready = false
			break
		}
	}
	if ready {
		log.Printf("Task {%s} preconditions met, applying Task action", t.Name())
		// Apply the Task action and update the state
		err := t.Action(state)
		if err != nil {
			return nil, err
		}
		// Set the Task to 'complete' so it doesn't execute again
		t.Complete = true
	}
	return state, nil
}

func (t *PrimitiveTask) IsComplete() bool {
	return t.Complete
}

func (t *PrimitiveTask) Name() string {
	return t.TaskName
}

func (t *PrimitiveTask) String() string {
	preconditions := make([]string, 0)
	for _, condition := range t.Preconditions {
		preconditions = append(preconditions, condition.String())
	}
	return fmt.Sprintf("[%s] preconditions: [%s], complete: %t", t.Name(), strings.Join(preconditions, ","), t.Complete)
}

// GoalTask implements the HTN goal Task, composed of preconditions that are other TaskResolvers.  The goal Task is considered
// complete when all condition TaskResolvers are themselves complete.
type GoalTask struct {
	Preconditions []*TaskCondition `json:"preconditions"`
	Complete      bool             `json:"complete"`
	TaskName      string           `json:"name"`
}

func (g *GoalTask) Execute(state *State) (*State, error) {
	log.Println("executing goal Task")
	if !g.Complete {
		log.Println("goal Task is not complete checking preconditions")
		for _, condition := range g.Preconditions {
			if !condition.IsMet(state) {
				log.Println("goal precondition not met, exiting")
				return state, nil
			}
		}
		log.Println("goal conditions met, goal Task is complete.")
		g.Complete = true
	}
	return state, nil
}

func (g *GoalTask) IsComplete() bool {
	return g.Complete
}

func (g *GoalTask) Name() string {
	return g.TaskName
}

func (g *GoalTask) String() string {
	preconditions := make([]string, 0)
	for _, condition := range g.Preconditions {
		preconditions = append(preconditions, fmt.Sprintf("{%s}", condition.String()))
	}
	return fmt.Sprintf("goal: preconditions: [%s], complete: %t", strings.Join(preconditions, ","), g.Complete)
}

type Method struct {
	Conditions    []Condition
	TaskResolvers TaskResolvers
	Name          string
}

func (m *Method) Applies(state *State) bool {
	log.Printf("checking if method {%s} applies", m.Name)
	for _, condition := range m.Conditions {
		if !condition.IsMet(state) {
			log.Printf("method {%s} condition {%s} not met, exiting", m.Name, condition.String())
			return false
		}
	}
	return true
}

func (m *Method) Execute(state *State) (int64, error) {
	log.Printf("executing method {%s}", m.Name)
	var executed = int64(0)
	tasks := make([]Task, 0)
	for _, taskResolver := range m.TaskResolvers {
		task, err := taskResolver()
		if err != nil {
			return 0, err
		}
		tasks = append([]Task{task}, tasks...)
	}
	for _, task := range tasks {
		if !task.IsComplete() {
			log.Printf("method {%s} task {%s} not complete, executing it", m.Name, task.String())
			_, err := task.Execute(state)
			if err != nil {
				return -1, err
			}
			executed++
		}
	}
	return executed, nil
}

func (m *Method) String() string {
	conditions := make([]string, 0)
	for _, condition := range m.Conditions {
		conditions = append(conditions, fmt.Sprintf("{%s}", condition.String()))
	}
	tasks := make([]string, 0)
	for taskName := range m.TaskResolvers {
		tasks = append(tasks, fmt.Sprintf("{%s}", taskName))
	}
	return fmt.Sprintf("Method %s: conditions: [%s], tasks: [%s]", m.Name, strings.Join(conditions, ","), strings.Join(tasks, ","))
}

// CompoundTask implements the HTN compound task, which consists of a ranked list of methods and a name.
// The task selects a method at execution time by checking the conditions on each.  Since the method list
// is in priority order, the first match is selected when more than one apply.
type CompoundTask struct {
	Methods  []*Method `json:"methods"`
	TaskName string    `json:"name"`
	Complete bool      `json:"complete"`
}

func (c *CompoundTask) Execute(state *State) (*State, error) {
	log.Printf("executing compound task {%s}", c.Name())
	applicableMethods := make([]*Method, 0)
	for _, method := range c.Methods {
		if method.Applies(state) {
			applicableMethods = append(applicableMethods, method)
		}
	}
	if len(applicableMethods) == 0 {
		log.Println("no applicable methods found")
		c.Complete = true
		return state, nil
	}
	// The methods are stored in priority order, so the first one is the selected choice
	selectedMethod := applicableMethods[0]
	executedTasks, err := selectedMethod.Execute(state)
	if err != nil {
		return nil, err
	}
	if executedTasks == 0 {
		c.Complete = true
	}

	return state, nil
}

func (c *CompoundTask) Name() string {
	return c.TaskName
}

func (c *CompoundTask) IsComplete() bool {
	return c.Complete
}

func (c *CompoundTask) String() string {
	methods := make([]string, 0)
	for _, method := range c.Methods {
		methods = append(methods, fmt.Sprintf("{%s}", method.String()))
	}
	return fmt.Sprintf("CompoundTask %s: methods: [%s]", c.Name(), strings.Join(methods, ","))
}
