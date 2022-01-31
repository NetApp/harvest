package rest

import (
	"testing"
)

func Test_HandleIfTimeField(t *testing.T) {

	type test struct {
		timeFieldValue string
		want           float64
	}

	var tests = []test{
		{
			timeFieldValue: "PT54S",
			want:           54,
		},
		{
			timeFieldValue: "PT48M",
			want:           2880,
		},
		{
			timeFieldValue: "P428DT22H45M19S",
			want:           37061119,
		},
		{
			timeFieldValue: "PT8H35M42S",
			want:           30942,
		},
		{
			timeFieldValue: "2020-12-02T18:36:19-08:00",
			want:           1606962979,
		},
		{
			timeFieldValue: "2022-01-31T04:05:02-05:00",
			want:           1643619902,
		},
		{
			timeFieldValue: "1234",
			want:           1234,
		},
		{
			timeFieldValue: "True",
			want:           0,
		},
		{
			timeFieldValue: "online",
			want:           0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.timeFieldValue, func(t *testing.T) {
			if got := HandleIfTimeField(tt.timeFieldValue); got != tt.want {
				t.Errorf("actual value = %v, want %v", got, tt.want)
			}
		})
	}
}
