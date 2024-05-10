package gohtn

import "fmt"

type Sensor interface {
}

type State struct {
	Sensors []*Sensor
}

type Task interface {
	Execute(state *State) (*State, error)
}

type PrimitiveTask struct {
}

type TaskNode struct {
	task     Task
	children []*TaskNode
}

type TaskGraph struct {
	root *TaskNode
}

type Plan []Task

type Planner struct {
	tasks *TaskGraph
}

func (p *Planner) Plan(state *State) (Plan, error) {
	return nil, nil
}

func Execute(plan Plan, state *State) (*State, error) {
	return nil, nil
}

func main() {
	tasks := &TaskGraph{
		root: &TaskNode{
			task:     nil,
			children: make([]*TaskNode, 0),
		},
	}
	planner := &Planner{
		tasks: tasks,
	}
	state := &State{
		Sensors: make([]*Sensor, 0),
	}

	plan, err := planner.Plan(state)
	if err != nil {
		panic(err)
	}
	result, err := Execute(plan, state)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
