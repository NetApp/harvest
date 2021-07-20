package version

import "testing"

func Test_isNewerAvailable(t *testing.T) {
	type test struct {
		name          string
		curVersion    string
		remoteVersion string
		shouldUpgrade bool
	}
	tests := []test{
		{name: "do not upgrade 21.05.1 to same", curVersion: "v21.05.1", remoteVersion: "v21.05.1", shouldUpgrade: false},
		{name: "upgrade 21.05.1 to 21.05.2", curVersion: "v21.05.1", remoteVersion: "v21.05.2", shouldUpgrade: true},
		{name: "upgrade 21.05.1 to 21.11.1", curVersion: "v21.05.1", remoteVersion: "v21.11.1", shouldUpgrade: true},
		{name: "upgrade 21.05.1 to 22.02.1", curVersion: "v21.05.1", remoteVersion: "v22.02.1", shouldUpgrade: true},
		{name: "do not upgrade 21.07.2017 to v21.05.1", curVersion: "21.07.2017", remoteVersion: "v21.05.1", shouldUpgrade: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			available, err := isNewerAvailable(tt.curVersion, tt.remoteVersion)
			if err != nil {
				panic(err)
			}
			if available != tt.shouldUpgrade {
				t.Errorf("expected %t got %t for %+v", tt.shouldUpgrade, available, tt)
			}
		})
	}
}
