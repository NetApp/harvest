package eseriesperf

import (
	"fmt"
	"testing"

	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

type resetPoll struct {
	readOps   float64 // cumulative
	lastReset float64 // lastResetTimeInMS
}

const (
	resetPollStart    = 1_767_069_729_000.0 // observedTimeInMS of the first poll
	resetPollInterval = 60_000.0            // ms
	resetInitialReset = 1_767_065_505_000.0
)

// skip marks a poll that must not export readOps
const skip = -1.0

func volumeResults(p resetPoll, observedMS float64) []gjson.Result {
	j := fmt.Sprintf(`{"volumeStats":[{
		"volumeName":"v1",
		"volumeId":"020000006D039EA00045C1B6000047B465FBD1F5",
		"observedTimeInMS":%.0f,
		"lastResetTimeInMS":%.0f,
		"sourceController":"070000000000000000000002",
		"readOps":%.0f,
		"readTimeTotal":%.0f,
		"writeOps":0,
		"writeTimeTotal":0,
		"otherOps":0,
		"readBytes":0,
		"writeBytes":0,
		"readHitOps":0,
		"writeHitOps":0,
		"queueDepthTotal":0,
		"queueDepthMax":0,
		"idleTime":%.0f
	}]}`, observedMS, p.lastReset, p.readOps, p.readOps*500, observedMS)
	return []gjson.Result{gjson.Parse(j)}
}

// runResetPolls drives polls through pollData and cookCounters the way PollData does and
// returns the exported readOps rate per cooked poll (skip when it is not exported).
func runResetPolls(t *testing.T, polls []resetPoll) []float64 {
	t.Helper()
	ep := newEseriesPerf("Volume", "volume.yaml")
	mat := ep.Matrix[ep.Object]
	mat.SetGlobalLabel("array_id", "test-system")
	mat.SetGlobalLabel("array", "test")

	ep.pollData(mat, volumeResults(polls[0], resetPollStart), set.New())
	if _, err := ep.cookCounters(mat, mat); err != nil {
		t.Fatal(err)
	}

	rates := make([]float64, 0, len(polls)-1)
	for i, p := range polls[1:] {
		prevMat := ep.Matrix[ep.Object]
		curMat := prevMat.CloneForCollection()
		curMat.Reset()
		ep.pollData(curMat, volumeResults(p, resetPollStart+float64(i+1)*resetPollInterval), set.New())
		got, err := ep.cookCounters(curMat, prevMat)
		if err != nil {
			t.Fatal(err)
		}
		res := got[ep.Object]
		inst := res.GetInstance("v1")
		if inst == nil {
			t.Fatalf("poll %d: instance v1 not found", i+1)
		}
		v, ok := res.GetMetric("readOps").GetValueFloat64(inst)
		if !ok {
			v = skip
		}
		rates = append(rates, v)
	}
	return rates
}

func assertRates(t *testing.T, got, want []float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d polls, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("poll %d readOps got=%v want=%v (skip=%v)\n got=%v\nwant=%v", i+1, got[i], want[i], skip, got, want)
		}
	}
}

func TestEseriesPerf_Reset(t *testing.T) {
	reset := resetInitialReset + 100_000
	tests := []struct {
		name  string
		polls []resetPoll
		want  []float64
	}{
		{
			// A flagged reset whose counter climbs back past the pre-reset value must not cause an extra skip
			name: "reset to small value then climbs past old value",
			polls: []resetPoll{
				{1000, resetInitialReset}, {1060, resetInitialReset}, {1120, resetInitialReset},
				{10, reset}, {610, reset}, {1210, reset}, {1810, reset},
			},
			want: []float64{1, 1, skip, skip, 10, 10},
		},
		{
			name: "reset to zero",
			polls: []resetPoll{
				{1000, resetInitialReset}, {1060, resetInitialReset}, {1120, resetInitialReset},
				{0, reset}, {600, reset}, {1200, reset}, {1800, reset},
			},
			want: []float64{1, 1, skip, skip, 10, 10},
		},
		{
			name: "reset with counter already above previous value",
			polls: []resetPoll{
				{1000, resetInitialReset}, {1060, resetInitialReset}, {1120, resetInitialReset},
				{5000, reset}, {5600, reset}, {6200, reset},
			},
			want: []float64{1, 1, skip, skip, 10},
		},
		{
			name: "two resets in a row",
			polls: []resetPoll{
				{1000, resetInitialReset}, {1060, resetInitialReset},
				{10, reset}, {5, reset + 60_000}, {605, reset + 60_000}, {1205, reset + 60_000},
			},
			want: []float64{1, skip, skip, skip, 10},
		},
		{
			// An unflagged dip remembers the pre-drop value. A later flagged reset must forget it,
			// so the reset counter climbing past it is exported normally, as on main.
			name: "flagged reset after an unflagged dip",
			polls: []resetPoll{
				{1000, resetInitialReset}, {1060, resetInitialReset}, {900, resetInitialReset},
				{960, resetInitialReset}, {1020, resetInitialReset},
				{10, reset}, {610, reset}, {1210, reset}, {1810, reset},
			},
			want: []float64{1, skip, skip, 1, skip, skip, 10, 10},
		},
		{
			// No reset flag: E-Series did not explain the drop. The restore poll is skipped instead of
			// being exported as a spike. This is the only case that differs from main.
			name: "unflagged dip then restore",
			polls: []resetPoll{
				{1_000_000, resetInitialReset}, {1_000_060, resetInitialReset},
				{900_120, resetInitialReset}, {900_180, resetInitialReset}, {900_240, resetInitialReset},
				{1_000_300, resetInitialReset}, {1_000_360, resetInitialReset},
			},
			want: []float64{1, skip, skip, 1, skip, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertRates(t, runResetPolls(t, tt.polls), tt.want)
		})
	}
}
