package volumesnaplock

import (
	"testing"

	"github.com/netapp/harvest/v2/assert"
)

func TestDurationSecondsAndDisplay_Defaults(t *testing.T) {
	type tc struct {
		in        string
		wantSec   int64
		wantDisp  string
		wantError bool
	}

	tests := []tc{
		{in: "P1Y", wantSec: 31536000, wantDisp: "1 year"},
		{in: "p30y", wantSec: 946080000, wantDisp: "30 years"},
		{in: "P6M", wantSec: 15552000, wantDisp: "6 months"},
		{in: "P10D", wantSec: 864000, wantDisp: "10 days"},
		{in: "PT2H", wantSec: 7200, wantDisp: "2 hours"},
		{in: "PT45M", wantSec: 2700, wantDisp: "45 minutes"},
		{in: "PT0M", wantSec: 0, wantDisp: "0 minutes"},
		{in: "infinite", wantSec: -1, wantDisp: "infinite"},
		{in: "none", wantSec: 0, wantDisp: "none"},
		{in: "P1Y10M", wantError: true},
		{in: "", wantError: true},
	}

	for _, tt := range tests {
		sec, disp, err := DurationSecondsAndDisplay(tt.in, DefaultConfig())
		if tt.wantError {
			assert.NotNil(t, err)
			continue
		}
		assert.Nil(t, err)
		assert.Equal(t, sec, tt.wantSec)
		assert.Equal(t, disp, tt.wantDisp)
	}
}
