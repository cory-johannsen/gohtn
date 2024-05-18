package gohtn

import (
	"log"
)

type TaskNode struct {
	TaskResolver TaskResolver
	Children     []*TaskNode
}

type TaskGraph struct {
	Root *TaskNode
}

type Plan []Task

type Planner struct {
	Tasks *TaskGraph
}

func evaluateNode(node *TaskNode, state *State) []Task {
	task, err := node.TaskResolver()
	if err != nil {
		panic(err)
	}
	log.Printf("evaluating node {%s}", task.String())
	tasks := make([]Task, 0)
	if !task.IsComplete() {
		log.Printf("node {%s} is not complete", task.String())
		tasks = append(tasks, task)
	}
	for _, child := range node.Children {
		childTasks := evaluateNode(child, state)
		for _, childTask := range childTasks {
			tasks = append([]Task{childTask}, tasks...)
		}
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
	log.Printf("plan contains %d TaskResolvers", len(plan))
	return plan, nil
}

func Execute(plan Plan, state *State) (*State, error) {
	log.Printf("executing plan with %d TaskResolvers", len(plan))
	for _, task := range plan {
		postState, err := task.Execute(state)
		if err != nil {
			return nil, err
		}
		log.Printf("postState: %s", postState.String())
	}
	return state, nil
}
