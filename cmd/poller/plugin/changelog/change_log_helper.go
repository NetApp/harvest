package changelog

import (
	"github.com/netapp/harvest/v2/pkg/logging"
	"gopkg.in/yaml.v3"
	"strings"
)

// Entry represents a single ChangeLog entry
type Entry struct {
	Object        string   `yaml:"object"`
	TrackLabels   []string `yaml:"track_labels"`
	PublishLabels []string `yaml:"publish_labels"`
}

// Config represents the structure of the ChangeLog configuration
type Config struct {
	ChangeLogs []Entry `yaml:"ChangeLog"`
}

// defaultChangeLogTemplate is the default ChangeLog configuration
const defaultChangeLogTemplate = `
ChangeLog:
  - object: svm
    track_labels:
      - svm
      - state
      - type
    publish_labels:
      - svm
  - object: node
    track_labels:
      - node
      - location
    publish_labels:
      - node
  - object: volume
    track_labels:
      - node
      - volume
      - svm
      - style
      - type
    publish_labels:
      - volume
      - svm
`

// getChangeLogConfig returns a map of ChangeLog entries for the given object
func getChangeLogConfig(object string, s string, logger *logging.Logger) (map[string]Entry, error) {
	var config Config
	temp := defaultChangeLogTemplate
	if s != "" {
		temp = preprocessOverwrite(object, s)
		err := yaml.Unmarshal([]byte(temp), &config)
		if err != nil {
			logger.Warn().Err(err).Str("template", s).Msg("failed to parse changelog dsl. Trying default")
			temp = defaultChangeLogTemplate
		}
	}
	err := yaml.Unmarshal([]byte(temp), &config)
	if err != nil {
		return nil, err
	}

	objectMap := make(map[string]Entry)
	for _, obj := range config.ChangeLogs {
		if obj.Object == object {
			objectMap[obj.Object] = obj
		}
	}

	return objectMap, nil
}

// preprocessOverwrite updates the ChangeLog configuration by adding the given object and its properties
func preprocessOverwrite(object string, configStr string) string {
	// Split the YAML content into lines
	lines := strings.Split(configStr, "\n")

	// Add four spaces to indent each line, making them at the same level as object
	indentedLines := make([]string, len(lines))
	for i, line := range lines {
		indentedLines[i] = "    " + line
	}

	// Join the indented lines back into a single string
	indentedYaml := strings.Join(indentedLines, "\n")

	// Add the ChangeLog prefix
	prefix := `
ChangeLog:
  - object: ` + object

	newYaml := strings.Join([]string{prefix, indentedYaml}, "\n")
	return newYaml

}
