package certificate

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
	"strings"
)

type Certificate struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Certificate{AbstractPlugin: p}
}

func (my *Certificate) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {

	var (
		dateTimeVal string
		gmt         string
	)
	data := dataMap[my.Object]

	// update certificate instance based on admin vaserver serial
	for _, certificateInstance := range data.GetInstances() {
		expiryTime := certificateInstance.GetLabel("expiry_time")
		// parsing 2027-08-10T23:33:22+12:00 and converting to Date:2027-08-10 23:33:22 GMT:+12:00
		fields := strings.Split(expiryTime, "T")
		if strings.Contains(fields[1], "+") {
			f := strings.Split(fields[1], "+")
			dateTimeVal = fields[0] + " " + f[0]
			gmt = "+" + f[1]
		} else if strings.Contains(fields[1], "-") {
			f := strings.Split(fields[1], "-")
			dateTimeVal = fields[0] + " " + f[0]
			gmt = "-" + f[1]
		}
		certificateInstance.SetLabel("dateTime", dateTimeVal)
		certificateInstance.SetLabel("gmt", gmt)
	}

	return nil, nil, nil
}
