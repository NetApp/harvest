package clusterschedule

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"strconv"
	"strings"
)

type ClusterScheule struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &ClusterScheule{AbstractPlugin: p}
}

func (c *ClusterScheule) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	for _, instance := range dataMap[c.Object].GetInstances() {
		intervalVal := collectors.HandleDuration(instance.GetLabel("interval"))
		instance.SetLabel("interval", strconv.FormatFloat(intervalVal, 'f', -1, 64))

		cron := instance.GetLabel("cron")
		updateDetailsJSON := gjson.Result{Type: gjson.JSON, Raw: cron}
		var cronVal, minStr, hourStr, weekDayStr string

		minStr = list(updateDetailsJSON.Get("minutes"))
		hourStr = list(updateDetailsJSON.Get("hours"))
		weekDayStr = list(updateDetailsJSON.Get("weekdays"))

		if minStr != "" {
			cronVal = cronVal + "minutes: " + "[" + minStr + "] "
		}
		if hourStr != "" {
			cronVal = cronVal + "hours: " + "[" + hourStr + "] "
		}
		if weekDayStr != "" {
			cronVal = cronVal + "weekdays: " + "[" + weekDayStr + "]"
		}
		instance.SetLabel("cron", strings.TrimSpace(cronVal))
	}
	return nil, nil, nil
}

func list(get gjson.Result) string {
	if !get.IsArray() {
		return ""
	}
	array := get.Array()
	items := make([]string, 0, len(array))
	for _, e := range array {
		items = append(items, e.ClonedString())
	}
	return strings.Join(items, ", ")
}
