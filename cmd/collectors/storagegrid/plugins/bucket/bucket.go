package bucket

import (
	"github.com/netapp/harvest/v2/cmd/collectors/storagegrid/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"log/slog"
)

type Bucket struct {
	*plugin.AbstractPlugin
	client *rest.Client
	data   *matrix.Matrix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Bucket{AbstractPlugin: p}
}

func (b *Bucket) Init() error {

	var err error

	if err := b.InitAbc(); err != nil {
		return err
	}

	clientTimeout := b.ParentParams.GetChildContentS("client_timeout")
	if b.client, err = rest.NewClient(b.Options.Poller, clientTimeout, b.Auth); err != nil {
		b.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if err := b.client.Init(5); err != nil {
		return err
	}

	b.data = matrix.New(b.Parent+".Bucket", "bucket", "bucket")
	_, _ = b.data.NewMetricFloat64("objects")
	_, _ = b.data.NewMetricFloat64("bytes")

	return nil
}

func (b *Bucket) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	var (
		instanceKey string
	)

	data := dataMap[b.Object]
	metricToJSON := map[string]string{
		"objects": "objectCount",
		"bytes":   "dataBytes",
	}
	// Purge and reset data
	b.data.PurgeInstances()
	b.data.Reset()
	b.client.Metadata.Reset()

	// Set all global labels from Rest.go if already not exist
	b.data.SetGlobalLabels(data.GetGlobalLabels())

	// request the buckets for each tenant
	for instKey, inst := range data.GetInstances() {
		href := "grid/accounts/" + instKey + "/usage?includeBucketDetail=true"
		var records []gjson.Result
		err := b.client.Fetch(href, &records)
		tenantName := inst.GetLabel("tenant")
		if err != nil {
			b.SLogger.Error("Unable to fetch bucket details",
				slogx.Err(err),
				slog.String("id", instKey),
				slog.String("tenantName", tenantName),
			)
			continue
		}

		for _, record := range records {
			if !record.IsObject() {
				b.SLogger.Warn("Bucket is not object, skipping", slog.String("type", record.Type.String()))
				continue
			}

			bucketsJSON := record.Get("buckets")
			for _, bucketJSON := range bucketsJSON.Array() {
				bucket := bucketJSON.Get("name").String()
				region := bucketJSON.Get("region").String()

				instanceKey = instKey + "#" + bucket
				bucketInstance, err2 := b.data.NewInstance(instanceKey)
				if err2 != nil {
					b.SLogger.Error("Failed to add instance",
						slog.Any("err", err2),
						slog.String("instanceKey", instanceKey),
					)
					break
				}
				b.SLogger.Debug("add instance", slog.String("instanceKey", instanceKey))
				bucketInstance.SetLabel("bucket", bucket)
				bucketInstance.SetLabel("tenant", tenantName)
				bucketInstance.SetLabel("region", region)
				for metricKey, m := range b.data.GetMetrics() {
					jsonKey := metricToJSON[metricKey]
					if value := bucketJSON.Get(jsonKey); value.Exists() {
						if err = m.SetValueString(bucketInstance, value.String()); err != nil {
							b.SLogger.Error(
								"Unable to set float key on metric",
								slogx.Err(err),
								slog.String("key", metricKey),
								slog.String("metric", m.GetName()),
								slog.String("value", value.String()),
							)
						} else {
							b.SLogger.Debug(
								"added",
								slog.String("metricKey", metricKey),
								slog.String("value", value.String()),
							)
						}
					}
				}
			}
		}
	}

	return []*matrix.Matrix{b.data}, b.client.Metadata, nil
}
