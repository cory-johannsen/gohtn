package loader

import (
	"encoding/json"
	"fmt"
	"github.com/cory-johannsen/gohtn/config"
	"github.com/cory-johannsen/gohtn/engine"
	"github.com/cory-johannsen/gohtn/gohtn"
	"os"
	"path/filepath"
	"strings"
)

type ConditionType string

const (
	Comparison ConditionType = "comparison"
	Flag       ConditionType = "flag"
	NotFlag    ConditionType = "notflag"
)

func LoadConditions(cfg *config.Config) (engine.Conditions, error) {
	conditionsPath := fmt.Sprintf("%s/%s", cfg.AssetRoot, cfg.ConditionPath)
	conditions := make(engine.Conditions)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		conditionSubpath := strings.TrimPrefix(path, fmt.Sprintf("%s/", conditionsPath))
		pathComponents := strings.Split(conditionSubpath, "/")
		conditionType := ConditionType(pathComponents[0])
		conditionName := strings.TrimSuffix(info.Name(), ".json")
		condition, err := loadCondition(conditionType, path)
		if err != nil {
			return err
		}
		conditions[conditionName] = condition
		return nil
	}
	err := filepath.Walk(conditionsPath, walkFn)
	if err != nil {
		return nil, fmt.Errorf("error walking the path %q: %v", conditionsPath, err)
	}
	return conditions, nil
}

func initCondition(conditionType ConditionType) (gohtn.Condition, error) {
	switch conditionType {
	case Comparison:
		return &gohtn.ComparisonCondition{}, nil
	case Flag:
		return &gohtn.FlagCondition{}, nil
	case NotFlag:
		return &gohtn.NotFlagCondition{}, nil
	}
	return nil, fmt.Errorf("unknown condition type %s", conditionType)
}

func loadCondition(conditionType ConditionType, path string) (gohtn.Condition, error) {
	condition, err := initCondition(conditionType)
	if err != nil {
		return nil, err
	}
	buffer, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buffer, condition)
	if err != nil {
		return nil, err
	}
	return condition, nil
}
