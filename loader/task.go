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
	tasks := make([]gohtn.Task, 0)
	// filepath.Walk traverses in lexicographical order, but the tasks needed to be loaded primitive, compound, then goal to satisfy dependencies in order
	// load the primitive tasks
	primitivePath := fmt.Sprintf("%s/%s/%s", cfg.AssetRoot, cfg.TaskPath, Primitive)
	primitiveTasks, err := loadTasks(cfg, Primitive, primitivePath, engine)
	if err != nil {
		return nil, err
	}
	for _, primitiveTask := range primitiveTasks {
		tasks = append(tasks, primitiveTask)
	}
	// load the compound tasks
	compoundPath := fmt.Sprintf("%s/%s/%s", cfg.AssetRoot, cfg.TaskPath, Compound)
	compoundTasks, err := loadTasks(cfg, Compound, compoundPath, engine)
	if err != nil {
		return nil, err
	}
	for _, compoundTask := range compoundTasks {
		tasks = append(tasks, compoundTask)
	}
	// load the goal tasks
	goalPath := fmt.Sprintf("%s/%s/%s", cfg.AssetRoot, cfg.TaskPath, Goal)
	goalTasks, err := loadTasks(cfg, Goal, goalPath, engine)
	if err != nil {
		return nil, err
	}
	for _, goalTask := range goalTasks {
		tasks = append(tasks, goalTask)
	}
	return tasks, nil
}

func loadTasks(cfg *config.Config, taskType TaskType, path string, engine *engine.Engine) ([]gohtn.Task, error) {
	tasks := make([]gohtn.Task, 0)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		task, err := loadTask(cfg, taskType, path, engine)
		if err != nil {
			return err
		}
		tasks = append(tasks, task)
		return nil
	}
	err := filepath.Walk(path, walkFn)
	if err != nil {
		return nil, fmt.Errorf("error walking the path %q: %v", path, err)
	}
	return tasks, nil
}

func loadTask(cfg *config.Config, taskType TaskType, path string, engine *engine.Engine) (gohtn.Task, error) {
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
	t, err := instantiateTask(cfg, task, spec, engine)
	if err != nil {
		return nil, err
	}
	engine.Tasks[t.Name()] = t
	return task, nil
}

func instantiateTask(cfg *config.Config, task gohtn.Task, spec *TaskSpec, engine *engine.Engine) (gohtn.Task, error) {
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
		task.(*gohtn.PrimitiveTask).TaskName = spec.TaskName
	case Compound:
		// compound task preconditions are Methods
		for _, methodName := range spec.Preconditions {
			method, ok := engine.Methods[methodName]
			if !ok {
				// direct load the method
				methodPath := fmt.Sprintf("%s/%s/%s.json", cfg.AssetRoot, cfg.MethodPath, methodName)
				loadedMethod, err := LoadMethod(methodPath, engine)
				if err != nil {
					return nil, err
				}
				engine.Methods[methodName] = loadedMethod
				method = loadedMethod
			}
			task.(*gohtn.CompoundTask).Methods = append(task.(*gohtn.CompoundTask).Methods, method)
			task.(*gohtn.CompoundTask).TaskName = spec.TaskName
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
			task.(*gohtn.GoalTask).TaskName = spec.TaskName
		}
	}
	return task, nil
}
