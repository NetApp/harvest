package bucket

import (
	"github.com/netapp/harvest/v2/cmd/collectors/storagegrid/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
)

var metricToJSON = map[string]string{
	"objects":     "objectCount",
	"bytes":       "dataBytes",
	"quota_bytes": "quotaObjectBytes",
	"labels":      "labels",
}

type Bucket struct {
	*plugin.AbstractPlugin
	client *rest.Client
	data   *matrix.Matrix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Bucket{AbstractPlugin: p}
}

func (b *Bucket) Init(remote conf.Remote) error {

	var err error

	if err := b.InitAbc(); err != nil {
		return err
	}

	clientTimeout := b.ParentParams.GetChildContentS("client_timeout")
	if b.client, err = rest.NewClient(b.Options.Poller, clientTimeout, b.Auth); err != nil {
		b.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if err := b.client.Init(5, remote); err != nil {
		return err
	}

	b.data = matrix.New(b.Parent+".Bucket", "bucket", "bucket")
	for k := range metricToJSON {
		_, _ = b.data.NewMetricFloat64(k)
	}
	return nil
}

func (b *Bucket) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	var (
		instanceKey string
	)

	data := dataMap[b.Object]
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
				bucket := bucketJSON.Get("name").ClonedString()
				region := bucketJSON.Get("region").ClonedString()
				versioningEnabled := bucketJSON.Get("versioningEnabled").ClonedString()

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
					if metricKey == "labels" {
						bucketInstance.SetLabel("versioningEnabled", versioningEnabled)
						bucketInstance.SetLabel("tenantID", instKey)
						m.SetValueFloat64(bucketInstance, 1.0)
						continue
					}
					jsonKey := metricToJSON[metricKey]
					if value := bucketJSON.Get(jsonKey); value.Exists() {
						valueStr := value.ClonedString()
						if valueStr == "" {
							continue
						}
						if err = m.SetValueString(bucketInstance, valueStr); err != nil {
							b.SLogger.Error(
								"Unable to set float key on metric",
								slogx.Err(err),
								slog.String("key", metricKey),
								slog.String("metric", m.GetName()),
								slog.String("value", valueStr),
							)
						} else {
							b.SLogger.Debug(
								"added",
								slog.String("metricKey", metricKey),
								slog.String("value", valueStr),
							)
						}
					}
				}
			}
		}
	}

	return []*matrix.Matrix{b.data}, b.client.Metadata, nil
}
