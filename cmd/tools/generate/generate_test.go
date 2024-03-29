package generate

import "testing"

func Test_toMount(t *testing.T) {
	tests := []struct {
		name     string
		hostPath string
		want     string
	}{
		{name: "dot prefix", hostPath: "./abc/d", want: "./abc/d:/opt/harvest/abc/d"},
		{name: "absolute", hostPath: "/x/y/z", want: "/x/y/z:/x/y/z"},
		{name: "cwd", hostPath: "abc/d", want: "./abc/d:/opt/harvest/abc/d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toMount(tt.hostPath); got != tt.want {
				t.Errorf("toMount() = %v, want %v", got, tt.want)
			}
		})
	}
}
