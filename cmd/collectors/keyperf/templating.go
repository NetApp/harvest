package keyperf

import (
	"gopkg.in/yaml.v3"
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

type Object struct {
	CounterDefinitions []CounterDefinition `yaml:"counter_definitions"`
}

type StaticCounterDefinitions struct {
	Objects map[string]Object `yaml:"objects"`
}

func LoadStaticCounterDefinitions(object string, filePath string, logger *slog.Logger) (ObjectCounters, error) {
	var staticDefinitions StaticCounterDefinitions
	var objectCounters ObjectCounters

	data, err := os.ReadFile(filePath)
	if err != nil {
		return objectCounters, err
	}

	err = yaml.Unmarshal(data, &staticDefinitions)
	if err != nil {
		return objectCounters, err
	}

	if obj, exists := staticDefinitions.Objects[object]; exists {
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

	return objectCounters, nil
}
