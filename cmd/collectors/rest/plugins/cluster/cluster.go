package cluster

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"strings"
)

type Cluster struct {
	*plugin.AbstractPlugin
	tags *matrix.Matrix
	addr string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Cluster{AbstractPlugin: p}
}

func (c *Cluster) Init(_ conf.Remote) error {
	var err error

	if err := c.InitAbc(); err != nil {
		return err
	}

	c.tags = matrix.New(c.Parent+".Cluster", "cluster", "cluster")
	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "tag")
	c.tags.SetExportOptions(exportOptions)
	_, err = c.tags.NewMetricFloat64("tags", "tags")
	if err != nil {
		c.SLogger.Error("add metric", slogx.Err(err))
		return err
	}

	ap, err := conf.PollerNamed(c.Options.Poller)
	if err != nil {
		return err
	}
	c.addr = ap.Addr
	return nil
}

func (c *Cluster) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[c.Object]

	// Based on the tags array, cluster_tags instances/metrics would be created
	c.handleTags(data)

	data.SetGlobalLabel("addr", c.addr)
	data.SetGlobalLabel("poller", c.Options.Poller)

	return []*matrix.Matrix{c.tags}, nil, nil
}

func (c *Cluster) handleTags(data *matrix.Matrix) {
	var (
		tagInstance *matrix.Instance
		err         error
	)

	// Purge and reset data
	c.tags.PurgeInstances()
	c.tags.Reset()

	// Set all global labels
	c.tags.SetGlobalLabels(data.GetGlobalLabels())

	// Based on the tags array, cluster_tags instances/metrics would be created.
	for _, cluster := range data.GetInstances() {
		if tags := cluster.GetLabel("tags"); tags != "" {
			for _, tag := range strings.Split(tags, ",") {
				tagInstanceKey := data.GetGlobalLabels()["cluster"] + tag
				if tagInstance, err = c.tags.NewInstance(tagInstanceKey); err != nil {
					c.SLogger.Error(
						"Failed to create tag instance",
						slogx.Err(err),
						slog.String("tagInstanceKey", tagInstanceKey),
					)
					return
				}

				tagInstance.SetLabel("tag", tag)

				m := c.tags.GetMetric("tags")
				// populate numeric data
				m.SetValueFloat64(tagInstance, 1.0)
			}
		}
	}
}
