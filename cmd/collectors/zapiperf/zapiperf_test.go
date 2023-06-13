package zapiperf

import (
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"testing"
)

var data map[string]*matrix.Matrix

func Test_ZapiPerf(t *testing.T) {
	data = make(map[string]*matrix.Matrix)
	// Initialize the ZapiPerf collector
	z := NewZapiperf()

	// Case1: pollInstance has 5 records and pollData has 5 records, expected exported instances are 5
	expectedInstances := 5
	z.testPollInstanceAndData(t, "testdata/pollInstance1.xml", "testdata/pollData1.xml", expectedInstances)

	// Case2: pollInstance has 6 records and pollData has 7 records, expected exported instances are 6
	expectedInstances = 6
	z.testPollInstanceAndData(t, "testdata/pollInstance2.xml", "testdata/pollData2.xml", expectedInstances)

	// Case3: pollInstance has 5 records and pollData has 3 records, expected exported instances are 3
	expectedInstances = 3
	z.testPollInstanceAndData(t, "testdata/pollInstance3.xml", "testdata/pollData3.xml", expectedInstances)
}

func NewZapiperf() *ZapiPerf {
	// homepath is harvest directory level
	homePath := "../../../"
	zapiperfConfigPath := homePath + "conf/zapiperf/default.yaml"
	zapiperfPoller := "testZapiperf"

	conf.TestLoadHarvestConfig("testdata/config.yml")
	opts := options.Options{
		Poller:   zapiperfPoller,
		HomePath: homePath,
		IsTest:   true,
	}
	ac := collector.New("Zapiperf", "Zapiperf", &opts, readFile(zapiperfConfigPath, "yml"), nil)
	z := &ZapiPerf{}
	if err := z.Init(ac); err != nil {
		log.Fatal().Err(err)
	}

	// Test for volume object
	z.Object = "Volume"
	z.instanceKey = "name"
	mx := matrix.New(z.Object, z.Object, z.Object)
	z.Matrix = make(map[string]*matrix.Matrix)
	z.Matrix[z.Object] = mx
	return z
}

func readFile(path string, kind string) *node.Node {
	var root *node.Node
	bytes, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	if kind == "xml" {
		root, err = tree.LoadXML(bytes)
	} else if kind == "yml" {
		root, err = tree.LoadYaml(bytes)
	}
	if err != nil {
		panic(err)
	}
	return root
}

func (z *ZapiPerf) testPollData(pollDataFile string) map[string]*matrix.Matrix {
	var (
		err   error
		skips int
	)

	prevMat := z.Matrix[z.Object]
	// clone matrix without numeric data and non-exportable all instances
	curMat := prevMat.CloneWithNonExportableInstances(false, true, true)
	curMat.Reset()

	// for updating metadata
	count := uint64(0)

	response := readFile(pollDataFile, "xml")

	// fetch instances
	instances := response.GetChildS("instances")
	if instances == nil || len(instances.GetChildren()) == 0 {
		return nil
	}
	z.Logger.Debug().
		Int("instances", len(instances.GetChildren())).
		Msg("Fetched instances")

	for instIndex, i := range instances.GetChildren() {
		key := i.GetChildContentS(z.instanceKey)
		if key == "" {
			z.Logger.Debug().
				Str("instanceKey", z.instanceKey).
				Str("name", i.GetChildContentS("name")).
				Str("uuid", i.GetChildContentS("uuid")).
				Msg("Skip instance, key is empty")
			continue
		}

		instance := curMat.GetInstance(key)
		if instance == nil {
			z.Logger.Debug().
				Str("key", key).
				Msg("Skip instance key, not found in cache")
			continue
		}

		// Set instance to exportable
		instance.SetExportable(true)
		counters := i.GetChildS("counters")
		if counters == nil {
			z.Logger.Debug().
				Str("key", key).
				Msg("Skip instance key, no data counters")
			continue
		}

		z.Logger.Trace().
			Str("key", key).
			Msg("Fetching data of instance")

		for _, cnt := range counters.GetChildren() {
			name := cnt.GetChildContentS("name")
			value := cnt.GetChildContentS("value")

			// sanity check
			if name == "" || value == "" {
				z.Logger.Debug().
					Str("counter", name).
					Str("value", value).
					Msg("Skipping incomplete counter")
				continue
			}

			// store as instance label
			if display, has := z.instanceLabels[name]; has {
				instance.SetLabel(display, value)
				z.Logger.Trace().
					Str("display", display).
					Int("instIndex", instIndex).
					Str("value", value).
					Msg("SetLabel")
				continue
			}

			// store as array counter / histogram
			if labels, has := z.histogramLabels[name]; has {

				values := strings.Split(value, ",")

				if len(labels) != len(values) {
					// warn & skip
					z.Logger.Error().
						Stack().
						Str("labels", name).
						Str("value", value).
						Int("instIndex", instIndex).
						Msg("Histogram labels don't match parsed values")
					continue
				}

				for i, label := range labels {
					if metric := curMat.GetMetric(name + "." + label); metric != nil {
						if err = metric.SetValueString(instance, values[i]); err != nil {
							z.Logger.Error().
								Stack().
								Err(err).
								Str("name", name).
								Str("label", label).
								Str("value", values[i]).
								Int("instIndex", instIndex).
								Msg("Set histogram value failed")
						} else {
							z.Logger.Trace().
								Str("name", name).
								Str("label", label).
								Str("value", values[i]).
								Int("instIndex", instIndex).
								Msg("Set histogram name.label = value")
							count++
						}
					} else {
						z.Logger.Warn().
							Str("name", name).
							Str("label", label).
							Str("value", value).
							Int("instIndex", instIndex).
							Msg("Histogram name. Label not in cache")
					}
				}
				continue
			}

			// store as scalar metric
			if metric := curMat.GetMetric(name); metric != nil {
				if err = metric.SetValueString(instance, value); err != nil {
					z.Logger.Error().
						Err(err).
						Str("name", name).
						Str("value", value).
						Int("instIndex", instIndex).
						Msg("Set metric failed")
				} else {
					z.Logger.Trace().
						Int("instIndex", instIndex).
						Str("key", key).
						Str("counter", name).
						Str("value", value).
						Msg("Set metric")
					count++
				}
				continue
			}

			z.Logger.Warn().
				Int("instIndex", instIndex).
				Str("counter", name).
				Str("value", value).
				Msg("Counter not found in cache")
		} // end loop over counters
	} // end loop over instances

	z.Logger.Trace().
		Uint64("count", count).
		Msg("Collected data points in polls")

	// skip calculating from delta if no data from previous poll
	if z.isCacheEmpty {
		z.Logger.Debug().Msg("skip postprocessing until next poll (previous cache empty)")
		z.Matrix[z.Object] = curMat
		z.isCacheEmpty = false
		return nil
	}

	z.Logger.Debug().Msg("starting delta calculations from previous cache")

	// cache raw data for next poll
	cachedData := curMat.Clone(true, true, true) // @TODO implement copy data

	// order metrics, such that those requiring base counters are processed last
	orderedMetrics := make([]*matrix.Metric, 0, len(curMat.GetMetrics()))
	orderedKeys := make([]string, 0, len(orderedMetrics))

	for key, metric := range curMat.GetMetrics() {
		if metric.GetComment() == "" && metric.Buckets() == nil { // does not require base counter
			orderedMetrics = append(orderedMetrics, metric)
			orderedKeys = append(orderedKeys, key)
		}
	}
	for key, metric := range curMat.GetMetrics() {
		if metric.GetComment() != "" && metric.Buckets() == nil { // requires base counter
			orderedMetrics = append(orderedMetrics, metric)
			orderedKeys = append(orderedKeys, key)
		}
	}

	var base *matrix.Metric
	var totalSkips int

	for i, metric := range orderedMetrics {

		property := metric.GetProperty()
		key := orderedKeys[i]

		// RAW - submit without post-processing
		if property == "raw" {
			continue
		}

		// all other properties - first calculate delta
		if skips, err = curMat.Delta(key, prevMat, z.Logger); err != nil {
			z.Logger.Error().Err(err).Str("key", key).Msg("Calculate delta")
			continue
		}
		totalSkips += skips

		// DELTA - subtract previous value from current
		if property == "delta" {
			// already done
			continue
		}

		// RATE - delta, normalized by elapsed time
		if property == "rate" {
			// defer calculation, so we can first calculate averages/percents
			// Note: calculating rate before averages are averages/percentages are calculated
			// used to be a bug in Harvest 2.0 (Alpha, RC1, RC2) resulting in very high latency values
			continue
		}

		// For the next two properties we need base counters
		// We assume that delta of base counters is already calculated
		// (name of base counter is stored as Comment)
		if base = curMat.GetMetric(metric.GetComment()); base == nil {
			z.Logger.Warn().
				Str("key", key).
				Str("property", property).
				Str("comment", metric.GetComment()).
				Msg("Base counter missing")
			continue
		}

		// remaining properties: average and percent
		//
		// AVERAGE - delta, divided by base-counter delta
		//
		// PERCENT - average * 100
		// special case for latency counter: apply minimum number of iops as threshold
		if property == "average" || property == "percent" {

			if strings.HasSuffix(metric.GetName(), "latency") {
				skips, err = curMat.DivideWithThreshold(key, metric.GetComment(), z.latencyIoReqd, z.Logger)
			} else {
				skips, err = curMat.Divide(key, metric.GetComment(), z.Logger)
			}

			if err != nil {
				z.Logger.Error().Err(err).Str("key", key).Msg("Division by base")
			}
			totalSkips += skips

			if property == "average" {
				continue
			}
		}

		if property == "percent" {
			if skips, err = curMat.MultiplyByScalar(key, 100, z.Logger); err != nil {
				z.Logger.Error().Err(err).Str("key", key).Msg("Multiply by scalar")
			} else {
				totalSkips += skips
			}
			continue
		}
		z.Logger.Error().Err(err).
			Str("key", key).
			Str("property", property).
			Msg("Unknown property")
	}

	// calculate rates (which we deferred to calculate averages/percents first)
	for i, metric := range orderedMetrics {
		if metric.GetProperty() == "rate" {
			if skips, err = curMat.Divide(orderedKeys[i], "timestamp", z.Logger); err != nil {
				z.Logger.Error().Err(err).
					Int("i", i).
					Str("key", orderedKeys[i]).
					Msg("Calculate rate")
				continue
			}
			totalSkips += skips
		}
	}

	// store cache for next poll
	z.Matrix[z.Object] = cachedData

	newDataMap := make(map[string]*matrix.Matrix)
	newDataMap[z.Object] = curMat
	return newDataMap
}

func (z *ZapiPerf) testPollInstance(pollInstanceFile string) error {
	var (
		err                                        error
		request, results                           *node.Node
		oldInstances                               *set.Set
		oldSize, newSize, removed, added           int
		keyAttr, instancesAttr, nameAttr, uuidAttr string
	)

	oldInstances = set.New()
	mat := z.Matrix[z.Object]
	for key := range mat.GetInstances() {
		oldInstances.Add(key)
	}
	oldSize = oldInstances.Size()

	z.Logger.Debug().Msgf("updating instance cache (old cache has: %d)", oldInstances.Size())

	nameAttr = "name"
	uuidAttr = "uuid"
	keyAttr = z.instanceKey

	request = node.NewXMLS("perf-object-instance-list-info-iter")
	request.NewChildS("objectname", z.Query)
	instancesAttr = "attributes-list"

	results = readFile(pollInstanceFile, "xml")
	if results == nil {
		return nil
	}

	// fetch instances
	instances := results.GetChildS(instancesAttr)
	if instances == nil || len(instances.GetChildren()) == 0 {
		return nil
	}

	for _, i := range instances.GetChildren() {
		if key := i.GetChildContentS(keyAttr); key == "" {
			// instance key missing
			name := i.GetChildContentS(nameAttr)
			uuid := i.GetChildContentS(uuidAttr)
			z.Logger.Debug().Msgf("skip instance, missing key [%s] (name=%s, uuid=%s)", z.instanceKey, name, uuid)
		} else if oldInstances.Has(key) {
			// instance already in cache
			oldInstances.Remove(key)
			z.Logger.Trace().Msgf("updated instance [%s%s%s%s]", color.Bold, color.Yellow, key, color.End)
			continue
		} else if _, err := mat.NewInstance(key); err != nil {
			z.Logger.Error().Err(err).Msg("add instance")
		} else {
			z.Logger.Trace().
				Str("key", key).
				Msg("Added new instance")
		}
	}

	for key := range oldInstances.Iter() {
		mat.RemoveInstance(key)
		z.Logger.Debug().Msgf("removed instance [%s]", key)
	}

	removed = oldInstances.Size()
	newSize = len(mat.GetInstances())
	added = newSize - (oldSize - removed)

	z.Logger.Info().Msgf("added %d new, removed %d (total instances %d)", added, removed, newSize)

	if newSize == 0 {
		return errs.New(errs.ErrNoInstance, "")
	}

	return err
}

func (z *ZapiPerf) testPollInstanceAndData(t *testing.T, pollInstanceFile, pollDataFile string, exportableExportedInstance int) {
	var err error

	_ = z.testPollInstance(pollInstanceFile)
	if data = z.testPollData(pollDataFile); err != nil {
		t.Fatalf("Failed to fetch poll data %v", err)
	}

	exportableInstance := 0
	for _, instance := range data[z.Object].GetInstances() {
		if instance.IsExportable() {
			exportableInstance++
		}
	}

	if exportableInstance != exportableExportedInstance {
		t.Errorf("Exported instances got= %d, expected: %d", exportableInstance, exportableExportedInstance)
	}
}
