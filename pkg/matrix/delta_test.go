package matrix

import (
	"fmt"
	"log/slog"
	"math"
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
)

type replayPoll struct {
	ops     float64 // raw cumulative ops
	latency float64 // raw cumulative latency
	ts      float64 // seconds
	missing bool    // ops and latency are not reported
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

// dropPolls lowers the counters of polls [start, start+n) by replayDropOps, like ONTAP 9.17.1P9 did in #4496
func dropPolls(polls []replayPoll, start, n int) []replayPoll {
	dropped := make([]replayPoll, len(polls))
	copy(dropped, polls)
	for i := start; i < start+n && i < len(dropped); i++ {
		dropped[i].ops -= replayDropOps
		dropped[i].latency -= replayDropOps * replayLatencyPerOp
	}
	return dropped
}

func setReplayRaw(t *testing.T, mat *Matrix, p replayPoll) {
	t.Helper()
	instance := mat.GetInstance("vol1")
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

		// KeyPerf calculates the timestamp delta in map order, so do it last here
		for _, key := range []string{"total_ops", "avg_latency", "timestamp"} {
			_, err = curMat.Delta(key, prevMat, cachedData, "timestamp", false, logger)
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

func assertSane(t *testing.T, cooked []cookedPoll) {
	t.Helper()
	for i, c := range cooked {
		if c.hasRate && (c.rate < 0 || c.rate > replayMaxRate) {
			t.Errorf("poll %d exported rate %f, want between 0 and %f", i+1, c.rate, replayMaxRate)
		}
		if c.hasLatency {
			assertClose(t, fmt.Sprintf("poll %d latency", i+1), c.latency, replayLatencyPerOp)
		}
	}
}

// TestDelta_TemporaryDrop replays #4496: two good polls, then the counters drop by a constant amount
// for N polls while ONTAP reports them as ok, then they recover.
func TestDelta_TemporaryDrop(t *testing.T) {
	const dropStart = 2
	tests := []struct {
		name    string
		dropLen int
		missing int // index of a poll inside the drop with no reading, or 0
	}{
		{name: "1 low poll", dropLen: 1},
		{name: "2 low polls", dropLen: 2},
		{name: "5 low polls", dropLen: 5},
		{name: "9 low polls", dropLen: 9},
		{name: "5 low polls with a missing reading", dropLen: 5, missing: dropStart + 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truth := replayPolls(dropStart + tt.dropLen + 3)
			polls := dropPolls(truth, dropStart, tt.dropLen)
			if tt.missing != 0 {
				polls[tt.missing].missing = true
			}
			cooked := replay(t, polls)
			assertSane(t, cooked)

			recovery := dropStart + tt.dropLen
			lastGood := dropStart - 1
			for i := 1; i < len(polls); i++ {
				c := cooked[i-1]
				switch {
				case i >= dropStart && i < recovery:
					assert.False(t, c.hasRate)
					assert.False(t, c.hasLatency)
				case i == recovery:
					assert.True(t, c.hasRate)
					assert.True(t, c.hasLatency)
					assertClose(t, "recovery rate", c.rate, averageRate(truth, lastGood, recovery))
				default:
					assert.True(t, c.hasRate)
					assert.True(t, c.hasLatency)
					assertClose(t, fmt.Sprintf("poll %d rate", i), c.rate, averageRate(truth, i-1, i))
				}
			}
		})
	}
}

// TestDelta_PersistentDrop checks that a counter that stays lower than its last good reading for longer
// than maxHeldBaselineAge is accepted as a genuine counter reset
func TestDelta_PersistentDrop(t *testing.T) {
	const dropStart = 2
	lastGood := dropStart - 1
	// the first poll at least maxHeldBaselineAge after the last good reading accepts the reset
	acceptPoll := lastGood + int(math.Ceil(maxHeldBaselineAge.Seconds()/replayInterval))

	truth := replayPolls(acceptPoll + 4)
	polls := dropPolls(truth, dropStart, len(truth))
	cooked := replay(t, polls)
	assertSane(t, cooked)

	for i := 1; i < len(polls); i++ {
		c := cooked[i-1]
		switch {
		case i < dropStart:
			assert.True(t, c.hasRate)
		case i <= acceptPoll:
			assert.False(t, c.hasRate)
			assert.False(t, c.hasLatency)
		default:
			assert.True(t, c.hasRate)
			assert.True(t, c.hasLatency)
			assertClose(t, fmt.Sprintf("poll %d rate", i), c.rate, averageRate(truth, i-1, i))
		}
	}
}

// TestDelta_LargeDropIsReset checks that a counter that falls below minHeldFraction of its last good reading,
// like a counter that restarts after a node reboot, is accepted as reset at once instead of being held
func TestDelta_LargeDropIsReset(t *testing.T) {
	const resetPoll = 3
	truth := replayPolls(8)
	polls := make([]replayPoll, len(truth))
	copy(polls, truth)
	// the counters restart at 1,000 ops at resetPoll and then keep counting at the true rate
	for i := resetPoll; i < len(polls); i++ {
		polls[i].ops = truth[i].ops - truth[resetPoll].ops + 1_000
		polls[i].latency = polls[i].ops * replayLatencyPerOp
	}

	cooked := replay(t, polls)
	assertSane(t, cooked)

	for i := 1; i < len(polls); i++ {
		c := cooked[i-1]
		switch i {
		case resetPoll, resetPoll + 1:
			// the reset poll and the next one are skipped, as before held baselines
			assert.False(t, c.hasRate)
			assert.False(t, c.hasLatency)
		default:
			assert.True(t, c.hasRate)
			assert.True(t, c.hasLatency)
			assertClose(t, fmt.Sprintf("poll %d rate", i), c.rate, averageRate(truth, i-1, i))
		}
	}
}

// TestDelta_LatencyBaselineMismatch checks that a latency average is skipped when its numerator and
// denominator recover from baselines taken at different polls
func TestDelta_LatencyBaselineMismatch(t *testing.T) {
	truth := replayPolls(7)
	polls := make([]replayPoll, len(truth))
	copy(polls, truth)
	// ops drop at polls 2-3, latency only at poll 3. Both recover at poll 4.
	for i := 2; i <= 3; i++ {
		polls[i].ops -= replayDropOps
	}
	polls[3].latency -= replayDropOps * replayLatencyPerOp

	cooked := replay(t, polls)
	assertSane(t, cooked)

	recovery := cooked[4-1]
	assert.True(t, recovery.hasRate)
	assertClose(t, "recovery rate", recovery.rate, averageRate(truth, 1, 4))
	assert.False(t, recovery.hasLatency)

	next := cooked[5-1]
	assert.True(t, next.hasRate)
	assert.True(t, next.hasLatency)
}
