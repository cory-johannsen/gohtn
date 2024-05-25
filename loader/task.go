package loader

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cory-johannsen/gohtn/config"
	"github.com/cory-johannsen/gohtn/engine"
	"github.com/cory-johannsen/gohtn/gohtn"
	"log"
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

type TaskLoader struct {
	Specs map[string]*TaskSpec
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

func (l *TaskLoader) LoadTaskResolvers(cfg *config.Config, engine *engine.Engine) (map[string]gohtn.TaskResolver, error) {
	if l.Specs == nil {
		l.Specs = make(map[string]*TaskSpec)
	}

	// filepath.Walk traverses in lexicographical order, but the taskResolvers need to be loaded primitive, compound, then goal to satisfy dependencies in order
	// load the primitive task specs
	log.Printf("loading primitive task specs")
	primitivePath := fmt.Sprintf("%s/%s/%s", cfg.AssetRoot, cfg.TaskPath, Primitive)
	primitiveTasks, err := loadTaskSpecs(Primitive, primitivePath)
	if err != nil {
		return nil, err
	}
	for name, primitiveTask := range primitiveTasks {
		l.Specs[name] = primitiveTask
	}
	// load the compound task specs
	log.Printf("loading compound task specs")
	compoundPath := fmt.Sprintf("%s/%s/%s", cfg.AssetRoot, cfg.TaskPath, Compound)
	compoundTasks, err := loadTaskSpecs(Compound, compoundPath)
	if err != nil {
		return nil, err
	}
	for name, compoundTask := range compoundTasks {
		l.Specs[name] = compoundTask
	}
	// load the goal task specs
	log.Printf("loading goal task specs")
	goalPath := fmt.Sprintf("%s/%s/%s", cfg.AssetRoot, cfg.TaskPath, Goal)
	goalTasks, err := loadTaskSpecs(Goal, goalPath)
	if err != nil {
		return nil, err
	}
	for name, goalTask := range goalTasks {
		l.Specs[name] = goalTask
	}

	// iterate the specs and load the taskResolvers
	log.Printf("loading taskResolvers from specs")
	taskResolvers := make(map[string]gohtn.TaskResolver)
	for _, taskSpec := range l.Specs {
		taskResolver, err := l.LoadTask(cfg, taskSpec, engine)
		if err != nil {
			return nil, err
		}
		taskResolvers[taskSpec.TaskName] = taskResolver
	}
	return taskResolvers, nil
}

func loadTaskSpecs(taskType TaskType, path string) (map[string]*TaskSpec, error) {
	specs := make(map[string]*TaskSpec)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}
		log.Printf("loading %s task spec %s", taskType, path)
		spec, err := loadTaskSpec(taskType, path)
		if err != nil {
			return err
		}
		specs[spec.TaskName] = spec
		return nil
	}
	err := filepath.Walk(path, walkFn)
	if err != nil {
		return nil, fmt.Errorf("error walking the path %q: %v", path, err)
	}
	return specs, nil
}

func loadTaskSpec(taskType TaskType, path string) (*TaskSpec, error) {
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
	return spec, nil
}

func (l *TaskLoader) LoadTask(cfg *config.Config, spec *TaskSpec, engine *engine.Engine) (gohtn.TaskResolver, error) {
	task, err := initTask(spec.TaskType)
	if err != nil {
		return nil, err
	}
	resolver := func() (gohtn.Task, error) {
		existing, ok := engine.Tasks[spec.TaskName]
		if ok {
			return existing, nil
		}
		log.Printf("instantiating task %s", spec.TaskName)
		t, err := l.instantiateTask(cfg, task, spec, engine)
		if err != nil {
			return nil, err
		}
		engine.Tasks[spec.TaskName] = t
		return t, nil
	}
	engine.TaskResolvers[spec.TaskName] = resolver
	return resolver, nil
}

func (l *TaskLoader) instantiateTask(cfg *config.Config, task gohtn.Task, spec *TaskSpec, engine *engine.Engine) (gohtn.Task, error) {
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
			log.Printf("task %s fetching method %s", spec.TaskName, methodName)
			method, ok := engine.Methods[methodName]
			if !ok {
				// direct load the method
				log.Printf("task %s method %s not found, loading it", spec.TaskName, methodName)
				methodPath := fmt.Sprintf("%s/%s/%s.json", cfg.AssetRoot, cfg.MethodPath, methodName)
				loadedMethod, err := LoadMethod(cfg, methodPath, l, engine)
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
			taskResolver, ok := engine.TaskResolvers[taskName]
			if !ok {
				return nil, fmt.Errorf("task %s precondition task %s not found", spec.TaskName, taskName)
			}
			task, err := taskResolver()
			if err != nil {
				return nil, err
			}
			task.(*gohtn.GoalTask).Preconditions = append(task.(*gohtn.GoalTask).Preconditions, &gohtn.TaskCondition{
				Task: task,
			})
			task.(*gohtn.GoalTask).TaskName = spec.TaskName
		}
	}
	return task, nil
}
