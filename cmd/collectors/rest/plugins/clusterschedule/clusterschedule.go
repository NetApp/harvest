package clusterschedule

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"strings"
)

type ClusterSchedule struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &ClusterSchedule{AbstractPlugin: p}
}

func (c *ClusterSchedule) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[c.Object]
	localClusterName := data.GetGlobalLabels()["cluster"]
	for _, instance := range data.GetInstances() {
		if cron := instance.GetLabel("cron"); cron != "" {
			updateDetailsJSON := gjson.Result{Type: gjson.JSON, Raw: cron}
			var minStr, hourStr, dayStr, monthStr, weekDayStr string
			var cronVal []string

			minStr = list(updateDetailsJSON.Get("minutes"))
			hourStr = list(updateDetailsJSON.Get("hours"))
			dayStr = list(updateDetailsJSON.Get("days"))
			monthStr = list(updateDetailsJSON.Get("months"))
			weekDayStr = list(updateDetailsJSON.Get("weekdays"))
			cronVal = append(cronVal, minStr, hourStr, dayStr, monthStr, weekDayStr)
			cronData := strings.Join(cronVal, " ")
			instance.SetLabel("cron", cronData)
			instance.SetLabel("schedule", cronData)
		}
		if interval := instance.GetLabel("interval"); interval != "" {
			instance.SetLabel("schedule", interval)
		}
		if localClusterName == instance.GetLabel("cluster_name") {
			instance.SetLabel("site", "local")
		} else {
			instance.SetLabel("site", "remote")
		}
	}
	return nil, nil, nil
}

func list(get gjson.Result) string {
	if !get.IsArray() {
		return "*"
	}
	array := get.Array()
	items := make([]string, 0, len(array))
	for _, e := range array {
		items = append(items, e.ClonedString())
	}
	return strings.Join(items, ",")
}
