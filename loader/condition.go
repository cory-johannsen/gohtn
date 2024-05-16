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
	Comparison         ConditionType = "comparison"
	PropertyComparison ConditionType = "propertycomparison"
	Flag               ConditionType = "flag"
	NotFlag            ConditionType = "notflag"
	Logical            ConditionType = "logical"
)

func LoadConditions(cfg *config.Config) (engine.Conditions, error) {
	conditionsPath := filepath.Join(cfg.AssetRoot, cfg.ConditionPath)
	conditions := make(engine.Conditions)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		conditionSubpath := strings.TrimPrefix(path, conditionsPath)
		pathComponents := strings.Split(conditionSubpath, string(os.PathSeparator))
		conditionType := ConditionType(pathComponents[0])
		if len(conditionType) == 0 {
			// Handle the possibility of a leading slash
			conditionType = ConditionType(pathComponents[1])
		}
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
	case PropertyComparison:
		return &gohtn.PropertyComparisonCondition{}, nil
	case Flag:
		return &gohtn.FlagCondition{}, nil
	case NotFlag:
		return &gohtn.NotFlagCondition{}, nil
	case Logical:
		return &gohtn.LogicalCondition{}, nil
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
