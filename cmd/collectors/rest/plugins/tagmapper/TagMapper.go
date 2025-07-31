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
	tags *set.Set
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
	for _, tag := range tagMapper {
		t.tags.Add(tag.GetContentS())
	}
	exportOption := t.ParentParams.GetChildS("export_options")
	if exportOption != nil {
		if exportedLabels := exportOption.GetChildS("instance_labels"); exportedLabels != nil {
			for _, label := range t.tags.Values() {
				exportedLabels.NewChildS("", label)
			}
		}
	}
	return nil
}

func (t *TagMapper) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[t.Object]
	if t.tags.IsEmpty() {
		t.SLogger.Warn("No tags are configured in plugin")
		return nil, nil, nil
	}
	for _, instance := range data.GetInstances() {
		tags := instance.GetLabel("tags")
		if tags != "" {
			tagMap := t.parseTagsToMap(tags)
			for key, value := range tagMap {
				if t.tags.Has(key) {
					if instance.GetLabel(key) == "" {
						instance.SetLabel(key, value)
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

func (t *TagMapper) parseTagsToMap(tags string) map[string]string {
	tagMap := make(map[string]string)
	if tags == "" {
		return tagMap
	}
	pairs := strings.Split(tags, ",")
	for _, pair := range pairs {
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
