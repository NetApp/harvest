package staticcounter

import (
	"github.com/goccy/go-yaml"
	"log/slog"
	"os"
)

type CounterDefinition struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	BaseCounter string `yaml:"base_counter,omitempty"`
}

type ObjectCounters struct {
	CounterDefinitions map[string]CounterDefinition `yaml:"-"`
}

type object struct {
	CounterDefinitions []CounterDefinition `yaml:"counter_definitions"`
}

type definitions struct {
	Objects map[string]object `yaml:"objects"`
}

func LoadStaticCounterDefinitions(objectName, filePath string, logger *slog.Logger) (ObjectCounters, error) {
	var staticDefinitions definitions
	var objectCounters ObjectCounters

	data, err := os.ReadFile(filePath)
	if err != nil {
		return objectCounters, err
	}
	if err := yaml.Unmarshal(data, &staticDefinitions); err != nil {
		return objectCounters, err
	}

	obj, exists := staticDefinitions.Objects[objectName]
	if !exists {
		return objectCounters, nil
	}

	allCounterDefs := make(map[string]CounterDefinition)
	for _, def := range obj.CounterDefinitions {
		allCounterDefs[def.Name] = def
	}

	objectCounters.CounterDefinitions = make(map[string]CounterDefinition)
	for _, def := range obj.CounterDefinitions {
		if def.Type == "" {
			logger.Error("Missing type in counter definition", slog.String("filePath", filePath), slog.String("counterName", def.Name))
			continue
		}
		if def.BaseCounter != "" {
			if _, baseCounterExists := allCounterDefs[def.BaseCounter]; !baseCounterExists {
				logger.Error("Base counter definition not found", slog.String("filePath", filePath), slog.String("counterName", def.Name), slog.String("baseCounter", def.BaseCounter))
				continue
			}
		}
		objectCounters.CounterDefinitions[def.Name] = def
	}
	return objectCounters, nil
}
