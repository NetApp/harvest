package matrix

import (
	"fmt"
	"log/slog"
	"math"
	"slices"
	"testing"

	"github.com/netapp/harvest/v2/assert"
)

const (
	replayInterval     = 60.0
	replayLatencyPerOp = 500.0
	// replayDropOps is how far volume STGNETAPP_17_EOF's statistics.iops_raw dropped in #4496
	replayDropOps = 106_500_000.0
	// replayMaxRate is far above the ~2,000 ops/s the replay generates and far below the ~1.777M ops/s spike from #4496
	replayMaxRate = 10_000.0
	// dropStart is the first poll whose counters dropPolls lowers
	dropStart = 2
)

type replayPoll struct {
	ops     float64 // raw cumulative ops
	latency float64 // raw cumulative latency
	ts      float64 // seconds
	missing bool    // ops and latency are not reported
	partial bool    // instance flagged partial, as for ONTAP partial aggregation
}

type cookedPoll struct {
	rate       float64
	latency    float64
	hasRate    bool
	hasLatency bool
}

// replayPolls returns raw polls 60s apart. The ops rate varies around 2,000 ops/s and every op takes replayLatencyPerOp.
func replayPolls(numPolls int) []replayPoll {
	polls := make([]replayPoll, numPolls)
	ops := 5_000_000_000.0
	ts := 1_760_000_000.0
	for i := range polls {
		polls[i] = replayPoll{ops: ops, latency: ops * replayLatencyPerOp, ts: ts}
		ops += (2_000 + 50*float64(i%7)) * replayInterval
		ts += replayInterval
	}
	return polls
}

// pollsFromOps returns raw polls 60s apart with the given cumulative ops
func pollsFromOps(ops ...float64) []replayPoll {
	polls := make([]replayPoll, len(ops))
	for i, v := range ops {
		polls[i] = replayPoll{ops: v, latency: v * replayLatencyPerOp, ts: 1_760_000_000 + replayInterval*float64(i)}
	}
	return polls
}

// lowerPolls returns a copy of polls with the counters of polls [from, to) lowered by ops
func lowerPolls(polls []replayPoll, from, to int, ops float64) []replayPoll {
	lowered := make([]replayPoll, len(polls))
	copy(lowered, polls)
	for i := from; i < to; i++ {
		lowered[i].ops -= ops
		lowered[i].latency -= ops * replayLatencyPerOp
	}
	return lowered
}

// dropPolls lowers the counters of polls [dropStart, dropStart+n) by replayDropOps, like ONTAP 9.17.1P9 did in #4496
func dropPolls(polls []replayPoll, n int) []replayPoll {
	return lowerPolls(polls, dropStart, dropStart+n, replayDropOps)
}

// resetCounters restarts the counters at start from poll `from` on, as after an unflagged reset,
// and lets them climb at the replay rate
func resetCounters(polls []replayPoll, from int, start float64) {
	for i := from; i < len(polls); i++ {
		polls[i].ops = start + float64(i-from)*2_000*replayInterval
		polls[i].latency = polls[i].ops * replayLatencyPerOp
	}
}

func setReplayRaw(t *testing.T, mat *Matrix, p replayPoll) {
	t.Helper()
	instance := mat.GetInstance("vol1")
	instance.SetPartial(p.partial)
	mat.GetMetric("timestamp").SetValueFloat64(instance, p.ts)
	if p.missing {
		return
	}
	mat.GetMetric("total_ops").SetValueFloat64(instance, p.ops)
	mat.GetMetric("avg_latency").SetValueFloat64(instance, p.latency)
}

// replay cooks raw polls the way the perf collectors do: deltas, then latency averages, then rates.
// The first poll only primes the cache, so the result has one entry per later poll.
func replay(t *testing.T, polls []replayPoll) []cookedPoll {
	t.Helper()
	logger := slog.Default()
	prevMat := New("Test", "test", "test")
	for _, key := range []string{"total_ops", "avg_latency", "timestamp"} {
		_, err := prevMat.NewMetricFloat64(key)
		assert.Nil(t, err)
	}
	_, err := prevMat.NewInstance("vol1")
	assert.Nil(t, err)
	setReplayRaw(t, prevMat, polls[0])

	cooked := make([]cookedPoll, 0, len(polls)-1)
	for _, p := range polls[1:] {
		curMat := prevMat.CloneForCollection()
		curMat.Reset()
		setReplayRaw(t, curMat, p)
		cachedData := curMat.Clone()

		for _, key := range []string{"total_ops", "avg_latency", "timestamp"} {
			_, err = curMat.Delta(key, prevMat, cachedData, false, logger)
			assert.Nil(t, err)
		}
		_, err = curMat.DivideWithThreshold("avg_latency", "total_ops", 0, cachedData, prevMat, "timestamp", logger)
		assert.Nil(t, err)
		_, err = curMat.Divide("total_ops", "timestamp")
		assert.Nil(t, err)

		instance := curMat.GetInstance("vol1")
		var c cookedPoll
		c.rate, c.hasRate = curMat.GetMetric("total_ops").GetValueFloat64(instance)
		c.latency, c.hasLatency = curMat.GetMetric("avg_latency").GetValueFloat64(instance)
		cooked = append(cooked, c)

		prevMat = cachedData
	}
	return cooked
}

// averageRate is the true ops rate between polls from and to
func averageRate(polls []replayPoll, from, to int) float64 {
	return (polls[to].ops - polls[from].ops) / (polls[to].ts - polls[from].ts)
}

func assertClose(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-6*math.Max(1, math.Abs(want)) {
		t.Errorf("%s got=%f want=%f", name, got, want)
	}
}

// assertSkippedPolls checks that exactly the polls in skipped export nothing. Every other poll must export
// the true rate between it and the previous poll, and the true latency per op.
func assertSkippedPolls(t *testing.T, cooked []cookedPoll, truth []replayPoll, skipped []int) {
	t.Helper()
	for i := 1; i < len(truth); i++ {
		c := cooked[i-1]
		if slices.Contains(skipped, i) {
			assert.False(t, c.hasRate)
			assert.False(t, c.hasLatency)
			continue
		}
		assert.True(t, c.hasRate)
		assert.True(t, c.hasLatency)
		assertClose(t, fmt.Sprintf("poll %d rate", i), c.rate, averageRate(truth, i-1, i))
		assertClose(t, fmt.Sprintf("poll %d latency", i), c.latency, replayLatencyPerOp)
	}
}

type deltaStep struct {
	value  float64 // exported delta when ok
	ok     bool
	cached *Matrix // what the collector keeps for the next poll
}

// deltaSteps runs Delta on one counter with one raw reading per poll. The first reading primes the cache.
// A NaN reading means the instance is absent from that poll.
func deltaSteps(t *testing.T, allowPartial bool, raws ...float64) []deltaStep {
	t.Helper()
	logger := slog.Default()
	prevMat := New("Test", "test", "test")
	_, err := prevMat.NewMetricFloat64("ops")
	assert.Nil(t, err)
	_, err = prevMat.NewInstance("vol1")
	assert.Nil(t, err)
	prevMat.GetMetric("ops").SetValueFloat64(prevMat.GetInstance("vol1"), raws[0])

	steps := make([]deltaStep, 0, len(raws)-1)
	for _, raw := range raws[1:] {
		curMat := prevMat.CloneForCollection()
		curMat.Reset()
		var v float64
		var ok bool
		if math.IsNaN(raw) {
			curMat.RemoveInstance("vol1")
		} else {
			inst := curMat.GetInstance("vol1")
			if inst == nil {
				inst, err = curMat.NewInstance("vol1")
				assert.Nil(t, err)
			}
			curMat.GetMetric("ops").SetValueFloat64(inst, raw)
		}
		cached := curMat.Clone()
		_, err = curMat.Delta("ops", prevMat, cached, allowPartial, logger)
		assert.Nil(t, err)
		if inst := curMat.GetInstance("vol1"); inst != nil {
			v, ok = curMat.GetMetric("ops").GetValueFloat64(inst)
		}
		steps = append(steps, deltaStep{value: v, ok: ok, cached: cached})
		prevMat = cached
	}
	return steps
}

// TestDelta_TemporaryDrop replays #4496: counters drop by a constant amount for N polls then recover.
// The drop poll and the poll after it are skipped, low-to-low polls in the dip are exported,
// and the restore poll is skipped instead of being exported as a spike.
func TestDelta_TemporaryDrop(t *testing.T) {
	for _, dropLen := range []int{1, 2, 9, 60} {
		t.Run(fmt.Sprintf("%d low polls", dropLen), func(t *testing.T) {
			truth := replayPolls(dropStart + dropLen + 3)
			cooked := replay(t, dropPolls(truth, dropLen))
			assertSkippedPolls(t, cooked, truth, []int{dropStart, dropStart + 1, dropStart + dropLen})
		})
	}
}

// TestDelta_DropWithMissingReading checks that a poll with no reading inside the dip does not forget the
// remembered value, so the restore poll is still skipped.
func TestDelta_DropWithMissingReading(t *testing.T) {
	const (
		dropLen  = 5
		missing  = dropStart + 2
		recovery = dropStart + dropLen
	)
	truth := replayPolls(recovery + 3)
	polls := dropPolls(truth, dropLen)
	polls[missing].missing = true

	cooked := replay(t, polls)
	assertSkippedPolls(t, cooked, truth, []int{dropStart, dropStart + 1, missing, missing + 1, recovery})
}

// TestDelta_CustomerDipPattern replays the dips in a customer's 60s statistics (#4496): on each volume the
// counters stepped down by one constant amount and back up, six times in 90 polls. Five dips are one poll
// long, so the poll after the drop, which Delta already skipped, is the restore. The dip at poll 18 is two
// polls long, so its restore at poll 20 used to be exported as a spike.
func TestDelta_CustomerDipPattern(t *testing.T) {
	truth := replayPolls(50)
	polls := truth
	for _, dip := range []struct{ from, to int }{{2, 3}, {4, 5}, {7, 8}, {12, 13}, {18, 20}, {44, 45}} {
		polls = lowerPolls(polls, dip.from, dip.to, replayDropOps)
	}

	cooked := replay(t, polls)
	assertSkippedPolls(t, cooked, truth, []int{2, 3, 4, 5, 7, 8, 12, 13, 18, 19, 20, 44, 45})
}

// TestDelta_SmallDropInsideDipLeaks records a known limit: a drop inside a dip replaces the remembered value,
// traffic then carries the counter past it, and the real restore is exported as a spike.
func TestDelta_SmallDropInsideDipLeaks(t *testing.T) {
	const (
		dropLen  = 8
		extra    = 5         // first poll of the small extra drop, which lasts until the end of the dip
		extraOps = 200_000.0 // more than one poll of traffic (~132,000), less than two (~267,000)
		recovery = dropStart + dropLen
	)
	truth := replayPolls(recovery + 3)
	polls := lowerPolls(dropPolls(truth, dropLen), extra, recovery, extraOps)

	cooked := replay(t, polls)

	// poll extra+1 is skipped after the extra drop and passes the replaced value, which releases it
	skipped := []int{dropStart, dropStart + 1, extra, extra + 1}
	for i := 1; i < len(polls); i++ {
		c := cooked[i-1]
		switch {
		case slices.Contains(skipped, i):
			assert.False(t, c.hasRate)
		case i == recovery:
			assert.True(t, c.hasRate)
			if c.rate <= replayMaxRate {
				t.Errorf("poll %d rate %f, want a spike above %f", i, c.rate, replayMaxRate)
			}
		default:
			assert.True(t, c.hasRate)
			assertClose(t, fmt.Sprintf("poll %d rate", i), c.rate, averageRate(truth, i-1, i))
		}
	}
}

// TestDelta_BusyGenuineReset checks that a real reset on a busy counter skips the reset poll and the next one
func TestDelta_BusyGenuineReset(t *testing.T) {
	const resetPoll = 3
	polls := replayPolls(10)
	resetCounters(polls, resetPoll, 1_000)

	cooked := replay(t, polls)
	assertSkippedPolls(t, cooked, polls, []int{resetPoll, resetPoll + 1})
}

// TestDelta_IdleGenuineResetClimbingPastHigh checks that an idle counter that is reset and then climbs past
// its old value costs one extra skipped poll, never a wrong value.
func TestDelta_IdleGenuineResetClimbingPastHigh(t *testing.T) {
	// 1 op/s, reset to 10 at poll 3, then 10 ops/s. Poll 5 reaches 1210, above the old 1120.
	polls := pollsFromOps(1000, 1060, 1120, 10, 610, 1210, 1810)

	cooked := replay(t, polls)
	assertSkippedPolls(t, cooked, polls, []int{3, 4, 5})
}

// TestDelta_StaleHighThenDip checks that a value left by an unflagged reset, which the counter never climbs
// back to, does not hide a later #4496-style dip. The dip replaces the stale value, and its restore reaches it.
func TestDelta_StaleHighThenDip(t *testing.T) {
	const (
		resetPoll = 3
		dipStart  = 6
		dipLen    = 3
		dipOps    = 1_000_000.0 // well above one poll of traffic, far below the pre-reset value
	)
	truth := replayPolls(12)
	resetCounters(truth, resetPoll, 50_000_000)
	polls := lowerPolls(truth, dipStart, dipStart+dipLen, dipOps)

	cooked := replay(t, polls)
	assertSkippedPolls(t, cooked, truth, []int{resetPoll, resetPoll + 1, dipStart, dipStart + 1, dipStart + dipLen})
}

// TestDelta_PartialAggregationDuringDrop checks that an ONTAP partial aggregation poll inside an
// unflagged drop does not forget the remembered value, so the restore poll is still skipped.
func TestDelta_PartialAggregationDuringDrop(t *testing.T) {
	const dropLen = 8
	tests := []struct {
		name    string
		partial int
		skipped []int
	}{
		{name: "partial in the middle of the drop", partial: 5, skipped: []int{2, 3, 5, 6, 10}},
		{name: "partial on the last dropped poll", partial: 9, skipped: []int{2, 3, 9, 10}},
		{name: "partial on the restore poll", partial: 10, skipped: []int{2, 3, 10, 11}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truth := replayPolls(dropStart + dropLen + 4)
			// a node is missing from the sum, so the partial value is low
			polls := lowerPolls(dropPolls(truth, dropLen), tt.partial, tt.partial+1, 1_000_000)
			polls[tt.partial].partial = true

			cooked := replay(t, polls)
			assertSkippedPolls(t, cooked, truth, tt.skipped)
		})
	}
}

// TestDelta_AllowedPartialAggregation covers allow_partial_aggregation: the collector clears the partial
// flag, so a node missing from the sum looks like an unflagged drop and its return is skipped as a restore.
func TestDelta_AllowedPartialAggregation(t *testing.T) {
	steps := deltaSteps(t, true, 100, 62, 64, 66, 108, 110)

	// 62: drop, skipped; 64: poll after the drop, skipped; 66: 2; 108: node back above 100, skipped; 110: 2
	want := []struct {
		value float64
		ok    bool
	}{{ok: false}, {ok: false}, {value: 2, ok: true}, {ok: false}, {value: 2, ok: true}}
	for i, w := range want {
		assert.Equal(t, steps[i].ok, w.ok)
		if w.ok {
			assertClose(t, fmt.Sprintf("delta %d", i), steps[i].value, w.value)
		}
	}
}

// TestDelta_HighsLifecycle checks what Delta keeps in highs poll by poll: nothing while the counter only
// increases, the value before a drop, that value replaced by a later drop, and released when it is reached
func TestDelta_HighsLifecycle(t *testing.T) {
	want := []struct {
		ok      bool    // delta exported
		delta   float64 // exported delta when ok
		held    bool    // a value is kept for the next poll
		high    float64 // the kept value when held
		comment string
	}{
		{ok: true, delta: 10, comment: "110: healthy"},
		{ok: true, delta: 10, comment: "120: healthy"},
		{held: true, high: 120, comment: "80: unflagged drop"},
		{held: true, high: 120, comment: "90: poll after the drop"},
		{held: true, high: 90, comment: "70: drops again while low, replaces the value"},
		{held: true, high: 90, comment: "80: poll after the drop"},
		{comment: "95: reaches the value, skipped and released"},
		{ok: true, delta: 35, comment: "130: nothing held, exported"},
		{ok: true, delta: 10, comment: "140: healthy"},
	}

	steps := deltaSteps(t, false, 100, 110, 120, 80, 90, 70, 80, 95, 130, 140)
	for i, w := range want {
		if steps[i].ok != w.ok {
			t.Fatalf("%s: exported=%v want %v", w.comment, steps[i].ok, w.ok)
		}
		if w.ok {
			assertClose(t, w.comment, steps[i].value, w.delta)
		}

		high, held := steps[i].cached.highs["ops"]["vol1"]
		if held != w.held {
			t.Fatalf("%s: held=%v want %v", w.comment, held, w.held)
		}
		if held {
			assertClose(t, w.comment, high, w.high)
		} else if steps[i].cached.highs != nil {
			// nothing is allocated while no counter is held
			t.Fatalf("%s: highs=%v want nil", w.comment, steps[i].cached.highs)
		}
	}
}

// TestDelta_ResetToZeroHoldsNothing checks that a reset to exactly 0 is treated as an invalid zero, which is
// skipped and never remembered
func TestDelta_ResetToZeroHoldsNothing(t *testing.T) {
	steps := deltaSteps(t, false, 100, 0, 10, 20)

	// 0: invalid zero, skipped; 10: poll after it, skipped; 20: healthy
	assert.False(t, steps[0].ok)
	assert.False(t, steps[1].ok)
	assert.True(t, steps[2].ok)
	assertClose(t, "delta", steps[2].value, 10)
	for _, s := range steps {
		assert.True(t, s.cached.highs == nil)
	}
}

// TestDelta_MissingInstanceLosesHigh records a known limit: an instance that is absent from a poll loses its
// remembered value, so when the counter comes back above it, that poll is exported instead of skipped.
func TestDelta_MissingInstanceLosesHigh(t *testing.T) {
	steps := deltaSteps(t, false, 100, 80, math.NaN(), 85, 120)

	_, held := steps[0].cached.highs["ops"]["vol1"]
	assert.True(t, held) // 80: unflagged drop from 100
	assert.True(t, steps[3].cached.highs == nil)
	// 85: instance is new to the cache, skipped; 120: exported, a restore would have been skipped
	assert.False(t, steps[2].ok)
	assert.True(t, steps[3].ok)
	assertClose(t, "delta", steps[3].value, 35)
}

// TestMatrix_CloneHighs checks that Clone copies highs independently, ForgetHighs only affects the matrix it
// is called on, and CloneForCollection does not copy them
func TestMatrix_CloneHighs(t *testing.T) {
	m := New("Test", "test", "test")
	m.setHigh("ops", "vol1", 120)

	c := m.Clone()
	assert.Equal(t, c.highs["ops"]["vol1"], 120.0)
	c.ForgetHighs("vol1")
	_, held := c.highs["ops"]["vol1"]
	assert.False(t, held)
	assert.Equal(t, m.highs["ops"]["vol1"], 120.0)

	assert.True(t, m.CloneForCollection().highs == nil)
}
