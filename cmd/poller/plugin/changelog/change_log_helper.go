package changelog

import (
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"gopkg.in/yaml.v3"
	"strconv"
	"strings"
)

// Entry represents a single ChangeLog entry
type Entry struct {
	Object        string   `yaml:"object"`
	Track         []string `yaml:"track"`
	PublishLabels []string `yaml:"publish_labels"`
	includeAll    bool
}

// Config represents the structure of the ChangeLog configuration
type Config struct {
	ChangeLogs []Entry `yaml:"ChangeLog"`
}

// defaultChangeLogTemplate is the default ChangeLog configuration
const defaultChangeLogTemplate = `
ChangeLog:
  - object: svm
    track:
      - svm
      - state
      - type
  - object: node
    track:
      - node
      - location
  - object: volume
    track:
      - node
      - volume
      - svm
      - style
      - type
`

// getChangeLogConfig returns a map of ChangeLog entries for the given object
func getChangeLogConfig(parentParams *node.Node, s string, logger *logging.Logger) (Entry, error) {
	var config Config
	object := parentParams.GetChildS("object").GetContentS()

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
		return Entry{}, err
	}

	for _, obj := range config.ChangeLogs {
		if obj.Object == object {
			// populate publish_labels if they are empty
			if obj.PublishLabels == nil {
				if exportOption := parentParams.GetChildS("export_options"); exportOption != nil {
					if exportedKeys := exportOption.GetChildS("instance_keys"); exportedKeys != nil {
						obj.PublishLabels = append(obj.PublishLabels, exportedKeys.GetAllChildContentS()...)
					} else if x := exportOption.GetChildContentS("include_all_labels"); x != "" {
						if includeAllLabels, err := strconv.ParseBool(x); err != nil {
							logger.Logger.Error().Err(err).Msg("parameter: include_all_labels")
						} else {
							if includeAllLabels {
								obj.includeAll = true
							}
						}
					}
				}
			}
			return obj, nil
		}
	}

	return Entry{}, nil
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
