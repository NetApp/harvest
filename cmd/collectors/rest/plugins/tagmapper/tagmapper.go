package tagmapper

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"log/slog"
	"strings"
)

// TagMapper is a plugin that parses tags configured in plugin and maps them to instance labels.
// It extracts key-value pairs from comma-separated tag strings and applies them
// as labels on matrix instances for configured tag keys.
type TagMapper struct {
	*plugin.AbstractPlugin
	tags           *set.Set
	exportedLabels *set.Set // Track what's already in export_options
}

// New creates a new TagMapper plugin instance.
func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &TagMapper{AbstractPlugin: p}
}

func (t *TagMapper) Init(_ conf.Remote) error {
	if err := t.InitAbc(); err != nil {
		return err
	}
	tagMapper := t.Params.GetChildren()
	t.tags = set.New()
	t.exportedLabels = set.New()
	for _, tag := range tagMapper {
		t.tags.Add(tag.GetContentS())
	}
	exportOption := t.ParentParams.GetChildS("export_options")
	if exportOption != nil {
		if exportedLabels := exportOption.GetChildS("instance_labels"); exportedLabels != nil {
			// Add existing labels to our tracking set
			for _, label := range exportedLabels.GetChildren() {
				t.exportedLabels.Add(label.GetContentS())
			}

			for _, label := range t.tags.Values() {
				if !t.exportedLabels.Has(label) {
					exportedLabels.NewChildS("", label)
					t.exportedLabels.Add(label)
				}
			}
		}
	}
	return nil
}

func (t *TagMapper) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[t.Object]
	for _, instance := range data.GetInstances() {
		tags := instance.GetLabel("tags")
		if tags != "" {
			tagMap := t.parseTagsToMap(tags)
			for key, value := range tagMap {
				// If no tags are configured, set all parsed tags as labels
				// If tags are configured, only set labels for configured tags
				if t.tags.IsEmpty() || t.tags.Has(key) {
					if instance.GetLabel(key) == "" {
						instance.SetLabel(key, value)

						// If tags is empty, add to export_options instance_labels if not already present
						if t.tags.IsEmpty() && !t.exportedLabels.Has(key) {
							t.addToExportOptions(key)
							t.exportedLabels.Add(key) // Track that we've added it
						}
					} else {
						t.SLogger.Warn("label already exists",
							slog.String("label", key),
							slog.String("value", value),
							slog.String("volume", instance.GetLabel("volume")),
						)
					}
				}
			}
		}
	}
	return nil, nil, nil
}

func (t *TagMapper) addToExportOptions(labelKey string) {
	exportOption := t.ParentParams.GetChildS("export_options")
	if exportOption == nil {
		exportOption = t.ParentParams.NewChildS("export_options", "")
	}

	instanceLabels := exportOption.GetChildS("instance_labels")
	if instanceLabels == nil {
		instanceLabels = exportOption.NewChildS("instance_labels", "")
	}

	// No need to check if exists - we already track this in memory
	instanceLabels.NewChildS("", labelKey)
}

func (t *TagMapper) parseTagsToMap(tags string) map[string]string {
	tagMap := make(map[string]string)
	if tags == "" {
		return tagMap
	}
	pairs := strings.SplitSeq(tags, ",")
	for pair := range pairs {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" && value != "" {
				tagMap[key] = value
			}
		}
	}
	return tagMap
}
