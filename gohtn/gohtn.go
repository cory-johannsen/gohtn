package gohtn

import (
	"log"
)

type TaskNode struct {
	Task     Task
	Children []*TaskNode
}

type TaskGraph struct {
	Root *TaskNode
}

type Plan []Task

type Planner struct {
	Tasks *TaskGraph
}

func evaluateNode(node *TaskNode, state *State) []Task {
	log.Printf("evaluating node %v", node.Task)
	tasks := make([]Task, 0)
	if !node.Task.IsComplete() {
		log.Printf("node %v is not complete", node.Task)
		tasks = append(tasks, node.Task)
	}
	for _, child := range node.Children {
		childTasks := evaluateNode(child, state)
		tasks = append(tasks, childTasks...)
	}
	return tasks
}

func (p *Planner) Plan(state *State) (Plan, error) {
	log.Println("building plan")
	plan := make(Plan, 0)
	// walk the Task graph, starting at the root, and find the executable plan
	node := p.Tasks.Root
	if node != nil {
		tasks := evaluateNode(node, state)
		for _, task := range tasks {
			plan = append(plan, task)
		}
	}
	log.Printf("plan contains %d Tasks", len(plan))
	return plan, nil
}

func Execute(plan Plan, state *State) (*State, error) {
	log.Printf("executing plan with %d Tasks", len(plan))
	for _, task := range plan {
		postState, err := task.Execute(state)
		if err != nil {
			return nil, err
		}
		log.Printf("postState: %v", postState)
	}
	return state, nil
}
