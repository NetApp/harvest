package main

import "testing"

func TestPingParsing(t *testing.T) {
	poller := Poller{}

	type test struct {
		name string
		out  string
		ping float32
		isOK bool
		want int
	}

	tests := []test{
		{
			name: "NotBusy",
			ping: 0.032,
			isOK: true,
			out: `PING 127.0.0.1 (127.0.0.1) 56(84) bytes of data.

	--- 127.0.0.1 ping statistics ---
	1 packets transmitted, 1 received, 0% packet loss, time 0ms
	rtt min/avg/max/mdev = 0.032/0.032/0.032/0.000 ms`,
		},
		{
			name: "BusyBox",
			ping: 0.088,
			isOK: true,
			out: `PING 127.0.0.1 (127.0.0.1): 56 data bytes

--- 127.0.0.1 ping statistics ---
1 packets transmitted, 1 packets received, 0% packet loss
round-trip min/avg/max = 0.088/0.088/0.088 ms`,
		},
		{
			name: "BadInput",
			ping: 0,
			isOK: false,
			out:  `foo`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ping, b := poller.parsePing(tt.out)
			if ping != tt.ping {
				t.Errorf("parsePing ping got = %v, want %v", ping, tt.ping)
			}
			if b != tt.isOK {
				t.Errorf("parsePing isOK got = %v, want %v", b, tt.isOK)
			}
		})
	}
}
