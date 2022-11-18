package restperf

import (
	"github.com/tidwall/gjson"
	"os"
	"reflect"
	"testing"
)

func Test_parseMetricResponse(t *testing.T) {
	bytes, err := os.ReadFile("testdata/submetrics.json")
	if err != nil {
		t.Fatal(err)
	}
	instanceData := gjson.GetBytes(bytes, "records.0")
	type args struct {
		instanceData gjson.Result
		metric       string
	}
	tests := []struct {
		name string
		args args
		want *metricResponse
	}{
		{
			name: "rss_matrix",
			args: args{
				instanceData: instanceData, metric: "rss_matrix",
			}, want: &metricResponse{
				label:   "tx_frames#queue_0,tx_frames#queue_1,tx_frames#queue_2,tx_bytes#queue_0,tx_bytes#queue_1,tx_bytes#queue_2",
				value:   "6177010,1605252882,0,3,1,4",
				isArray: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseMetricResponse(tt.args.instanceData, tt.args.metric)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\ngot  %v\nwant %v", got, tt.want)
			}
		})
	}
}
