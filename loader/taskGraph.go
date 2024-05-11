package loader

import (
	"encoding/json"
	"fmt"
	"github.com/cory-johannsen/gohtn/config"
	"github.com/cory-johannsen/gohtn/engine"
	"github.com/cory-johannsen/gohtn/gohtn"
	"os"
)

type TaskNodeSpec struct {
	Task     string          `json:"task,omitempty"`
	Children []*TaskNodeSpec `json:"children,omitempty"`
}

type TaskGraphSpec struct {
	Root *TaskNodeSpec `json:"root"`
}

func LoadTaskGraph(cfg *config.Config, engine *engine.Engine) (*gohtn.TaskGraph, error) {
	taskGraphPath := fmt.Sprintf("%s/%s", cfg.AssetRoot, cfg.TaskGraphPath)
	spec := &TaskGraphSpec{}
	buffer, err := os.ReadFile(taskGraphPath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buffer, spec)
	if err != nil {
		return nil, err
	}
	root, err := loadTaskNode(spec.Root, engine)
	if err != nil {
		return nil, err
	}
	taskGraph := &gohtn.TaskGraph{
		Root: root,
	}

	return taskGraph, nil
}

func loadTaskNode(spec *TaskNodeSpec, engine *engine.Engine) (*gohtn.TaskNode, error) {
	task, ok := engine.Tasks[spec.Task]
	if !ok {
		return nil, fmt.Errorf("task %s not found", spec.Task)
	}
	children := make([]*gohtn.TaskNode, 0)
	for _, childSpec := range spec.Children {
		child, err := loadTaskNode(childSpec, engine)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}
	node := &gohtn.TaskNode{
		Task:     task,
		Children: children,
	}
	return node, nil
}
