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
		if minutes := updateDetailsJSON.Get("minutes"); minutes.Exists() {
			for _, m := range minutes.Array() {
				minStr = minStr + m.String() + ", "
			}
			minStr = strings.TrimSuffix(minStr, ", ")
		}
		if hours := updateDetailsJSON.Get("hours"); hours.Exists() {
			for _, h := range hours.Array() {
				hourStr = hourStr + h.String() + ", "
			}
			hourStr = strings.TrimSuffix(hourStr, ", ")
		}
		if weekdays := updateDetailsJSON.Get("weekdays"); weekdays.Exists() {
			for _, w := range weekdays.Array() {
				weekDayStr = weekDayStr + w.String() + ", "
			}
			weekDayStr = strings.TrimSuffix(weekDayStr, ", ")
		}

		if minStr != "" {
			cronVal = cronVal + "minutes: " + "[" + minStr + "] "
		}
		if hourStr != "" {
			cronVal = cronVal + "hours: " + "[" + hourStr + "] "
		}
		if weekDayStr != "" {
			cronVal = cronVal + "weekdays: " + "[" + weekDayStr + "]"
		}
		instance.SetLabel("cron", cronVal)
	}
	return nil, nil, nil
}
