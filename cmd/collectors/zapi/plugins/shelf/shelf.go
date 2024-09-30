// Package shelf Copyright NetApp Inc, 2021 All rights reserved
package shelf

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
	"strings"
)

const BatchSize = "500"

type Shelf struct {
	*plugin.AbstractPlugin
	data                map[string]*matrix.Matrix
	shelfMetrics        *matrix.Matrix
	instanceKeys        map[string]string
	instanceLabels      map[string]map[string]string
	shelfInstanceKeys   []string
	shelfInstanceLabels []shelfInstanceLabel
	batchSize           string
	client              *zapi.Client
	query               string
}

type shelfInstanceLabel struct {
	label        string
	labelDisplay string
	parent       string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Shelf{AbstractPlugin: p}
}

func (s *Shelf) Init() error {

	var err error

	if err := s.InitAbc(); err != nil {
		return err
	}

	if s.client, err = zapi.New(conf.ZapiPoller(s.ParentParams), s.Auth); err != nil {
		s.SLogger.Error("connecting", slog.Any("err", err))
		return err
	}

	if err := s.client.Init(5); err != nil {
		return err
	}

	if s.client.IsClustered() {
		return nil
	}

	s.query = "storage-shelf-environment-list-info"

	s.SLogger.Debug("plugin connected!")

	// populating shelfMetrics metric shape from template parsing
	s.create7ModeShelfMetrics()

	s.data = make(map[string]*matrix.Matrix)
	s.instanceKeys = make(map[string]string)
	s.instanceLabels = make(map[string]map[string]string)

	objects := s.Params.GetChildS("objects")
	if objects == nil {
		return errs.New(errs.ErrMissingParam, "objects")
	}

	for _, obj := range objects.GetChildren() {

		attribute := obj.GetNameS()
		objectName := strings.ReplaceAll(attribute, "-", "_")

		if x := strings.Split(attribute, "=>"); len(x) == 2 {
			attribute = strings.TrimSpace(x[0])
			objectName = strings.TrimSpace(x[1])
		}

		s.instanceLabels[attribute] = make(map[string]string)

		s.data[attribute] = matrix.New(s.Parent+".Shelf", "shelf_"+objectName, "shelf_"+objectName)
		s.data[attribute].SetGlobalLabel("datacenter", s.ParentParams.GetChildContentS("datacenter"))

		exportOptions := node.NewS("export_options")
		instanceLabels := exportOptions.NewChildS("instance_labels", "")
		instanceKeys := exportOptions.NewChildS("instance_keys", "")
		instanceKeys.NewChildS("", "shelf")
		instanceKeys.NewChildS("", "channel")

		// artificial metric for status of child object of shelf
		_, _ = s.data[attribute].NewMetricUint8("status")

		for _, x := range obj.GetChildren() {

			for _, c := range x.GetAllChildContentS() {

				metricName, display, kind, _ := util.ParseMetric(c)

				switch kind {
				case "key":
					s.instanceKeys[attribute] = metricName
					s.instanceLabels[attribute][metricName] = display
					instanceKeys.NewChildS("", display)
					s.SLogger.Debug(
						"added instance key",
						slog.String("attribute", attribute),
						slog.String("x", x.GetNameS()),
						slog.String("display", display),
					)
				case "label":
					s.instanceLabels[attribute][metricName] = display
					instanceLabels.NewChildS("", display)
					s.SLogger.Debug(
						"added instance label",
						slog.String("attribute", attribute),
						slog.String("x", x.GetNameS()),
						slog.String("display", display),
					)
				case "float":
					_, err := s.data[attribute].NewMetricFloat64(metricName, display)
					if err != nil {
						s.SLogger.Error("add metric", slog.Any("err", err))
						return err
					}
					s.SLogger.Debug(
						"added metric",
						slog.String("attribute", attribute),
						slog.String("x", x.GetNameS()),
						slog.String("display", display),
					)
				}
			}
		}

		s.SLogger.Debug(
			"added object",
			slog.String("attribute", attribute),
			slog.Int("metrics count", len(s.data[attribute].GetMetrics())),
		)

		s.data[attribute].SetExportOptions(exportOptions)
	}

	s.SLogger.Debug("initialized", slog.Int("objects count", len(s.data)))

	// setup batchSize for request
	s.batchSize = BatchSize

	return nil
}

func (s *Shelf) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {

	var (
		err    error
		output []*matrix.Matrix
	)

	data := dataMap[s.Object]
	s.client.Metadata.Reset()

	if s.client.IsClustered() {
		for _, instance := range data.GetInstances() {
			if !instance.IsExportable() {
				continue
			}

			model := instance.GetLabel("model")
			moduleType := instance.GetLabel("module_type")

			isEmbed := collectors.IsEmbedShelf(model, moduleType)
			if isEmbed {
				instance.SetLabel("isEmbedded", "Yes")
			} else {
				instance.SetLabel("isEmbedded", "No")
			}
		}
		return nil, nil, nil
	}

	// 7 mode handling
	for _, instance := range data.GetInstances() {
		if !instance.IsExportable() {
			continue
		}
		instance.SetLabel("shelf", instance.GetLabel("shelf_id"))
	}

	// Set all global labels from zapi.go if already not exist
	for a := range s.instanceLabels {
		s.data[a].SetGlobalLabels(data.GetGlobalLabels())
	}

	request := node.NewXMLS(s.query)

	result, err := s.client.InvokeZapiCall(request)
	if err != nil {
		return nil, nil, err
	}

	output, err = s.handle7Mode(data, result)

	if err != nil {
		return output, nil, err
	}

	s.SLogger.Debug("", slog.Int("Shelves instance count", len(data.GetInstances())))
	output = append(output, data)
	return output, s.client.Metadata, nil
}

func (s *Shelf) handle7Mode(data *matrix.Matrix, result []*node.Node) ([]*matrix.Matrix, error) {
	var (
		shelves  []*node.Node
		channels []*node.Node
		output   []*matrix.Matrix
	)

	// Result would be the zapi response itself with only one record.
	if len(result) != 1 {
		s.SLogger.Debug("no shelves found")
		return output, nil
	}
	// fallback to 7mode
	channels = result[0].SearchChildren([]string{"shelf-environ-channel-info"})

	if len(channels) == 0 {
		s.SLogger.Debug("no channels found")
		return output, nil
	}

	// Purge and reset data
	for _, data1 := range s.data {
		data1.PurgeInstances()
		data1.Reset()
	}

	// reset instance and matrix of shelfMetrics
	s.shelfMetrics.PurgeInstances()
	s.shelfMetrics.Reset()

	// Purge instances and metrics generated from template and updated data metrics and instances from shelfMetrics
	data.PurgeInstances()
	data.PurgeMetrics()
	for metricName, m := range s.shelfMetrics.GetMetrics() {
		_, err := data.NewMetricFloat64(metricName, m.GetName())
		if err != nil {
			s.SLogger.Error("add metric", slog.Any("err", err))
		}
		s.SLogger.Debug("added", slog.String("metric", m.GetName()))
	}

	for _, channel := range channels {
		channelName := channel.GetChildContentS("channel-name")
		shelves = channel.SearchChildren([]string{"shelf-environ-shelf-list", "shelf-environ-shelf-info"})

		if len(shelves) == 0 {
			s.SLogger.Debug("no shelves found", slog.String("channel", channelName))
			continue
		}

		for _, shelf := range shelves {
			uid := shelf.GetChildContentS("shelf-id")
			shelfName := uid // no shelf name in 7mode
			shelfID := uid

			shelfInstanceKey := shelfID + "." + channelName
			// generating new instances from plugin and adding into data
			newShelfInstance, err := data.NewInstance(shelfInstanceKey)
			if err != nil {
				s.SLogger.Error("add shelf instance", slog.Any("err", err))
				return nil, err
			}

			for _, key := range s.shelfInstanceKeys {
				newShelfInstance.SetLabel(key, shelf.GetChildContentS(key))
			}
			for _, shelfLabelData := range s.shelfInstanceLabels {
				if shelfLabelData.parent == "" {
					newShelfInstance.SetLabel(shelfLabelData.labelDisplay, shelf.GetChildContentS(shelfLabelData.label))
				} else {
					child := shelf.GetChildS(shelfLabelData.parent)
					newShelfInstance.SetLabel(shelfLabelData.labelDisplay, child.GetChildContentS(shelfLabelData.label))
				}
			}

			newShelfInstance.SetLabel("channel", channelName)
			newShelfInstance.SetLabel("shelf", newShelfInstance.GetLabel("shelf_id"))

			// populate numeric data
			for metricKey, m := range data.GetMetrics() {
				if value := strings.Split(shelf.GetChildContentS(metricKey), " ")[0]; value != "" {
					if err := m.SetValueString(newShelfInstance, value); err != nil {
						s.SLogger.Debug("failed to parse", slog.String("metricKey", metricKey), slog.String("value", value), slog.Any("err", err))
					} else {
						s.SLogger.Debug("added", slog.String("metricKey", metricKey), slog.String("value", value))
					}
				}
			}

			for attribute, data1 := range s.data {
				statusMetric := data1.GetMetric("status")
				if statusMetric == nil {
					continue
				}

				if s.instanceKeys[attribute] == "" {
					s.SLogger.Warn("no instance keys defined", slog.String("attribute", attribute))
					continue
				}

				objectElem := shelf.GetChildS(attribute)
				if objectElem == nil {
					s.SLogger.Warn("no instances on this system", slog.String("attribute", attribute))
					continue
				}

				s.SLogger.Debug(
					"fetching",
					slog.Int("instances", len(objectElem.GetChildren())),
					slog.String("attribute", attribute),
				)

				for _, obj := range objectElem.GetChildren() {

					if key := obj.GetChildContentS(s.instanceKeys[attribute]); key != "" {
						instanceKey := shelfID + "." + key + "." + channelName
						instance, err := data1.NewInstance(instanceKey)

						if err != nil {
							s.SLogger.Error("add instance", slog.String("attribute", attribute), slog.Any("err", err))
							return nil, err
						}
						s.SLogger.Debug(
							"add instance",
							slog.String("attribute", attribute),
							slog.String("key", key),
							slog.String("shelfID", shelfID),
						)

						for label, labelDisplay := range s.instanceLabels[attribute] {
							if value := obj.GetChildContentS(label); value != "" {
								instance.SetLabel(labelDisplay, value)
							}
						}

						instance.SetLabel("shelf", shelfName)
						instance.SetLabel("shelf_id", shelfID)
						instance.SetLabel("channel", channelName)

						// Each child would have different possible values which is an ugly way to write all of them,
						// so normal value would be mapped to 1 and rest all are mapped to 0.
						if instance.GetLabel("status") == "normal" {
							_ = statusMetric.SetValueInt64(instance, 1)
						} else {
							_ = statusMetric.SetValueInt64(instance, 0)
						}

						// populate numeric data
						for metricKey, m := range data1.GetMetrics() {
							if value := strings.Split(obj.GetChildContentS(metricKey), " ")[0]; value != "" {
								if err := m.SetValueString(instance, value); err != nil {
									s.SLogger.Debug(
										"failed to parse",
										slog.String("metricKey", metricKey),
										slog.String("value", value),
										slog.Any("err", err),
									)
								} else {
									s.SLogger.Debug("added", slog.String("metricKey", metricKey), slog.String("value", value))
								}
							}
						}
					} else {
						s.SLogger.Debug("instance without skipping", slog.String("inst", s.instanceKeys[attribute]))
					}
				}

				output = append(output, data1)
			}
		}
	}
	return output, nil
}

func (s *Shelf) create7ModeShelfMetrics() {
	s.shelfMetrics = matrix.New(s.Parent+".Shelf", "shelf", "shelf")
	s.shelfInstanceKeys = make([]string, 0)
	s.shelfInstanceLabels = []shelfInstanceLabel{}
	shelfExportOptions := node.NewS("export_options")
	shelfInstanceKeys := shelfExportOptions.NewChildS("instance_keys", "")
	shelfInstanceLabels := shelfExportOptions.NewChildS("instance_labels", "")

	if counters := s.ParentParams.GetChildS("counters"); counters != nil {
		if channelInfo := counters.GetChildS("shelf-environ-channel-info"); channelInfo != nil {
			if shelfList := channelInfo.GetChildS("shelf-environ-shelf-list"); shelfList != nil {
				if shelfInfo := shelfList.GetChildS("shelf-environ-shelf-info"); shelfInfo != nil {
					s.parse7ModeTemplate(shelfInfo, shelfInstanceKeys, shelfInstanceLabels, "")
				}
			}
		}
	}

	shelfInstanceKeys.NewChildS("", "channel")
	shelfInstanceKeys.NewChildS("", "shelf")
}

func (s *Shelf) parse7ModeTemplate(shelfInfo *node.Node, shelfInstanceKeys, shelfInstanceLabels *node.Node, parent string) {
	for _, shelfProp := range shelfInfo.GetChildren() {
		if len(shelfProp.GetChildren()) > 0 {
			s.parse7ModeTemplate(shelfInfo.GetChildS(shelfProp.GetNameS()), shelfInstanceKeys, shelfInstanceLabels, shelfProp.GetNameS())
		} else {
			metricName, display, kind, _ := util.ParseMetric(shelfProp.GetContentS())
			switch kind {
			case "key":
				s.shelfInstanceKeys = append(s.shelfInstanceKeys, metricName)
				s.shelfInstanceLabels = append(s.shelfInstanceLabels, shelfInstanceLabel{label: metricName, labelDisplay: display, parent: parent})
				shelfInstanceKeys.NewChildS("", display)
				s.SLogger.Debug("added instance key", slog.String("display", display))
			case "label":
				s.shelfInstanceLabels = append(s.shelfInstanceLabels, shelfInstanceLabel{label: metricName, labelDisplay: display, parent: parent})
				shelfInstanceLabels.NewChildS("", display)
				s.SLogger.Debug("added instance label", slog.String("display", display))
			case "float":
				_, err := s.shelfMetrics.NewMetricFloat64(metricName, display)
				if err != nil {
					s.SLogger.Error("add metric", slog.Any("err", err))
				}
				s.SLogger.Debug("added", slog.String("metric", display))
			}
		}
	}
}
