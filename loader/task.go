package loader

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cory-johannsen/gohtn/config"
	"github.com/cory-johannsen/gohtn/engine"
	"github.com/cory-johannsen/gohtn/gohtn"
	"os"
	"path/filepath"
	"strings"
)

type TaskType string

const (
	Primitive TaskType = "primitive"
	Compound  TaskType = "compound"
	Goal      TaskType = "goal"
)

type TaskSpec struct {
	Preconditions []string `json:"preconditions"`
	Complete      bool     `json:"complete,omitempty"`
	Action        string   `json:"action,omitempty"`
	TaskName      string   `json:"name"`
	TaskType      TaskType `json:"type,omitempty"`
}

func initTask(taskType TaskType) (gohtn.Task, error) {
	switch taskType {
	case Primitive:
		return &gohtn.PrimitiveTask{}, nil
	case Compound:
		return &gohtn.CompoundTask{}, nil
	case Goal:
		return &gohtn.GoalTask{}, nil
	}
	return nil, errors.New("invalid task type")
}

func LoadTasks(cfg *config.Config, engine *engine.Engine) ([]gohtn.Task, error) {
	taskPath := fmt.Sprintf("%s/%s", cfg.AssetRoot, cfg.TaskPath)
	tasks := make([]gohtn.Task, 0)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		taskSubpath := strings.TrimPrefix(path, fmt.Sprintf("%s/", taskPath))
		pathComponents := strings.Split(taskSubpath, "/")
		taskType := TaskType(pathComponents[0])
		task, err := loadTask(taskType, path, engine)
		if err != nil {
			return err
		}
		tasks = append(tasks, task)
		return nil
	}
	err := filepath.Walk(taskPath, walkFn)
	if err != nil {
		return nil, fmt.Errorf("error walking the path %q: %v", taskPath, err)
	}
	return tasks, nil
}

func loadTask(taskType TaskType, path string, engine *engine.Engine) (gohtn.Task, error) {
	spec := &TaskSpec{}
	buffer, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buffer, spec)
	if err != nil {
		return nil, err
	}
	spec.TaskType = taskType
	task, err := initTask(taskType)
	if err != nil {
		return nil, err
	}
	t, err := instantiateTask(task, spec, engine)
	if err != nil {
		return nil, err
	}
	engine.Tasks[t.Name()] = t
	return task, nil
}

func instantiateTask(task gohtn.Task, spec *TaskSpec, engine *engine.Engine) (gohtn.Task, error) {
	var action gohtn.Action
	if len(spec.Action) > 0 {
		// the action is a name used to resolve the function from the action registry
		foundAction, ok := engine.Actions[spec.Action]
		if !ok {
			return nil, fmt.Errorf("task %s action %s not found", spec.TaskName, spec.Action)
		}
		action = foundAction
	}
	switch spec.TaskType {
	case Primitive:
		// primitive task preconditions are Conditions
		for _, preconditionName := range spec.Preconditions {
			precondition, ok := engine.Conditions[preconditionName]
			if !ok {
				return nil, fmt.Errorf("task %s precondition %s not found", spec.TaskName, preconditionName)
			}
			task.(*gohtn.PrimitiveTask).Preconditions = append(task.(*gohtn.PrimitiveTask).Preconditions, precondition)
		}
		task.(*gohtn.PrimitiveTask).Action = action
	case Compound:
		// compound task preconditions are Methods
		for _, methodName := range spec.Preconditions {
			method, ok := engine.Methods[methodName]
			if !ok {
				return nil, fmt.Errorf("task %s precondition method %s not found", spec.TaskName, methodName)
			}
			task.(*gohtn.CompoundTask).Methods = append(task.(*gohtn.CompoundTask).Methods, method)
		}
	case Goal:
		// goal task preconditions are TaskConditions
		for _, taskName := range spec.Preconditions {
			conditionTask, ok := engine.Tasks[taskName]
			if !ok {
				return nil, fmt.Errorf("task %s precondition task %s not found", spec.TaskName, taskName)
			}
			task.(*gohtn.GoalTask).Preconditions = append(task.(*gohtn.GoalTask).Preconditions, &gohtn.TaskCondition{
				Task: conditionTask,
			})
		}
	}
	return task, nil
}
