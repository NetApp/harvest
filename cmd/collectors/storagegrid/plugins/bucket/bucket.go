package bucket

import (
	"github.com/netapp/harvest/v2/cmd/collectors/storagegrid/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/tidwall/gjson"
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

	if err = b.InitAbc(); err != nil {
		return err
	}

	clientTimeout := b.ParentParams.GetChildContentS("client_timeout")
	if b.client, err = rest.NewClient(b.Options.Poller, clientTimeout); err != nil {
		b.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = b.client.Init(5); err != nil {
		return err
	}

	b.data = matrix.New(b.Parent+".Bucket", "bucket", "bucket")
	_, _ = b.data.NewMetricFloat64("objects")
	_, _ = b.data.NewMetricFloat64("bytes")

	return nil
}

func (b *Bucket) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {
	var (
		instanceKey string
	)

	metricToJSON := map[string]string{
		"objects": "objectCount",
		"bytes":   "dataBytes",
	}
	// Purge and reset data
	b.data.PurgeInstances()
	b.data.Reset()

	// Set all global labels from Rest.go if already not exist
	b.data.SetGlobalLabels(data.GetGlobalLabels())

	// request the buckets for each tenant
	for instKey, inst := range data.GetInstances() {
		href := "grid/accounts/" + instKey + "/usage?includeBucketDetail=true"
		var records []gjson.Result
		err := b.client.Fetch(href, &records)
		tenantName := inst.GetLabel("tenant")
		if err != nil {
			b.Logger.Error().Err(err).
				Str("id", instKey).
				Str("tenantName", tenantName).
				Msg("Unable to fetch bucket details")
			continue
		}

		for _, record := range records {
			if !record.IsObject() {
				b.Logger.Warn().Str("type", record.Type.String()).Msg("Bucket is not object, skipping")
				continue
			}

			bucketsJSON := record.Get("buckets")
			for _, bucketJSON := range bucketsJSON.Array() {
				bucket := bucketJSON.Get("name").String()
				region := bucketJSON.Get("region").String()

				instanceKey = instKey + "#" + bucket
				bucketInstance, err2 := b.data.NewInstance(instanceKey)
				if err2 != nil {
					b.Logger.Error().Err(err).Str("instanceKey", instanceKey).Msg("Failed to add instance")
					break
				}
				b.Logger.Debug().Str("instanceKey", instanceKey).Msg("add instance")
				bucketInstance.SetLabel("bucket", bucket)
				bucketInstance.SetLabel("tenant", tenantName)
				bucketInstance.SetLabel("region", region)
				for metricKey, m := range b.data.GetMetrics() {
					jsonKey := metricToJSON[metricKey]
					if value := bucketJSON.Get(jsonKey); value.Exists() {
						if err = m.SetValueString(bucketInstance, value.String()); err != nil {
							b.Logger.Error().Err(err).
								Str("key", metricKey).
								Str("metric", m.GetName()).
								Str("value", value.String()).
								Msg("Unable to set float key on metric")
						} else {
							b.Logger.Debug().
								Str("metricKey", metricKey).
								Str("value", value.String()).
								Msg("added")
						}
					}
				}
			}
		}
	}

	return []*matrix.Matrix{b.data}, nil
}
