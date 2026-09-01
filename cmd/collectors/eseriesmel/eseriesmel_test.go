package eseriesmel

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

const pollerName = "test"

func Params(object string, path string) *node.Node {
	yml := `
schedule:
  - data: 9999h
objects:
  %s: %s
`
	yml = fmt.Sprintf(yml, object, path)
	root, err := tree.LoadYaml([]byte(yml))
	if err != nil {
		panic(err)
	}
	return root
}

func JSONToGson(path string) []gjson.Result {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return gjson.ParseBytes(data).Array()
}

func newEseriesMel() *EseriesMel {
	const object = "Events"
	const path = "events.yaml"
	opts := options.New(options.WithConfPath("../../../conf"))
	opts.Poller = pollerName
	opts.HomePath = "testdata"
	opts.IsTest = true

	ac := collector.New("EseriesMel", object, opts, Params(object, path), nil, conf.Remote{})
	e := EseriesMel{}
	if err := e.Init(ac); err != nil {
		panic(err)
	}
	return &e
}

func TestMain(m *testing.M) {
	conf.TestLoadHarvestConfig("testdata/config.yml")
	os.Exit(m.Run())
}

func TestEseriesMel_Init(t *testing.T) {
	e := newEseriesMel()

	assert.NotNil(t, e.Prop)
	assert.NotNil(t, e.Client)
	assert.NotNil(t, e.Matrix)

	assert.Equal(t, e.Prop.InstanceKeys, []string{"eventType", "location"})
	assert.Equal(t, e.Prop.InstanceLabels["componentType"], "component_type")
	assert.Equal(t, e.Prop.InstanceLabels["priority"], "severity")
	assert.True(t, e.Prop.BatchSize > 0)
	assert.Equal(t, e.Prop.Filter, []string{"includeDebug=false"})
	assert.True(t, len(e.Prop.EventCatalog) > 0)
	assert.Equal(t, e.Prop.EventCatalog["8817"], "MEL_EV_DRIVE_UNSUPPORTED_CAPACITY")

	if e.Prop.Query == "" || !contains(e.Prop.Query, "{array_id}") {
		t.Error("query should contain array_id placeholder")
	}

	mat := e.Matrix[e.Object]
	_, ok := mat.GetMetrics()[eventsMetricName]
	assert.True(t, ok)
}

// TestEseriesMel_ParseTemplate_RequiresEvents locks in that the "events"
// catalog is mandatory -- there's no more "empty = collect everything" mode.
func TestEseriesMel_ParseTemplate_RequiresEvents(t *testing.T) {
	e := newEseriesMel()

	yml := `
object: Events
query: storage-systems/{array_id}/mel-events
`
	root, err := tree.LoadYaml([]byte(yml))
	if err != nil {
		t.Fatalf("LoadYaml failed: %v", err)
	}
	e.Params = root

	err = e.ParseTemplate()
	assert.NotNil(t, err)
}

// TestEseriesMel_ParseTemplate_RequiresInstanceKey locks in that a template
// with no ^^ counter is rejected instead of silently dropping every event.
func TestEseriesMel_ParseTemplate_RequiresInstanceKey(t *testing.T) {
	e := newEseriesMel()
	e.InitProp() // discard the real shipped template's instance keys

	yml := `
object: Events
query: storage-systems/{array_id}/mel-events
events:
  - event_type: 1
    name: MEL_EV_FIRST
counters:
  - ^priority => severity
`
	root, err := tree.LoadYaml([]byte(yml))
	if err != nil {
		t.Fatalf("LoadYaml failed: %v", err)
	}
	e.Params = root

	err = e.ParseTemplate()
	assert.NotNil(t, err)
}

// TestEseriesMel_ParseTemplate_BatchSize covers batch_size validation. It
// feeds count and decides paginate's short-page stop, so a bad value has to
// fail at init rather than quietly fall back to the default.
func TestEseriesMel_ParseTemplate_BatchSize(t *testing.T) {
	tests := []struct {
		name      string
		batchSize string
		wantErr   bool
		want      int
	}{
		{name: "positive value is used", batchSize: "250", want: 250},
		{name: "zero is rejected", batchSize: "0", wantErr: true},
		{name: "negative is rejected", batchSize: "-1", wantErr: true},
		{name: "non-numeric is rejected", batchSize: "many", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEseriesMel()

			yml := `
object: Events
query: storage-systems/{array_id}/mel-events
batch_size: ` + tt.batchSize + `
events:
  - event_type: 1
    name: MEL_EV_FIRST
`
			root, err := tree.LoadYaml([]byte(yml))
			if err != nil {
				t.Fatalf("LoadYaml failed: %v", err)
			}
			e.Params = root

			err = e.ParseTemplate()
			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}
			if err != nil {
				t.Fatalf("ParseTemplate failed: %v", err)
			}
			assert.Equal(t, e.Prop.BatchSize, tt.want)
		})
	}
}

// TestEseriesMel_ParseTemplate_DuplicateEventType locks in that a repeated
// event_type keeps the first-seen name (warns, doesn't error or overwrite).
func TestEseriesMel_ParseTemplate_DuplicateEventType(t *testing.T) {
	e := newEseriesMel()

	yml := `
object: Events
query: storage-systems/{array_id}/mel-events
events:
  - event_type: 1
    name: MEL_EV_FIRST
  - event_type: 1
    name: MEL_EV_SECOND
`
	root, err := tree.LoadYaml([]byte(yml))
	if err != nil {
		t.Fatalf("LoadYaml failed: %v", err)
	}
	e.Params = root

	if err := e.ParseTemplate(); err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	assert.Equal(t, e.Prop.EventCatalog["1"], "MEL_EV_FIRST")
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestEseriesMel_PollData_BoundedKey verifies that repeated occurrences of
// the same (event_type, location) collapse into a single instance/series, and
// that different locations for the same event_type stay distinct.
func TestEseriesMel_PollData_BoundedKey(t *testing.T) {
	e := newEseriesMel()
	e.arrayID = "600a098000f63714000000005e5cf5d2"
	e.arrayName = "eseries-test-system"
	// This test is about key-collapsing, not event_type filtering - swap in a
	// minimal catalog covering the fixture's event types, regardless of the
	// shipped template's real 61-event catalog.
	e.Prop.EventCatalog = map[string]string{
		"20515": "MEL_EV_TEST_DRIVE",
		"8466":  "MEL_EV_TEST_CONTROLLER",
		"9472":  "MEL_EV_TEST_RELATIVE",
	}
	mat := e.Matrix[e.Object]
	mat.SetGlobalLabel("array", e.arrayName)

	events := JSONToGson("testdata/mel-events-page1.json")
	count := e.pollData(mat, events)

	instances := mat.GetInstances()
	// 5 raw events, but eventType=20515 fires twice on the same location and
	// should collapse to 1 instance; eventType=8466 fires on two different
	// controllers/locations and should stay as 2 distinct instances; eventType=9472
	// is its own instance.
	assert.Equal(t, len(instances), 4)
	// count tracks every processed event (metric writes), not unique instances.
	assert.Equal(t, count, uint64(5))

	instanceKeys := set.New()
	for key := range instances {
		if instanceKeys.Has(key) {
			t.Errorf("duplicate instance key found: %s", key)
		}
		instanceKeys.Add(key)
	}

	drive := mat.GetInstance("20515" + keySeparator + "Tray 99, Slot A")
	assert.NotNil(t, drive)
	if drive != nil {
		assert.Equal(t, drive.GetLabels()["component_type"], "drive")
		assert.Equal(t, drive.GetLabels()["severity"], "info")
		assert.Equal(t, drive.GetLabels()["message"], "MEL_EV_TEST_DRIVE")
	}

	ctrl1 := mat.GetInstance("8466" + keySeparator + "Tray 99, Controller 1, Slot 1")
	ctrl2 := mat.GetInstance("8466" + keySeparator + "Tray 99, Controller 2, Slot 1")
	assert.NotNil(t, ctrl1)
	assert.NotNil(t, ctrl2)
	if ctrl1 != nil {
		assert.Equal(t, ctrl1.GetLabels()["severity"], "critical")
		assert.Equal(t, ctrl1.GetLabels()["message"], "MEL_EV_TEST_CONTROLLER")
	}

	// Real-world dominant case: componentType="relative" must resolve through
	// componentLocation.componentRelativeLocation.componentType, not leak the
	// placeholder into the label.
	relative := mat.GetInstance("9472" + keySeparator + "Tray 0, Controller A")
	assert.NotNil(t, relative)
	if relative != nil {
		assert.Equal(t, relative.GetLabels()["component_type"], "controller")
		assert.Equal(t, relative.GetLabels()["message"], "MEL_EV_TEST_RELATIVE")
	}

	metr, ok := mat.GetMetrics()[eventsMetricName]
	assert.True(t, ok)
	if ok && drive != nil {
		value, valOk := metr.GetValueFloat64(drive)
		assert.True(t, valOk)
		// eventType=20515 fires twice on the same (event_type, location) key;
		// the metric holds the latest occurrence's timestamp (1785499403), not
		// a count of the 2 occurrences.
		assert.Equal(t, value, float64(1785499403))
	}
}

func TestEseriesMel_PollData_EventTypesFilter(t *testing.T) {
	e := newEseriesMel()
	e.arrayID = "600a098000f63714000000005e5cf5d2"
	e.arrayName = "eseries-test-system"
	// Only 8466 is in the catalog, so only its 2 occurrences pass the filter.
	e.Prop.EventCatalog = map[string]string{"8466": "MEL_EV_TEST_CONTROLLER"}
	mat := e.Matrix[e.Object]

	events := JSONToGson("testdata/mel-events-page1.json")
	count := e.pollData(mat, events)

	// Only the two eventType=8466 occurrences should pass the catalog filter.
	assert.Equal(t, len(mat.GetInstances()), 2)
	assert.Equal(t, count, uint64(2))
}

// TestEseriesMel_PollData_LabelsMatchWinningTimestamp verifies labels come
// from the same occurrence as the winning (max-timestamp) metric value.
func TestEseriesMel_PollData_LabelsMatchWinningTimestamp(t *testing.T) {
	e := newEseriesMel()
	e.arrayID = "600a098000f63714000000005e5cf5d2"
	e.arrayName = "eseries-test-system"
	e.Prop.EventCatalog = map[string]string{"8466": "MEL_EV_TEST_CONTROLLER"}
	mat := e.Matrix[e.Object]

	// Same key, out of ts order: higher-timestamp occurrence processed first.
	events := gjson.Parse(`[
		{
			"sequenceNumber": "802",
			"timeStamp": "1785504000",
			"eventType": 8466,
			"priority": "priorityCritical",
			"componentType": "controller",
			"location": "Tray 99, Controller 1, Slot 1",
			"description": "Data cache scrub failure"
		},
		{
			"sequenceNumber": "801",
			"timeStamp": "1785503998",
			"eventType": 8466,
			"priority": "priorityInfo",
			"componentType": "controller",
			"location": "Tray 99, Controller 1, Slot 1",
			"description": "Stale occurrence processed second"
		}
	]`).Array()

	count := e.pollData(mat, events)
	assert.Equal(t, count, uint64(2))

	instance := mat.GetInstance("8466" + keySeparator + "Tray 99, Controller 1, Slot 1")
	assert.NotNil(t, instance)
	if instance == nil {
		return
	}
	// Must reflect the winning occurrence, not whichever was processed last.
	assert.Equal(t, instance.GetLabels()["severity"], "critical")

	metr, ok := mat.GetMetrics()[eventsMetricName]
	assert.True(t, ok)
	value, valOk := metr.GetValueFloat64(instance)
	assert.True(t, valOk)
	assert.Equal(t, value, float64(1785504000))
}

func TestEseriesMel_PollData_LocationMissingOrEmpty(t *testing.T) {
	e := newEseriesMel()
	e.arrayID = "600a098000f63714000000005e5cf5d2"
	e.Prop.EventCatalog = map[string]string{"37122": "MEL_EV_HOST_REDUNDANCY_LOST"}
	mat := e.Matrix[e.Object]

	events := gjson.Parse(`[
		{"sequenceNumber":"1","timeStamp":"100","eventType":37122,"priority":"priorityCritical","componentType":"host","location":""},
		{"sequenceNumber":"2","timeStamp":"200","eventType":37122,"priority":"priorityCritical","componentType":"host"}
	]`).Array()

	count := e.pollData(mat, events)
	assert.Equal(t, count, uint64(2))
	assert.Equal(t, len(mat.GetInstances()), 1)
}

func TestEseriesMel_DetectGap(t *testing.T) {
	tests := []struct {
		name           string
		nextSeqNum     int64
		startingSeqNum int64
		wantNextSeqNum int64
	}{
		{name: "no gap", nextSeqNum: 100, startingSeqNum: 0, wantNextSeqNum: 100},
		{name: "no gap, exact boundary", nextSeqNum: 100, startingSeqNum: 100, wantNextSeqNum: 100},
		{name: "gap, cursor realigned", nextSeqNum: 100, startingSeqNum: 150, wantNextSeqNum: 150},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEseriesMel()
			e.nextSeqNum = tt.nextSeqNum

			e.detectGap(melExtent{startingSeqNum: tt.startingSeqNum})

			assert.Equal(t, e.nextSeqNum, tt.wantNextSeqNum)
		})
	}
}

// TestShouldSkipFetch verifies the API-call-skipping optimization: if the true
// (unfiltered) tip hasn't moved since last poll, there's nothing new
// regardless of any server-side filter, so the mel-events fetch can be
// skipped entirely.
func TestShouldSkipFetch(t *testing.T) {
	tests := []struct {
		name         string
		endingSeqNum int64
		nextSeqNum   int64
		want         bool
	}{
		{name: "unchanged - skip", endingSeqNum: 100, nextSeqNum: 100, want: true},
		{name: "advanced - do not skip", endingSeqNum: 105, nextSeqNum: 100, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipFetch(melExtent{endingSeqNum: tt.endingSeqNum}, tt.nextSeqNum)
			assert.Equal(t, got, tt.want)
		})
	}
}

// TestNormalizeManagementVersion verifies major.minor extraction for
// alert-rule variant selection (e.g. "12.10.00.9001" -> "12.10").
func TestNormalizeManagementVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{name: "full version string", version: "12.10.00.9001", want: "12.10"},
		{name: "already major.minor", version: "11.90", want: "11.90"},
		{name: "single segment, unchanged", version: "12", want: "12"},
		{name: "empty, unchanged", version: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, normalizeManagementVersion(tt.version), tt.want)
		})
	}
}

// TestDedupedCount verifies the "mel events deduped" log line's math: how many
// raw occurrences collapsed into an already-existing instance this poll.
func TestDedupedCount(t *testing.T) {
	tests := []struct {
		name            string
		processed       uint64
		uniqueInstances int
		want            uint64
	}{
		{name: "heavy duplication, ratio taken from a live array", processed: 105, uniqueInstances: 31, want: 74},
		{name: "no duplicates", processed: 5, uniqueInstances: 5, want: 0},
		{name: "nothing processed", processed: 0, uniqueInstances: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, dedupedCount(tt.processed, tt.uniqueInstances), tt.want)
		})
	}
}

// TestRolledBack verifies rollback fires only when the tip moves backward
// from the previously observed tip, not when compared against the cursor.
func TestRolledBack(t *testing.T) {
	tests := []struct {
		name    string
		prevTip int64
		extent  melExtent
		want    bool
	}{
		{
			name:    "steady state, caught up",
			prevTip: 804,
			extent:  melExtent{startingSeqNum: 0, endingSeqNum: 805},
			want:    false,
		},
		{
			name:    "new events arrived, not a rollback",
			prevTip: 804,
			extent:  melExtent{startingSeqNum: 0, endingSeqNum: 900},
			want:    false,
		},
		// Regression: quiet poll, nothing new since last time - endingSeqNum
		// legitimately equals the previous tip, must not be mistaken for a rollback.
		{
			name:    "quiet poll, tip unchanged, not a rollback",
			prevTip: 32606,
			extent:  melExtent{startingSeqNum: 16222, endingSeqNum: 32606},
			want:    false,
		},
		{
			name:    "forward gap (ring-buffer purge), not a rollback",
			prevTip: 100,
			extent:  melExtent{startingSeqNum: 150, endingSeqNum: 900},
			want:    false,
		},
		{
			name:    "counter reset to near zero",
			prevTip: 810,
			extent:  melExtent{startingSeqNum: 0, endingSeqNum: 5},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, rolledBack(tt.prevTip, tt.extent), tt.want)
		})
	}
}

// TestPaginate verifies pagination stops once it reaches endSeqExclusive, on a
// short page, or on a non-advancing page
func TestPaginate(t *testing.T) {
	// Nothing to fetch: nextSeq already at or past endSeqExclusive.
	t.Run("nothing to fetch, zero-length range", func(t *testing.T) {
		fetchPage := func(_ int64) ([]gjson.Result, error) {
			t.Fatal("fetchPage should not be called for a zero-length range")
			return nil, nil
		}

		run, err := paginate(500, 500, 2, fetchPage)
		assert.Nil(t, err)
		assert.Equal(t, len(run.events), 0)
		assert.Equal(t, run.nextSeq, int64(500))
		assert.Equal(t, run.reason, stopCovered)
	})

	// An empty page is a short page with nothing to keep. It must leave the
	// cursor untouched and must not be mistaken for a non-advancing page.
	t.Run("stops on empty page", func(t *testing.T) {
		calls := 0
		fetchPage := func(_ int64) ([]gjson.Result, error) {
			calls++
			return nil, nil
		}

		run, err := paginate(101, 201, 2, fetchPage)
		assert.Nil(t, err)
		assert.Equal(t, len(run.events), 0)
		assert.Equal(t, calls, 1)
		assert.Equal(t, run.nextSeq, int64(101))
		assert.Equal(t, run.reason, stopShortPage)
	})

	// A full page that exactly reaches endSeqExclusive must not be mistaken
	// for a short page just because it happened to be the last one needed.
	t.Run("stops once endSeqExclusive is reached", func(t *testing.T) {
		page := gjson.Parse(`[{"sequenceNumber":"101"},{"sequenceNumber":"102"}]`).Array()
		calls := 0
		fetchPage := func(start int64) ([]gjson.Result, error) {
			calls++
			assert.Equal(t, start, int64(101))
			return page, nil
		}

		run, err := paginate(101, 103, 2, fetchPage)
		assert.Nil(t, err)
		assert.Equal(t, len(run.events), 2)
		assert.Equal(t, calls, 1)
		assert.Equal(t, run.nextSeq, int64(103))
		assert.Equal(t, run.reason, stopCovered)
	})

	// Sequence numbers within a page aren't guaranteed to arrive in order;
	// the scan must find the true max regardless of position in the array.
	t.Run("page with out-of-order sequence numbers", func(t *testing.T) {
		page := gjson.Parse(`[{"sequenceNumber":"102"},{"sequenceNumber":"101"}]`).Array()
		calls := 0
		fetchPage := func(start int64) ([]gjson.Result, error) {
			calls++
			assert.Equal(t, start, int64(101))
			return page, nil
		}

		run, err := paginate(101, 103, 2, fetchPage)
		assert.Nil(t, err)
		assert.Equal(t, len(run.events), 2)
		assert.Equal(t, calls, 1)
		assert.Equal(t, run.nextSeq, int64(103))
		assert.Equal(t, run.reason, stopCovered)
	})

	t.Run("keeps fetching past a full page until the array is exhausted", func(t *testing.T) {
		fullPage := gjson.Parse(`[{"sequenceNumber":"101"},{"sequenceNumber":"102"}]`).Array()
		var emptyPage []gjson.Result
		calls := 0
		fetchPage := func(start int64) ([]gjson.Result, error) {
			calls++
			if calls == 1 {
				assert.Equal(t, start, int64(101))
				return fullPage, nil
			}
			assert.Equal(t, start, int64(103))
			return emptyPage, nil
		}

		run, err := paginate(101, 106, 2, fetchPage)
		assert.Nil(t, err)
		assert.Equal(t, len(run.events), 2)
		assert.Equal(t, calls, 2)
		assert.Equal(t, run.nextSeq, int64(103))
		assert.Equal(t, run.reason, stopShortPage)
	})

	// Regression: a mid-pagination error must discard everything, including
	// page 1's already-fetched events - the caller retries the whole range.
	t.Run("error on later page discards all events", func(t *testing.T) {
		fullPage := gjson.Parse(`[{"sequenceNumber":"101"},{"sequenceNumber":"102"}]`).Array()
		calls := 0
		fetchPage := func(start int64) ([]gjson.Result, error) {
			calls++
			if calls == 1 {
				assert.Equal(t, start, int64(101))
				return fullPage, nil
			}
			assert.Equal(t, start, int64(103))
			return nil, errors.New("transient failure")
		}

		run, err := paginate(101, 106, 2, fetchPage)
		assert.NotNil(t, err)
		assert.Equal(t, len(run.events), 0)
		assert.Equal(t, calls, 2)
	})

	// Regression: a page whose sequenceNumber values never advance past
	// nextSeq (missing/stale field) must not be re-fetched forever.
	t.Run("stops on a non-advancing page", func(t *testing.T) {
		stalePage := gjson.Parse(`[{"sequenceNumber":"50"},{"sequenceNumber":"60"}]`).Array()
		calls := 0
		fetchPage := func(start int64) ([]gjson.Result, error) {
			calls++
			assert.Equal(t, start, int64(101))
			return stalePage, nil
		}

		run, err := paginate(101, 201, 2, fetchPage)
		assert.Nil(t, err)
		assert.Equal(t, len(run.events), 0)
		assert.Equal(t, calls, 1)
		assert.Equal(t, run.nextSeq, int64(101))
		assert.Equal(t, run.reason, stopStalled)
	})

	// count caps records returned, not records scanned, so fewer records than
	// count means the array is out of matches and paging must not ask again.
	// Server-side filters make this the ordinary way a poll ends: the 140
	// sequence numbers below the tip yield only 30 records, topping out short
	// of endSeqExclusive.
	t.Run("stops on a short page below endSeqExclusive", func(t *testing.T) {
		seqs := make([]int64, 0, 30)
		for seq := int64(200); len(seqs) < 29; seq += 4 {
			seqs = append(seqs, seq)
		}
		seqs = append(seqs, 337)

		var body strings.Builder
		body.WriteString("[")
		for i, seq := range seqs {
			if i > 0 {
				body.WriteString(",")
			}
			fmt.Fprintf(&body, `{"sequenceNumber":"%d"}`, seq)
		}
		body.WriteString("]")
		page := gjson.Parse(body.String()).Array()

		calls := 0
		fetchPage := func(start int64) ([]gjson.Result, error) {
			calls++
			assert.Equal(t, start, int64(200))
			return page, nil
		}

		run, err := paginate(200, 340, 1000, fetchPage)
		assert.Nil(t, err)
		assert.Equal(t, calls, 1)
		assert.Equal(t, len(run.events), 30)
		assert.Equal(t, run.nextSeq, int64(338))
		assert.Equal(t, run.reason, stopShortPage)

		extent := melExtent{startingSeqNum: 100, endingSeqNum: 340}
		assert.Equal(t, advanceCursor(extent, run), int64(340))
	})

	// The short page breaks after its records are appended, unlike the empty
	// and non-advancing pages, which discard theirs.
	t.Run("short page keeps its events", func(t *testing.T) {
		full := gjson.Parse(`[{"sequenceNumber":"101"},{"sequenceNumber":"102"}]`).Array()
		short := gjson.Parse(`[{"sequenceNumber":"103"}]`).Array()
		calls := 0
		fetchPage := func(start int64) ([]gjson.Result, error) {
			calls++
			if calls == 1 {
				assert.Equal(t, start, int64(101))
				return full, nil
			}
			assert.Equal(t, start, int64(103))
			return short, nil
		}

		run, err := paginate(101, 201, 2, fetchPage)
		assert.Nil(t, err)
		assert.Equal(t, calls, 2)
		assert.Equal(t, len(run.events), 3)
		assert.Equal(t, run.nextSeq, int64(104))
		assert.Equal(t, run.reason, stopShortPage)
	})
}

// TestAdvanceCursor covers where the next poll resumes for each way a
// pagination run can end.
func TestAdvanceCursor(t *testing.T) {
	extent := melExtent{startingSeqNum: 100, endingSeqNum: 900}

	tests := []struct {
		name string
		run  pageRun
		want int64
	}{
		{
			name: "covered, cursor moves to the tip",
			run:  pageRun{nextSeq: 900, reason: stopCovered},
			want: 900,
		},
		{
			name: "records arrived while paging, cursor leads the tip",
			run:  pageRun{nextSeq: 951, reason: stopCovered},
			want: 951,
		},
		{
			name: "short page before the tip, nothing retrievable left",
			run:  pageRun{nextSeq: 401, reason: stopShortPage},
			want: 900,
		},
		{
			name: "stalled, skip the unreadable range rather than wedge",
			run:  pageRun{nextSeq: 401, reason: stopStalled},
			want: 900,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, advanceCursor(extent, tt.run), tt.want)
		})
	}
}

func TestInstanceKey_Separator(t *testing.T) {
	e := newEseriesMel()
	e.Prop.InstanceKeys = []string{"eventType", "location"}

	key1, missing1 := e.instanceKey(gjson.Parse(`{"eventType":"10","location":"5 foo"}`))
	key2, missing2 := e.instanceKey(gjson.Parse(`{"eventType":"1","location":"05 foo"}`))

	assert.NotEqual(t, key1, key2)
	assert.Equal(t, key1, "10"+keySeparator+"5 foo")
	assert.Equal(t, key2, "1"+keySeparator+"05 foo")
	assert.Equal(t, len(missing1), 0)
	assert.Equal(t, len(missing2), 0)

	missingBothKey, missingBoth := e.instanceKey(gjson.Parse(`{}`))
	assert.Equal(t, missingBothKey, keySeparator)
	assert.Equal(t, missingBoth, []string{"eventType", "location"})

	missingSecondKey, missingSecond := e.instanceKey(gjson.Parse(`{"eventType":"10"}`))
	assert.Equal(t, missingSecondKey, "10"+keySeparator)
	assert.Equal(t, missingSecond, []string{"location"})
}

// TestResolvedComponentType verifies the "relative" discriminator resolves to
// the real nested component type instead of the meaningless placeholder.
func TestResolvedComponentType(t *testing.T) {
	// non-"relative" flat value passes through unchanged.
	drive := resolvedComponentType(gjson.Parse(`{"componentType":"drive"}`))
	assert.Equal(t, drive.String(), "drive")

	// "relative" resolves to the nested componentRelativeLocation.componentType.
	relative := resolvedComponentType(gjson.Parse(`{
		"componentType":"relative",
		"componentLocation":{"componentRelativeLocation":{"componentType":"controller"}}
	}`))
	assert.Equal(t, relative.String(), "controller")

	// "relative" with no nested type falls back to the flat "relative" value.
	relativeNoNested := resolvedComponentType(gjson.Parse(`{"componentType":"relative"}`))
	assert.Equal(t, relativeNoNested.String(), "relative")

	// componentType field entirely absent.
	missing := resolvedComponentType(gjson.Parse(`{}`))
	assert.False(t, missing.Exists())
}

// TestNormalizeComponentType verifies the __UNDEFINED sentinel is displayed the
// same way it is for priority, and that real enum values pass through.
func TestNormalizeComponentType(t *testing.T) {
	assert.Equal(t, normalizeComponentType("__UNDEFINED"), "unknown")
	assert.Equal(t, normalizeComponentType("relative"), "unknown")
	assert.Equal(t, normalizeComponentType("drive"), "drive")
	assert.Equal(t, normalizeComponentType(""), "")
}

// TestNormalizePriority verifies the raw priority enum (swagger
// MelEntryEx.priority) maps to a cleaner display string, with unknown/future
// values passing through unchanged rather than being dropped.
func TestNormalizePriority(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{"priorityDefault", "default"},
		{"priorityCritical", "critical"},
		{"priorityInfo", "info"},
		{"priorityEmergency", "emergency"},
		{"priorityAlert", "alert"},
		{"priorityError", "error"},
		{"priorityWarning", "warning"},
		{"priorityNotice", "notice"},
		{"priorityDebug", "debug"},
		{"__UNDEFINED", "unknown"},
		{"somethingNew", "somethingNew"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			assert.Equal(t, normalizePriority(tt.raw), tt.want)
		})
	}
}
