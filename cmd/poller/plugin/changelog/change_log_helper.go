package changelog

import (
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"gopkg.in/yaml.v3"
	"slices"
	"strconv"
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
      - anti_ransomware_state
  - object: node
    track:
      - node
      - location
      - healthy
  - object: volume
    track:
      - node
      - volume
      - svm
      - style
      - type
      - aggr
      - state
      - status
`

// getChangeLogConfig returns a map of ChangeLog entries for the given object
func getChangeLogConfig(parentParams *node.Node, overwriteConfig []byte, logger *logging.Logger) (Entry, error) {
	var (
		config Config
		entry  Entry
		err    error
	)
	object := parentParams.GetChildS("object").GetContentS()

	useDefault := true

	if len(overwriteConfig) > 0 {
		entry, err = preprocessOverwrite(object, overwriteConfig)
		if err != nil {
			logger.Warn().Err(err).Str("template", string(overwriteConfig)).Msg("failed to parse changelog dsl. Trying default")
		} else {
			useDefault = false
		}
	}

	if useDefault {
		err = yaml.Unmarshal([]byte(defaultChangeLogTemplate), &config)
		if err != nil {
			return Entry{}, err
		}
		i := slices.IndexFunc(config.ChangeLogs, func(entry Entry) bool {
			return entry.Object == object
		})
		if i == -1 {
			return Entry{}, nil
		}
		entry = config.ChangeLogs[i]
	}

	// populate publish_labels if they are empty
	if entry.PublishLabels == nil {
		if exportOption := parentParams.GetChildS("export_options"); exportOption != nil {
			if exportedKeys := exportOption.GetChildS("instance_keys"); exportedKeys != nil {
				entry.PublishLabels = append(entry.PublishLabels, exportedKeys.GetAllChildContentS()...)
			} else if x := exportOption.GetChildContentS("include_all_labels"); x != "" {
				if includeAllLabels, err := strconv.ParseBool(x); err != nil {
					logger.Logger.Error().Err(err).Msg("parameter: include_all_labels")
				} else {
					if includeAllLabels {
						entry.includeAll = true
					}
				}
			}
		}
	}

	return entry, nil
}

// preprocessOverwrite updates the ChangeLog configuration by adding the given object and its properties
func preprocessOverwrite(object string, configStr []byte) (Entry, error) {
	var entry Entry

	err := yaml.Unmarshal(configStr, &entry)
	if err != nil {
		return entry, err
	}

	entry.Object = object
	return entry, nil

}
