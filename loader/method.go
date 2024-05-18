package loader

import (
	"encoding/json"
	"fmt"
	"github.com/cory-johannsen/gohtn/config"
	"github.com/cory-johannsen/gohtn/engine"
	"github.com/cory-johannsen/gohtn/gohtn"
	"os"
	"path/filepath"
)

type MethodSpec struct {
	Name       string   `json:"name"`
	Conditions []string `json:"conditions"`
	Tasks      []string `json:"tasks"`
}

func LoadMethods(cfg *config.Config, taskLoader *TaskLoader, htnEngine *engine.Engine) (engine.Methods, error) {
	methodsPath := fmt.Sprintf("%s/%s", cfg.AssetRoot, cfg.MethodPath)
	methods := make(engine.Methods)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		methodName := info.Name()
		method, err := LoadMethod(cfg, path, taskLoader, htnEngine)
		if err != nil {
			return err
		}
		methods[methodName] = method
		return nil
	}
	err := filepath.Walk(methodsPath, walkFn)
	if err != nil {
		return nil, fmt.Errorf("error walking the path %q: %v", methodsPath, err)
	}
	return methods, nil
}

func LoadMethod(cfg *config.Config, path string, taskLoader *TaskLoader, htnEngine *engine.Engine) (*gohtn.Method, error) {
	spec := &MethodSpec{}
	buffer, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buffer, spec)
	if err != nil {
		return nil, err
	}
	method := &gohtn.Method{
		Name:          spec.Name,
		Conditions:    make([]gohtn.Condition, 0),
		TaskResolvers: make(gohtn.TaskResolvers),
	}
	for _, conditionName := range spec.Conditions {
		condition, ok := htnEngine.Conditions[conditionName]
		if !ok {
			return nil, fmt.Errorf("unknown condition: %s", conditionName)
		}
		method.Conditions = append(method.Conditions, condition)
	}
	for _, taskName := range spec.Tasks {
		taskResolver, ok := htnEngine.TaskResolvers[taskName]
		if !ok {
			taskSpec, ok := taskLoader.Specs[taskName]
			if !ok {
				return nil, fmt.Errorf("unknown taskResolver spec: %s", taskName)
			}
			loadedTask, err := taskLoader.LoadTask(cfg, taskSpec, htnEngine)
			if err != nil {
				return nil, err
			}
			taskResolver = loadedTask
			htnEngine.TaskResolvers[taskName] = taskResolver
		}
		method.TaskResolvers[taskName] = taskResolver
	}
	return method, nil
}
