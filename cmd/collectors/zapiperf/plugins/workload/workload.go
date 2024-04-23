package workload

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/qospolicyfixed"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"regexp"
	"strings"
)

const (
	batchSize = "500"
	token     = "#?"
)

var lunRegex = regexp.MustCompile("^/[^/]+/([^/]+)(?:/.*?|)/([^/]+)$")
var constituentRegex = regexp.MustCompile(`^(.*)__(\d{4})$`)

var metrics = []string{
	"used_ops_percent",
	"used_throughput_percent",
}

type Workload struct {
	*plugin.AbstractPlugin
	client       *zapi.Client
	isClientInit bool
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Workload{AbstractPlugin: p}
}

func (w *Workload) Init() error {
	if err := w.InitAbc(); err != nil {
		return err
	}

	return nil
}

func (w *Workload) initClient() error {
	var err error
	if w.client, err = zapi.New(conf.ZapiPoller(w.ParentParams), w.Auth); err != nil {
		w.Logger.Error().Stack().Err(err).Send()
		return err
	}
	return w.client.Init(5)
}

func (w *Workload) createMetrics(data *matrix.Matrix) error {
	for _, k := range metrics {
		err := matrix.CreateMetric(k, data)
		if err != nil {
			w.Logger.Warn().Err(err).Str("key", k).Msg("error while creating metric")
			return err
		}
	}
	return nil
}

func (w *Workload) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {

	var (
		err error
	)

	// The ZapiPerf test triggers the workload plugin, so the client init has been done to the Run method instead of the plugin init.
	// Ensure client is initialized only once
	if !w.isClientInit {
		if err := w.initClient(); err != nil {
			return nil, nil, fmt.Errorf("failed to initialize ZAPI client: %w", err)
		}
		w.isClientInit = true
	}

	data := dataMap[w.Object]
	w.client.Metadata.Reset()

	err = w.createMetrics(data)
	if err != nil {
		return nil, nil, err
	}

	qospolicyMap, err := w.fetchQOSAdaptive()
	if err != nil {
		return nil, nil, err
	}

	volumeWorkloadMap := make(map[string]string)
	lunWorkloadMap := make(map[string]string)
	volumeNames := set.New()
	svmNames := set.New()
	lunVolumeNames := set.New()
	lunSvmNames := set.New()
	for k, v := range data.GetInstances() {
		volume := v.GetLabel("volume")
		qtree := v.GetLabel("qtree")
		lun := v.GetLabel("lun")
		file := v.GetLabel("file")
		svm := v.GetLabel("svm")

		// ignore constituents as qos is applied to flexgroups as a whole and not at constituent level
		if match := constituentRegex.FindStringSubmatch(volume); len(match) == 3 {
			continue
		}

		// if qtree, lun, file are empty it means that it's a volume workload
		if qtree == "" && lun == "" && file == "" {
			key := volume + token + svm
			volumeWorkloadMap[key] = k
			volumeNames.Add(volume)
			svmNames.Add(svm)
		} else if lun != "" {
			key := volume + token + svm + token + lun
			lunWorkloadMap[key] = k
			lunVolumeNames.Add(volume)
			lunSvmNames.Add(svm)
		}
	}
	params := &VolumeWorkloadParams{
		data:              data,
		qosPolicyMap:      qospolicyMap,
		volumeWorkloadMap: volumeWorkloadMap,
		volumeNames:       volumeNames,
		svmNames:          svmNames,
	}
	err = w.populateUsedIOPsPercentVolume(params)
	if err != nil {
		return nil, nil, err
	}

	lunParams := &LunWorkloadParams{
		data:           data,
		qosPolicyMap:   qospolicyMap,
		lunWorkloadMap: lunWorkloadMap,
		volumeNames:    lunVolumeNames,
		svmNames:       lunSvmNames,
	}
	err = w.populateUsedIOPsPercentLun(lunParams)
	if err != nil {
		return nil, nil, err
	}

	return nil, w.client.Metadata, nil
}

type VolumeWorkloadParams struct {
	data              *matrix.Matrix
	qosPolicyMap      map[string]*collectors.QosAdaptive
	volumeWorkloadMap map[string]string
	volumeNames       *set.Set
	svmNames          *set.Set
}

func (w *Workload) populateUsedIOPsPercentVolume(params *VolumeWorkloadParams) error {
	var (
		result  *node.Node
		volumes []*node.Node
		err     error
	)

	if params.volumeNames.Size() == 0 {
		return nil
	}

	query := "volume-get-iter"
	tag := "initial"
	request := node.NewXMLS(query)
	request.NewChildS("max-records", batchSize)
	desired := node.NewXMLS("desired-attributes")
	volumeAttributes := node.NewXMLS("volume-attributes")
	volumeSpaceAttributes := node.NewXMLS("volume-space-attributes")
	volumeSpaceAttributes.NewChildS("size", "")
	volumeSpaceAttributes.NewChildS("size-used", "")
	volumeAttributes.AddChild(volumeSpaceAttributes)

	volumeIDAttributes := node.NewXMLS("volume-id-attributes")
	volumeIDAttributes.NewChildS("name", "")
	volumeIDAttributes.NewChildS("owning-vserver-name", "")
	volumeAttributes.AddChild(volumeIDAttributes)
	desired.AddChild(volumeAttributes)

	queryAttribute := node.NewXMLS("query")
	volumeQueryAttributes := node.NewXMLS("volume-attributes")
	volumeIDQueryAttributes := node.NewXMLS("volume-id-attributes")
	volumeIDQueryAttributes.NewChildS("name", strings.Join(params.volumeNames.Values(), "|"))
	volumeIDQueryAttributes.NewChildS("owning-vserver-name", strings.Join(params.svmNames.Values(), "|"))
	volumeQueryAttributes.AddChild(volumeIDQueryAttributes)
	queryAttribute.AddChild(volumeQueryAttributes)

	request.AddChild(queryAttribute)
	request.AddChild(desired)

	for {
		if result, tag, err = w.client.InvokeBatchRequest(request, tag, ""); err != nil {
			return err
		}

		if result == nil {
			break
		}

		if x := result.GetChildS("attributes-list"); x != nil {
			volumes = x.GetChildren()
		}
		if len(volumes) == 0 {
			return nil
		}

		for _, v := range volumes {
			size := v.GetChildS("volume-space-attributes").GetChildContentS("size")
			sizeUsed := v.GetChildS("volume-space-attributes").GetChildContentS("size-used")
			volume := v.GetChildS("volume-id-attributes").GetChildContentS("name")
			svm := v.GetChildS("volume-id-attributes").GetChildContentS("owning-vserver-name")
			key := volume + token + svm
			instance := params.data.GetInstance(params.volumeWorkloadMap[key])
			if instance != nil {
				err = w.processWorkloadInstance(instance, size, sizeUsed, params.qosPolicyMap, params.data)
				if err != nil {
					w.Logger.Error().Err(err).Str(key, key).Msg("Failed process workload")
					continue
				}
			}
		}
	}
	return nil
}

type LunWorkloadParams struct {
	data           *matrix.Matrix
	qosPolicyMap   map[string]*collectors.QosAdaptive
	lunWorkloadMap map[string]string
	svmNames       *set.Set
	volumeNames    *set.Set
}

func (w *Workload) populateUsedIOPsPercentLun(params *LunWorkloadParams) error {
	var (
		result *node.Node
		luns   []*node.Node
		err    error
	)

	if params.svmNames.Size() == 0 {
		return nil
	}

	query := "lun-get-iter"
	tag := "initial"
	request := node.NewXMLS(query)
	request.NewChildS("max-records", batchSize)
	desired := node.NewXMLS("desired-attributes")
	lunInfo := node.NewXMLS("lun-info")
	lunInfo.NewChildS("size", "")
	lunInfo.NewChildS("size-used", "")
	lunInfo.NewChildS("volume", "")
	lunInfo.NewChildS("vserver", "")
	lunInfo.NewChildS("path", "")
	desired.AddChild(lunInfo)

	queryAttribute := node.NewXMLS("query")
	lunQueryInfo := node.NewXMLS("lun-info")
	lunQueryInfo.NewChildS("volume", strings.Join(params.volumeNames.Values(), "|"))
	lunQueryInfo.NewChildS("vserver", strings.Join(params.svmNames.Values(), "|"))
	queryAttribute.AddChild(lunQueryInfo)

	request.AddChild(queryAttribute)
	request.AddChild(desired)

	for {
		if result, tag, err = w.client.InvokeBatchRequest(request, tag, ""); err != nil {
			return err
		}

		if result == nil {
			break
		}

		if x := result.GetChildS("attributes-list"); x != nil {
			luns = x.GetChildren()
		}
		if len(luns) == 0 {
			return nil
		}

		for _, l := range luns {
			size := l.GetChildContentS("size")
			sizeUsed := l.GetChildContentS("size-used")
			volume := l.GetChildContentS("volume")
			svm := l.GetChildContentS("vserver")
			path := l.GetChildContentS("path")
			var lun string
			matches := lunRegex.FindStringSubmatch(path)
			if len(matches) > 2 {
				lun = matches[2]
			}
			if lun == "" {
				continue
			}
			key := volume + token + svm + token + lun
			instance := params.data.GetInstance(params.lunWorkloadMap[key])
			if instance != nil {
				err = w.processWorkloadInstance(instance, size, sizeUsed, params.qosPolicyMap, params.data)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (w *Workload) processWorkloadInstance(instance *matrix.Instance, size, sizeUsed string, qosPolicyMap map[string]*collectors.QosAdaptive, data *matrix.Matrix) error {
	var (
		peakAllowedIOPS float64
	)

	policyGroup := instance.GetLabel("policy_group")
	policy, ok := qosPolicyMap[policyGroup]
	if !ok {
		return nil
	}

	sizeTB, err := util.StringBytesToTerabytes(size)
	if err != nil {
		return err
	}
	sizeUsedTB, err := util.StringBytesToTerabytes(sizeUsed)
	if err != nil {
		return err
	}

	peakIOPS, err := util.ConvertStringToFloat64(policy.PeakIOPS)
	if err != nil {
		return err
	}

	peakAllowedIOPS, err = collectors.CalculateIOPS(policy, sizeTB, sizeUsedTB)
	if err != nil {
		return err
	}

	if ops, record := data.GetMetric("ops").GetValueFloat64(instance); record {
		usedPercent := collectors.CalculateUsedPercent(ops, peakAllowedIOPS)
		err = data.GetMetric("used_ops_percent").SetValueFloat64(instance, usedPercent)
		if err != nil {
			return err
		}
	}

	if policy.BlockSize != "any" {
		totalDataBps, record := data.GetMetric("total_data").GetValueFloat64(instance)
		if record {
			// possible values are "4k", "8k", "16k", "32k", "64k", "128k"
			blockSizeTrim := strings.TrimSpace(strings.ReplaceAll(strings.ToUpper(policy.BlockSize), "K", ""))
			blockSize, err := util.ConvertStringToFloat64(blockSizeTrim)
			if err != nil {
				return err
			}
			usedPercentMbps := collectors.CalculateThroughputPercent(totalDataBps, blockSize, peakIOPS)

			err = data.GetMetric("used_throughput_percent").SetValueFloat64(instance, usedPercentMbps)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *Workload) fetchQOSAdaptive() (map[string]*collectors.QosAdaptive, error) {
	var (
		result        []*node.Node
		request       *node.Node
		err           error
		qosadativeMap map[string]*collectors.QosAdaptive
	)

	qosadativeMap = make(map[string]*collectors.QosAdaptive)
	request = node.NewXMLS("qos-adaptive-policy-group-get-iter")
	request.NewChildS("max-records", batchSize)

	// Fetching only admin SVMs
	if result, err = w.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	for _, r := range result {
		policyGroup := r.GetChildContentS("policy-group")
		absoluteMinIOPS := r.GetChildContentS("absolute-min-iops")
		peakIOPS := r.GetChildContentS("peak-iops")
		peakIOPSAllocation := r.GetChildContentS("peak-iops-allocation")
		expectedIOPS := r.GetChildContentS("expected-iops")
		expectedIOPSAllocation := r.GetChildContentS("expected-iops-allocation")
		blockSize := r.GetChildContentS("block-size")
		svm := r.GetChildContentS("vserver")

		if absoluteMinIOPS, err = convertIOPS(absoluteMinIOPS); err != nil {
			return nil, fmt.Errorf("failed to convert absoluteMinIOPS: %w", err)
		}

		if peakIOPS, err = convertIOPS(peakIOPS); err != nil {
			return nil, fmt.Errorf("failed to convert peakIOPS: %w", err)
		}

		if expectedIOPS, err = convertIOPS(expectedIOPS); err != nil {
			return nil, fmt.Errorf("failed to convert expectedIOPS: %w", err)
		}

		qosadativeMap[policyGroup] = &collectors.QosAdaptive{
			PolicyGroup:            policyGroup,
			AbsoluteMinIOPS:        absoluteMinIOPS,
			PeakIOPS:               peakIOPS,
			PeakIOPSAllocation:     peakIOPSAllocation,
			ExpectedIOPS:           expectedIOPS,
			ExpectedIOPSAllocation: expectedIOPSAllocation,
			BlockSize:              blockSize,
			Svm:                    svm,
		}
	}
	return qosadativeMap, nil
}

func convertIOPS(iopsValue string) (string, error) {
	xput, err := qospolicyfixed.ZapiXputToRest(iopsValue)
	if err != nil {
		return "", err
	}
	return xput.IOPS, nil
}
