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
	return NewEms()
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

// The window is not clamped. lastFilterTime only advances on a successful poll,
// so a run of hard errors leaves the window spanning every interval since the
// last success - and it is honored as-is. ONTAP's retention decides what can
// actually come back for it; only a warning is emitted.
func TestTimeWindowHonorsStaleWatermark(t *testing.T) {
	e := newWindowEms()
	clusterTime := time.Now()
	e.lastFilterTime = clusterTime.Add(-8 * time.Hour).Unix()

	_, fromTime, _ := e.timeWindow(clusterTime)

	assert.Equal(t, fromTime, e.lastFilterTime)
}

func TestTimeWindowKeepsFreshWatermark(t *testing.T) {
	e := newWindowEms()
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
		// A href with many events is not capped at 1/N of the
		// poll cap. Only a floor for the four hrefs after it is held back.
		{name: "first of five keeps all but the reserved floors", maxRecords: 20000, taken: 0, i: 0, n: 5, want: 16000},
		{name: "unused allowance rolls forward", maxRecords: 20000, taken: 500, i: 1, n: 5, want: 16500},
		{name: "last href gets all that remains", maxRecords: 20000, taken: 16000, i: 4, n: 5, want: 4000},
		// Too little left to honor the floor for everyone, so divide rather
		// than clamp up to it - clamping would hand this href all 1000 and
		// leave href 4 unqueried.
		{name: "divides when the floor cannot be honored", maxRecords: 20000, taken: 19000, i: 3, n: 5, want: 500},
		// More hrefs than the cap can floor (20000/1000 = 20). Equal shares
		// keep every href queried.
		{name: "many hrefs fall back to equal shares", maxRecords: 20000, taken: 0, i: 0, n: 25, want: 800},
		{name: "never hands out more than remains", maxRecords: 20000, taken: 19700, i: 3, n: 5, want: 150},
		{name: "cap reached", maxRecords: 20000, taken: 20000, i: 1, n: 5, want: 0},
		{name: "cap overshot", maxRecords: 20000, taken: 25000, i: 1, n: 5, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, hrefQuota(tt.maxRecords, tt.taken, tt.i, tt.n), tt.want)
		})
	}
}

// The reason the cap is held back rather than offered first-come: the event
// names are split across hrefs in a stable order, so a flooding name in an
// early href would otherwise leave the trailing names permanently unqueried.
//
// Walks several href counts, including more hrefs than the cap can give a floor
// to (20000/1000 = 20), which is where a naive floor silently starves the tail.
func TestHrefQuotaDoesNotStarveLaterHrefs(t *testing.T) {
	const maxRecords = 20000

	for _, hrefs := range []int{1, 2, 5, 19, 20, 21, 25, 40} {
		t.Run(strconv.Itoa(hrefs)+" hrefs", func(t *testing.T) {
			taken := 0
			for i := range hrefs {
				// Worst case: every href floods and takes its whole share.
				q := hrefQuota(maxRecords, taken, i, hrefs)
				if q <= 0 {
					t.Fatalf("href %d of %d got quota %d, every href must be able to ask for records",
						i, hrefs, q)
				}
				taken += q
			}
			assert.True(t, taken <= maxRecords)
		})
	}
}

// client_timeout defaults to 1m against a 3m data interval. A poller with a
// shorter interval must not end up reserving more than its whole budget, or
// every page after the first is refused and every poll looks truncated.
func TestPageReservation(t *testing.T) {
	tests := []struct {
		name                  string
		clientTimeout, budget time.Duration
		want                  time.Duration
	}{
		{
			// The shipped defaults: 1m is under half of a 3m interval, so the
			// whole timeout is held back and 2m stays usable.
			name:          "shipped defaults reserve the whole timeout",
			clientTimeout: time.Minute,
			budget:        3 * time.Minute,
			want:          time.Minute,
		},
		{
			// data: 1m with the 1m default. Reserving the whole timeout would
			// leave nothing, so the cap takes over.
			name:          "timeout equal to the budget is capped at half",
			clientTimeout: time.Minute,
			budget:        time.Minute,
			want:          30 * time.Second,
		},
		{
			name:          "timeout well inside the budget is reserved in full",
			clientTimeout: 30 * time.Second,
			budget:        3 * time.Minute,
			want:          30 * time.Second,
		},
		{
			// data: 1m with the 2m default. Without the cap this reserves more
			// than the budget and nothing after the first page ever runs.
			name:          "timeout larger than the budget is capped at half",
			clientTimeout: 2 * time.Minute,
			budget:        time.Minute,
			want:          30 * time.Second,
		},
		{
			name:          "timeout exactly half the budget is reserved in full",
			clientTimeout: 90 * time.Second,
			budget:        3 * time.Minute,
			want:          90 * time.Second,
		},
		{
			// GetTimeout reports 0 when there is no underlying HTTP client,
			// which is the case under Options.IsTest. Degrade to no reservation
			// rather than something negative.
			name:          "zero timeout reserves nothing",
			clientTimeout: 0,
			budget:        3 * time.Minute,
			want:          0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, pageReservation(tt.clientTimeout, tt.budget), tt.want)
		})
	}
}

// conf/ems/default.yaml documents that client_timeout belongs at or below the
// data interval. Nothing enforced it, so a shortened interval produced silently
// truncated polls rather than a warning.
func TestTimeoutExceedsInterval(t *testing.T) {
	tests := []struct {
		name                    string
		clientTimeout, interval time.Duration
		want                    bool
	}{
		{
			name:          "shipped defaults are fine",
			clientTimeout: time.Minute,
			interval:      3 * time.Minute,
			want:          false,
		},
		{
			// Equal is acceptable: a request may fill the interval, not exceed it.
			name:          "equal is acceptable",
			clientTimeout: time.Minute,
			interval:      time.Minute,
			want:          false,
		},
		{
			// data: 1m left with a 2m timeout, the case Rahul asked about.
			name:          "timeout beyond the interval is flagged",
			clientTimeout: 2 * time.Minute,
			interval:      time.Minute,
			want:          true,
		},
		{
			// No HTTP client configured, as under Options.IsTest. Not a
			// misconfiguration, so it must not warn.
			name:          "zero timeout is not a misconfiguration",
			clientTimeout: 0,
			interval:      time.Minute,
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, timeoutExceedsInterval(tt.clientTimeout, tt.interval), tt.want)
		})
	}
}

// The mid-page cap. A page can be far larger than the quota left - batch_size
// is 10000 by default - so the cap has to stop inside the batch. Checking after
// appending the whole page let a poll overshoot max_records by almost a full
// page.
func TestAppendInWindow(t *testing.T) {
	const toTime = int64(1000)

	at := func(secs int64) gjson.Result {
		return gjson.Parse(`{"time":` + strconv.FormatInt(secs, 10) + `}`)
	}
	batchOf := func(n int, secs int64) []gjson.Result {
		b := make([]gjson.Result, 0, n)
		for range n {
			b = append(b, at(secs))
		}
		return b
	}

	tests := []struct {
		name     string
		have     int
		batch    []gjson.Result
		quota    int
		wantLen  int
		wantFull bool
	}{
		{
			// batch_size > quota: the case the bug produced. A 10000-record
			// page against 50 of quota left must keep 50, not 10000.
			name: "page far larger than the quota stops at the cap",
			have: 0, batch: batchOf(10000, 900), quota: 50,
			wantLen: 50, wantFull: true,
		},
		{
			name: "partly consumed quota stops at the cap",
			have: 40, batch: batchOf(100, 900), quota: 50,
			wantLen: 50, wantFull: true,
		},
		{
			name: "batch inside the quota is kept whole",
			have: 0, batch: batchOf(30, 900), quota: 50,
			wantLen: 30, wantFull: false,
		},
		{
			// Landing exactly on the cap must still report full, or the walk
			// fetches another page and discards every record in it. Under the
			// shipped settings max_records is twice batch_size, so this is the
			// normal truncation path rather than a coincidence.
			name: "batch exactly filling the quota reports full",
			have: 0, batch: batchOf(50, 900), quota: 50,
			wantLen: 50, wantFull: true,
		},
		{
			// Out-of-window records belong to the next poll and must not
			// consume quota.
			name:    "records past the window do not count against the quota",
			have:    0,
			batch:   append(append(batchOf(5, 900), batchOf(100, 1001)...), batchOf(5, 950)...),
			quota:   50,
			wantLen: 10, wantFull: false,
		},
		{
			name: "no quota left keeps nothing",
			have: 50, batch: batchOf(10, 900), quota: 50,
			wantLen: 50, wantFull: true,
		},
		{
			name: "empty batch",
			have: 3, batch: nil, quota: 50,
			wantLen: 3, wantFull: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records := batchOf(tt.have, 900)

			got, full := appendInWindow(records, tt.batch, toTime, tt.quota)

			assert.Equal(t, len(got), tt.wantLen)
			assert.Equal(t, full, tt.wantFull)
			assert.True(t, len(got) <= tt.quota)
		})
	}
}
