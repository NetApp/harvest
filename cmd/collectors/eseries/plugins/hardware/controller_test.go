package hardware

import (
	"log/slog"
	"testing"

	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

// newTestHardware returns a Hardware ready to test processHostInterfacesFromAPI.
func newTestHardware() *Hardware {
	h := &Hardware{
		AbstractPlugin: &plugin.AbstractPlugin{
			SLogger: slog.Default(),
			Parent:  "test",
		},
		data: make(map[string]*matrix.Matrix),
	}
	h.initHostInterfaceMatrix()
	return h
}

func parseResults(jsonStr string) []gjson.Result {
	return gjson.Parse(jsonStr).Array()
}

func TestProcessHostInterfacesFromAPI(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		ctrlMap    map[string]string
		portMap    map[string]string
		wantCount  int
		wantKey    string
		wantLabels map[string]string
	}{
		{
			name:      "FC: key mapped to fibre, linkStatus and currentInterfaceSpeed used",
			wantCount: 1,
			wantKey:   "ctrl1_ref1",
			ctrlMap:   map[string]string{"ctrl1": "A"},
			portMap:   map[string]string{},
			wantLabels: map[string]string{
				"interface_type": "fc",
				"link_state":     "up",
				"speed":          "16000",
				"controller":     "A",
			},
			json: `[{"interfaceRef":"ref1","controllerRef":"ctrl1","ioInterfaceTypeData":{"interfaceType":"fc","fibre":{"channel":"1","linkStatus":"up","currentInterfaceSpeed":"speed16gig","physicalLocation":{"label":"0a"}}}}]`,
		},
		{
			name:       "nvmeCouplingDriver: key mapped to couplingDriverNvme",
			wantCount:  1,
			wantKey:    "ctrl1_ref1",
			ctrlMap:    map[string]string{},
			portMap:    map[string]string{},
			wantLabels: map[string]string{"interface_type": "nvmeCouplingDriver"},
			json:       `[{"interfaceRef":"ref1","controllerRef":"ctrl1","ioInterfaceTypeData":{"interfaceType":"nvmeCouplingDriver","couplingDriverNvme":{"channel":"1","physicalLocation":{"label":""}}}}]`,
		},
		{
			name:      "iSCSI: linkState and currentSpeed used, ip_address from ipv4Data",
			wantCount: 1,
			wantKey:   "ctrl1_ref1",
			ctrlMap:   map[string]string{},
			portMap:   map[string]string{},
			wantLabels: map[string]string{
				"interface_type": "iscsi",
				"link_state":     "up",
				"speed":          "10000",
				"ip_address":     "192.168.1.1",
			},
			json: `[{"interfaceRef":"ref1","controllerRef":"ctrl1","ioInterfaceTypeData":{"interfaceType":"iscsi","iscsi":{"channel":"1","linkState":"up","currentSpeed":"speed10gig","ipv4Data":{"ipv4AddressData":{"ipv4Address":"192.168.1.1"}},"physicalLocation":{"label":""}}},"commandProtocolPropertiesList":{"commandProtocolProperties":[]}}]`,
		},
		{
			name:       "IB: same key name, instance created",
			wantCount:  1,
			wantKey:    "ctrl1_ref1",
			ctrlMap:    map[string]string{},
			portMap:    map[string]string{},
			wantLabels: map[string]string{"interface_type": "ib"},
			json:       `[{"interfaceRef":"ref1","controllerRef":"ctrl1","ioInterfaceTypeData":{"interfaceType":"ib","ib":{"channel":"1","physicalLocation":{"label":""}}}}]`,
		},
		{
			name:       "ethernet: same key name, instance created",
			wantCount:  1,
			wantKey:    "ctrl1_ref1",
			ctrlMap:    map[string]string{},
			portMap:    map[string]string{},
			wantLabels: map[string]string{"interface_type": "ethernet"},
			json:       `[{"interfaceRef":"ref1","controllerRef":"ctrl1","ioInterfaceTypeData":{"interfaceType":"ethernet","ethernet":{"channel":"1","physicalLocation":{"label":""}}}}]`,
		},

		// --- Label correctness ---
		{
			name:       "portLabelMap takes priority over physicalLocation.label",
			wantCount:  1,
			wantKey:    "ctrl1_ref1",
			ctrlMap:    map[string]string{},
			portMap:    map[string]string{"1": "0a"},
			wantLabels: map[string]string{"port": "0a"},
			json:       `[{"interfaceRef":"ref1","controllerRef":"ctrl1","ioInterfaceTypeData":{"interfaceType":"fc","fibre":{"channel":"1","physicalLocation":{"label":"wronglabel"}}}}]`,
		},
		{
			name:       "portLabelMap miss falls back to physicalLocation.label",
			wantCount:  1,
			wantKey:    "ctrl1_ref1",
			ctrlMap:    map[string]string{},
			portMap:    map[string]string{},
			wantLabels: map[string]string{"port": "0b"},
			json:       `[{"interfaceRef":"ref1","controllerRef":"ctrl1","ioInterfaceTypeData":{"interfaceType":"fc","fibre":{"channel":"2","physicalLocation":{"label":"0b"}}}}]`,
		},
		{
			name:       "linkState fallback: empty linkState uses linkStatus",
			wantCount:  1,
			wantKey:    "ctrl1_ref1",
			ctrlMap:    map[string]string{},
			portMap:    map[string]string{},
			wantLabels: map[string]string{"link_state": "up"},
			json:       `[{"interfaceRef":"ref1","controllerRef":"ctrl1","ioInterfaceTypeData":{"interfaceType":"fc","fibre":{"channel":"1","linkStatus":"up","physicalLocation":{"label":""}}}}]`,
		},
		{
			name:       "speed fallback: empty currentSpeed uses currentInterfaceSpeed",
			wantCount:  1,
			wantKey:    "ctrl1_ref1",
			ctrlMap:    map[string]string{},
			portMap:    map[string]string{},
			wantLabels: map[string]string{"speed": "16000"},
			json:       `[{"interfaceRef":"ref1","controllerRef":"ctrl1","ioInterfaceTypeData":{"interfaceType":"fc","fibre":{"channel":"1","currentInterfaceSpeed":"speed16gig","physicalLocation":{"label":""}}}}]`,
		},
		{
			name:       "physical_port_state: linkUp cleaned to Up",
			wantCount:  1,
			wantKey:    "ctrl1_ref1",
			ctrlMap:    map[string]string{},
			portMap:    map[string]string{},
			wantLabels: map[string]string{"physical_port_state": "Up"},
			json:       `[{"interfaceRef":"ref1","controllerRef":"ctrl1","ioInterfaceTypeData":{"interfaceType":"fc","fibre":{"channel":"1","physPortState":"linkUp","physicalLocation":{"label":""}}}}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHardware()
			h.processHostInterfacesFromAPI(parseResults(tt.json), tt.ctrlMap, tt.portMap)

			instances := h.data[hostInterfaceMatrix].GetInstances()
			if len(instances) != tt.wantCount {
				t.Fatalf("instance count = %d, want %d", len(instances), tt.wantCount)
			}

			if tt.wantKey == "" || len(tt.wantLabels) == 0 {
				return
			}

			inst := instances[tt.wantKey]
			if inst == nil {
				t.Fatalf("instance %q not found", tt.wantKey)
			}
			for label, want := range tt.wantLabels {
				if got := inst.GetLabel(label); got != want {
					t.Errorf("label %q = %q, want %q", label, got, want)
				}
			}
		})
	}
}

func TestSpeedToMB(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Empty and special cases
		{"empty string", "", ""},
		{"100 meg", "speed100meg", "100"},

		// Gigabit speeds - small
		{"1 gig", "speed1gig", "1000"},
		{"2 gig", "speed2gig", "2000"},
		{"unknown", "speedUnknown", "Unknown"},
		{"auto", "speedAuto", "Auto"},
		{"undefined", "__UNDEFINED", "__UNDEFINED"},

		// Megabit speeds
		{"10 meg", "speed10meg", "10"},
		{"3 gig", "speed3gig", "3000"},
		{"4 gig", "speed4gig", "4000"},
		{"5 gig", "speed5gig", "5000"},
		{"6 gig", "speed6gig", "6000"},
		{"8 gig", "speed8gig", "8000"},

		// Gigabit speeds - medium
		{"10 gig", "speed10gig", "10000"},
		{"12 gig", "speed12gig", "12000"},
		{"15 gig", "speed15gig", "15000"},
		{"16 gig", "speed16gig", "16000"},
		{"20 gig", "speed20gig", "20000"},
		{"24 gig", "speed24gig", "24000"},
		{"25 gig", "speed25gig", "25000"},

		// Gigabit speeds - large
		{"30 gig", "speed30gig", "30000"},
		{"32 gig", "speed32gig", "32000"},
		{"40 gig", "speed40gig", "40000"},
		{"50 gig", "speed50gig", "50000"},
		{"56 gig", "speed56gig", "56000"},
		{"60 gig", "speed60gig", "60000"},
		{"64 gig", "speed64gig", "64000"},

		// Gigabit speeds - very large
		{"100 gig", "speed100gig", "100000"},
		{"128 gig", "speed128gig", "128000"},
		{"200 gig", "speed200gig", "200000"},

		// Decimal speeds with "pt" notation
		{"2.5 gig", "speed2pt5Gig", "2500"},
		{"22.5 gig", "speed22pt5Gig", "22500"},

		// Case insensitivity
		{"uppercase GIG", "speed10GIG", "10000"},
		{"mixed case", "speed10Gig", "10000"},
		{"uppercase MEG", "speed100MEG", "100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := speedToMB(tt.input)
			if result != tt.expected {
				t.Errorf("speedToMB(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatIPv6Address(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Invalid lengths
		{"empty string", "", ""},
		{"too short", "FE80", "FE80"},
		{"too long", "FE80000000000000D239EAFFFEDCD97C00", "FE80000000000000D239EAFFFEDCD97C00"},

		// Valid IPv6 conversions with compression
		{
			"link-local address",
			"FE80000000000000D239EAFFFEDCD97C",
			"fe80::d239:eaff:fedc:d97c",
		},
		{
			"loopback address",
			"00000000000000000000000000000001",
			"::1",
		},
		{
			"all zeros",
			"00000000000000000000000000000000",
			"::",
		},
		{
			"leading zeros compression",
			"00000000000000000000000012345678",
			"::1234:5678",
		},
		{
			"trailing zeros compression",
			"12345678000000000000000000000000",
			"1234:5678::",
		},
		{
			"middle zeros compression",
			"12340000000000000000000000005678",
			"1234::5678",
		},
		{
			"multiple zero runs - compress longest",
			"12340000000056780000000000009ABC",
			"1234:0:0:5678::9abc",
		},
		{
			"no compression needed",
			"12345678ABCDEF0123456789ABCDEF01",
			"1234:5678:abcd:ef01:2345:6789:abcd:ef01",
		},
		{
			"single zero group - no compression",
			"12340000567890ABCDEF012345678ABC",
			"1234:0:5678:90ab:cdef:123:4567:8abc",
		},
		{
			"uppercase input converted to lowercase",
			"FE80000000000000ABCDEFABCDEFABCD",
			"fe80::abcd:efab:cdef:abcd",
		},
		{
			"documentation prefix",
			"20010DB8000000000000000000000001",
			"2001:db8::1",
		},
		{
			"two equal-length runs - compress longest (second)",
			"12340000000056780000000000000000",
			"1234:0:0:5678::",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatIPv6Address(tt.input)
			if result != tt.expected {
				t.Errorf("formatIPv6Address(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
