package vscan

import (
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/matrix"
	"strconv"
	"strings"
)

type Vscan struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Vscan{AbstractPlugin: p}
}

func (v *Vscan) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {
	// defaults plugin options
	isPerSvm := true
	isPerScannerNode := false

	for _, instance := range data.GetInstances() {
		ontapName := instance.GetLabel("instance_uuid")
		// colon separated list of fields
		// vs_test4   :    2.2.2.2   :    umeng-aff300-05
		//  svm       :     scanner  :    node
		if split := strings.Split(ontapName, ":"); len(split) >= 3 {
			instance.SetLabel("svm", split[0])
			instance.SetLabel("scanner", split[1])
			instance.SetLabel("node", split[2])
		}
	}

	if s := v.Params.GetChildContentS("metricsPerSVM"); s != "" {
		if parseBool, err := strconv.ParseBool(s); err == nil {
			isPerSvm = parseBool
		}
	}
	if s := v.Params.GetChildContentS("metricsPerScannerNode"); s != "" {
		if parseBool, err := strconv.ParseBool(s); err == nil {
			isPerScannerNode = parseBool
		}
	}
	v.Logger.Debug().
		Bool("isPerSvm", isPerSvm).
		Bool("isPerScannerNode", isPerScannerNode).
		Msg("Vscan options")

	// Four cases to consider
	//| perSVM | perScannerNode | Meaning                       |
	//| :----: | :------------: | :---------------------------- |
	//|   F    |       F        | return metrics unchanged      |
	//|   F    |       T        | aggregate per scanner         |
	//|   T    |       F        | aggregate per svm             |
	//|   T    |       T        | aggregate per scanner and svm |

	if !isPerSvm && !isPerScannerNode {
		return nil, nil
	}
	return nil, nil
}
