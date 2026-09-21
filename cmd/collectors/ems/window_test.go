package ems

import (
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

func newWindowEms() *Ems {
	e := NewEms()
	e.maxLookback = defaultMaxLookback
	return e
}

// Only the lower bound is sent to ONTAP: this endpoint rejects a two-sided
// filter on `time` (400, code 262188). toTime is still returned so the
// client-side cutoff in fetchEMSData can apply it.
func TestTimeWindowSendsOnlyLowerBound(t *testing.T) {
	e := newWindowEms()
	clusterTime := time.Now()
	e.lastFilterTime = clusterTime.Add(-time.Minute).Unix()

	filters, fromTime, toTime := e.timeWindow(clusterTime)

	assert.Equal(t, filters, []string{"time=>=" + strconv.FormatInt(fromTime, 10)})
	assert.Equal(t, fromTime, e.lastFilterTime)
	assert.Equal(t, toTime, clusterTime.Add(-emsVisibilityLag).Unix())
}

// Regression test for the failure loop: lastFilterTime only advances on a
// successful poll, so a run of failures used to widen the window without limit
// and each poll became slower than the last. A 8h watermark must not be
// honored.
func TestTimeWindowClampsStaleWatermark(t *testing.T) {
	e := newWindowEms()
	e.maxLookback = 15 * time.Minute
	clusterTime := time.Now()
	e.lastFilterTime = clusterTime.Add(-8 * time.Hour).Unix()

	_, fromTime, _ := e.timeWindow(clusterTime)

	assert.Equal(t, fromTime, clusterTime.Add(-e.maxLookback).Unix())
}

func TestTimeWindowKeepsFreshWatermark(t *testing.T) {
	e := newWindowEms()
	e.maxLookback = 15 * time.Minute
	clusterTime := time.Now()
	e.lastFilterTime = clusterTime.Add(-2 * time.Minute).Unix()

	_, fromTime, _ := e.timeWindow(clusterTime)

	assert.Equal(t, fromTime, e.lastFilterTime)
}

// On the first poll there is no watermark, so the window opens one data
// interval back.
func TestTimeWindowFirstPollUsesDataInterval(t *testing.T) {
	e := newWindowEms()
	e.lastFilterTime = 0
	clusterTime := time.Now()

	_, fromTime, toTime := e.timeWindow(clusterTime)

	dataInterval, err := collectors.GetDataInterval(e.GetParams(), defaultDataPollDuration)
	assert.Nil(t, err)
	assert.Equal(t, fromTime, clusterTime.Add(-dataInterval).Unix())
	assert.True(t, fromTime < toTime)
}

func TestEventTime(t *testing.T) {
	tests := []struct {
		name string
		json string
		want int64
	}{
		{
			name: "rfc3339, the form ONTAP actually returns",
			json: `{"time":"2026-09-17T18:04:00Z"}`,
			want: time.Date(2026, 9, 17, 18, 4, 0, 0, time.UTC).Unix(),
		},
		{
			name: "epoch seconds",
			json: `{"time":1789748168}`,
			want: 1789748168,
		},
		// 0 means "cannot tell", and the caller keeps the record rather than
		// dropping events over an unreadable timestamp.
		{name: "missing is zero", json: `{"index":7}`, want: 0},
		{name: "unparseable is zero", json: `{"time":"not a date"}`, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, eventTime(gjson.Parse(tt.json)), tt.want)
		})
	}
}

// sortRecords replaces the order_by=index asc that used to be in the query,
// which capped ONTAP's pages at 5000 records and so doubled the request count.
func TestSortRecords(t *testing.T) {
	tests := []struct {
		name    string
		records []string
		want    []string
	}{
		{
			name: "orders chronologically",
			records: []string{
				`{"index":1,"time":"2026-09-17T18:04:03Z","message":{"name":"third"}}`,
				`{"index":9,"time":"2026-09-17T18:04:01Z","message":{"name":"first"}}`,
				`{"index":5,"time":"2026-09-17T18:04:02Z","message":{"name":"second"}}`,
			},
			want: []string{"first", "second", "third"},
		},
		{
			// The whole reason the collector orders records: a resolving ems
			// must be processed after the issuing ems it clears.
			name: "resolving ems lands after issuing ems",
			records: []string{
				`{"index":2,"time":"2026-09-17T18:05:00Z","message":{"name":"wafl.vvol.online"}}`,
				`{"index":1,"time":"2026-09-17T18:04:00Z","message":{"name":"wafl.vvol.offline"}}`,
			},
			want: []string{"wafl.vvol.offline", "wafl.vvol.online"},
		},
		{
			// index is a per-node sequence number, so it cannot order events
			// against each other across nodes - a high index on one node may
			// well be older than a low index on another.
			name: "time wins over index",
			records: []string{
				`{"index":9000,"time":"2026-09-17T18:04:00Z","message":{"name":"earlier"}}`,
				`{"index":2,"time":"2026-09-17T18:04:30Z","message":{"name":"later"}}`,
			},
			want: []string{"earlier", "later"},
		},
		{
			name: "index breaks ties on equal timestamps",
			records: []string{
				`{"index":7,"time":"2026-09-17T18:04:00Z","message":{"name":"b"}}`,
				`{"index":3,"time":"2026-09-17T18:04:00Z","message":{"name":"a"}}`,
			},
			want: []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records := make([]gjson.Result, 0, len(tt.records))
			for _, r := range tt.records {
				records = append(records, gjson.Parse(r))
			}

			sortRecords(records)

			got := make([]string, 0, len(records))
			for _, r := range records {
				got = append(got, r.Get("message.name").ClonedString())
			}
			assert.Equal(t, got, tt.want)
		})
	}
}

// getHref is called repeatedly with one filter slice while queries are sized,
// so appending in place would leak one call's name filter into the next.
func TestGetHrefDoesNotMutateFilter(t *testing.T) {
	e := newWindowEms()
	filter := make([]string, 0, 8)
	filter = append(filter, "message.severity=error", "time=>=100")
	before := slices.Clone(filter)

	e.getHref([]string{"a.b", "c.d"}, filter)
	e.getHref([]string{"e.f"}, filter)

	assert.Equal(t, filter, before)
}

func TestGetHrefHasNoOrderBy(t *testing.T) {
	e := newWindowEms()

	href := e.getHref([]string{"a.b"}, []string{"time=>=100"})

	assert.False(t, strings.Contains(href, "order_by"))
}

func TestGetHrefUsesConfiguredBatchSize(t *testing.T) {
	e := newWindowEms()
	e.batchSize = "2000"

	href := e.getHref([]string{"a.b"}, []string{"time=>=100"})

	assert.True(t, strings.Contains(href, "max_records=2000"))
}

// The window boundary is inWindow plus nextWatermark. Together they must cover
// every second exactly once: an overlap re-handles events (which can re-create
// a bookend instance updateMatrix has already cleared), a gap loses them.
// emsVisibilityLag puts this boundary 5s in the past, where events really do
// exist, so neither is hypothetical.
func TestWindowBoundaryCoversEachSecondOnce(t *testing.T) {
	const toTime = int64(1789748168)

	at := func(secs int64) gjson.Result {
		return gjson.Parse(`{"time":` + strconv.FormatInt(secs, 10) + `}`)
	}

	// Last second of this window is kept here...
	assert.True(t, inWindow(at(toTime), toTime))
	// ...and the next window starts after it, so it is not kept twice.
	assert.True(t, nextWatermark(toTime) > toTime)

	// No gap: the first second the next window asks for is the first second
	// this window rejected.
	assert.False(t, inWindow(at(nextWatermark(toTime)), toTime))
	assert.Equal(t, nextWatermark(toTime), toTime+1)
}

func TestInWindow(t *testing.T) {
	const toTime = int64(1789748168)

	tests := []struct {
		name string
		json string
		want bool
	}{
		{name: "before the end", json: `{"time":1789748100}`, want: true},
		{name: "exactly the end", json: `{"time":1789748168}`, want: true},
		{name: "after the end", json: `{"time":1789748169}`, want: false},
		// An unreadable timestamp is kept: dropping events over a field we
		// cannot parse would be worse than handling one twice.
		{name: "unreadable is kept", json: `{"index":7}`, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, inWindow(gjson.Parse(tt.json), toTime), tt.want)
		})
	}
}

func TestInitCacheRejectsBadParams(t *testing.T) {
	tests := []struct {
		name  string
		param string
		value string
		check func(*testing.T, *Ems)
	}{
		{
			name:  "zero batch_size falls back to the default",
			param: "batch_size",
			value: "0",
			check: func(t *testing.T, e *Ems) { assert.Equal(t, e.batchSize, defaultEmsBatchSize) },
		},
		{
			name:  "negative batch_size falls back to the default",
			param: "batch_size",
			value: "-5",
			check: func(t *testing.T, e *Ems) { assert.Equal(t, e.batchSize, defaultEmsBatchSize) },
		},
		{
			name:  "valid batch_size is honored",
			param: "batch_size",
			value: "2000",
			check: func(t *testing.T, e *Ems) { assert.Equal(t, e.batchSize, "2000") },
		},
		{
			// At or below emsVisibilityLag every window comes out empty and the
			// collector silently stops collecting.
			name:  "zero max_lookback falls back to the default",
			param: "max_lookback",
			value: "0",
			check: func(t *testing.T, e *Ems) { assert.Equal(t, e.maxLookback, defaultMaxLookback) },
		},
		{
			name:  "negative max_lookback falls back to the default",
			param: "max_lookback",
			value: "-1m",
			check: func(t *testing.T, e *Ems) { assert.Equal(t, e.maxLookback, defaultMaxLookback) },
		},
		{
			name:  "max_lookback at the visibility lag falls back to the default",
			param: "max_lookback",
			value: emsVisibilityLag.String(),
			check: func(t *testing.T, e *Ems) { assert.Equal(t, e.maxLookback, defaultMaxLookback) },
		},
		{
			name:  "valid max_lookback is honored",
			param: "max_lookback",
			value: "30m",
			check: func(t *testing.T, e *Ems) { assert.Equal(t, e.maxLookback, 30*time.Minute) },
		},
		{
			name:  "negative max_records falls back to the default",
			param: "max_records",
			value: "-1",
			check: func(t *testing.T, e *Ems) { assert.Equal(t, e.maxRecords, defaultMaxRecords) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEms()
			e.Params.NewChildS(tt.param, tt.value)
			assert.Nil(t, e.InitCache())
			tt.check(t, e)
		})
	}
}

func TestHrefQuota(t *testing.T) {
	tests := []struct {
		name                    string
		maxRecords, taken, i, n int
		want                    int
	}{
		{name: "single href gets the whole cap", maxRecords: 20000, taken: 0, i: 0, n: 1, want: 20000},
		{name: "first of three gets a third", maxRecords: 20000, taken: 0, i: 0, n: 3, want: 6666},
		// An href that returned less than its share leaves the rest to those
		// after it, so splitting the cap costs nothing when nothing is flooding.
		{name: "unused allowance rolls forward", maxRecords: 20000, taken: 100, i: 1, n: 3, want: 9950},
		{name: "last href gets all that remains", maxRecords: 20000, taken: 10000, i: 2, n: 3, want: 10000},
		{name: "cap reached", maxRecords: 20000, taken: 20000, i: 1, n: 3, want: 0},
		{name: "cap overshot", maxRecords: 20000, taken: 25000, i: 1, n: 3, want: 0},
		// Never hand out a zero share while the cap has room left, or a tiny
		// max_records would skip hrefs instead of under-serving them.
		{name: "share floors at one", maxRecords: 2, taken: 1, i: 0, n: 5, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, hrefQuota(tt.maxRecords, tt.taken, tt.i, tt.n), tt.want)
		})
	}
}

// The reason the cap is divided: the event names are split across hrefs in a
// stable order, so a first-come allowance lets one flooding name consume
// everything and leave the trailing names permanently unqueried.
func TestHrefQuotaDoesNotStarveLaterHrefs(t *testing.T) {
	const (
		maxRecords = 20000
		hrefs      = 4
	)

	taken := 0
	quotas := make([]int, 0, hrefs)
	for i := range hrefs {
		q := hrefQuota(maxRecords, taken, i, hrefs)
		quotas = append(quotas, q)
		// Worst case: every href floods and takes its whole share.
		taken += q
	}

	for i, q := range quotas {
		if q <= 0 {
			t.Errorf("href %d got quota %d, every href must be able to ask for records", i, q)
		}
	}
	assert.True(t, taken <= maxRecords)
}
