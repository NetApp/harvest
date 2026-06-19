package loader

import (
	"encoding/json"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
)

// LoadMetricDescriptions loads all metric descriptions from JSON files matching *_metrics.json
func LoadMetricDescriptions(metadataDir string, logger *slog.Logger) map[string]string {
	descriptions := make(map[string]string)

	matches, err := filepath.Glob(filepath.Join(metadataDir, "*_metrics.json"))
	if err != nil {
		logger.Warn("failed to glob metadata directory", slog.String("dir", metadataDir), slog.Any("error", err))
		return descriptions
	}

	loadedCount := 0
	for _, filePath := range matches {
		filename := filepath.Base(filePath)
		data, err := os.ReadFile(filePath)
		if err != nil {
			logger.Warn("failed to read metadata file", slog.String("file", filename), slog.Any("error", err))
			continue
		}
		var fileDescriptions map[string]string
		if err := json.Unmarshal(data, &fileDescriptions); err != nil {
			logger.Warn("failed to parse metadata file", slog.String("file", filename), slog.Any("error", err))
			continue
		}
		maps.Copy(descriptions, fileDescriptions)
		loadedCount++
		logger.Info("loaded metadata file", slog.String("file", filename), slog.Int("metrics", len(fileDescriptions)))
	}

	logger.Info("metadata loading complete",
		slog.Int("files_loaded", loadedCount),
		slog.Int("total_metrics", len(descriptions)))

	return descriptions
}

type PromptDefinition struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"-"`    // Loaded from file
	File        string `json:"file"` // File path for content
}

func LoadPromptDefinitions(promptsDir string, logger *slog.Logger) ([]PromptDefinition, error) {
	jsonPath := filepath.Join(promptsDir, "prompts.json")

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		logger.Warn("prompts.json not found, no prompts will be loaded",
			slog.String("path", jsonPath),
			slog.Any("error", err))
		return []PromptDefinition{}, nil
	}

	var prompts []PromptDefinition
	if err := json.Unmarshal(data, &prompts); err != nil {
		logger.Error("failed to parse prompts.json",
			slog.String("path", jsonPath),
			slog.Any("error", err))
		return []PromptDefinition{}, err
	}

	validPrompts := make([]PromptDefinition, 0, len(prompts))
	for _, prompt := range prompts {
		if prompt.File == "" {
			logger.Warn("prompt missing file field, skipping",
				slog.String("name", prompt.Name))
			continue
		}

		filePath := filepath.Join(promptsDir, prompt.File)
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			logger.Warn("failed to read prompt file, skipping prompt",
				slog.String("name", prompt.Name),
				slog.String("file", prompt.File),
				slog.String("path", filePath),
				slog.Any("error", err))
			continue
		}

		prompt.Content = string(fileData)
		validPrompts = append(validPrompts, prompt)

		logger.Info("loaded prompt",
			slog.String("name", prompt.Name),
			slog.String("file", prompt.File),
			slog.Int("content_length", len(prompt.Content)))
	}

	logger.Info("prompt loading complete",
		slog.String("file", "prompts.json"),
		slog.Int("total_defined", len(prompts)),
		slog.Int("successfully_loaded", len(validPrompts)))

	return validPrompts, nil
}
