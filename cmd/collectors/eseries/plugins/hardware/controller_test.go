package hardware

import (
	"log/slog"
	"testing"

	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

// newTestHardware returns a Hardware ready to test processHostInterfacesFromAPI,
// processDNSProperties, and processNTPProperties.
func newTestHardware() *Hardware {
	h := &Hardware{
		AbstractPlugin: &plugin.AbstractPlugin{
			SLogger: slog.Default(),
			Parent:  "test",
		},
		data: make(map[string]*matrix.Matrix),
	}
	h.initHostInterfaceMatrix()
	h.initDNSMatrix()
	h.initNTPMatrix()
	h.initNetInterfaceMatrix()
	h.initControllerMatrix()
	return h
}

func parseResults(jsonStr string) []gjson.Result {
	return gjson.Parse(jsonStr).Array()
}

// netInterfaceController wraps the DNS and NTP portion of one ethernet port in the controller
// shape processNetInterfaces expects, so each test case only spells out the part under test.
func netInterfaceController(ethernetProperties string) gjson.Result {
	return gjson.Parse(`{"netInterfaces":[{"ethernet":{"interfaceName":"wan0",` + ethernetProperties + `}}]}`)
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

func TestProcessDNSProperties(t *testing.T) {
	tests := []struct {
		name            string
		json            string
		wantCount       int
		wantServerCount *int // eseries_controller.dns_server_count; nil skips the check
		wantKey         string
		wantLabels      map[string]string
	}{
		{
			name:            "static: two IPv4 servers",
			wantCount:       2,
			wantServerCount: new(2),
			wantKey:         "ctrl1_10.192.0.250",
			wantLabels: map[string]string{
				"controller_id":    "ctrl1",
				"controller":       "A",
				"dns_server":       "10.192.0.250",
				"address_type":     "ipv4",
				"acquisition_type": "Static",
			},
			json: `{"controllerRef":"ctrl1","networkSettings":{"dnsProperties":{"acquisitionProperties":{"dnsAcquisitionType":"stat","dnsServers":[{"addressType":"ipv4","ipv4Address":"10.192.0.250"},{"addressType":"ipv4","ipv4Address":"10.193.0.250"}]}}}}`,
		},
		{
			name:            "static: IPv6 server uses ipv6Address field",
			wantCount:       1,
			wantServerCount: new(1),
			wantKey:         "ctrl1_2001:db8::1",
			wantLabels: map[string]string{
				"dns_server":   "2001:db8::1",
				"address_type": "ipv6",
			},
			json: `{"controllerRef":"ctrl1","networkSettings":{"dnsProperties":{"acquisitionProperties":{"dnsAcquisitionType":"stat","dnsServers":[{"addressType":"ipv6","ipv6Address":"2001:db8::1"}]}}}}`,
		},
		{
			name:            "static: mixed-case addressType is normalized to lowercase in the exported label",
			wantCount:       1,
			wantServerCount: new(1),
			wantKey:         "ctrl1_10.192.0.251",
			wantLabels: map[string]string{
				"dns_server":   "10.192.0.251",
				"address_type": "ipv4",
			},
			json: `{"controllerRef":"ctrl1","networkSettings":{"dnsProperties":{"acquisitionProperties":{"dnsAcquisitionType":"stat","dnsServers":[{"addressType":"IPv4","ipv4Address":"10.192.0.251"}]}}}}`,
		},
		{
			name:            "regression: dnsProperties at the controller top level is ignored, only networkSettings.dnsProperties is read",
			wantCount:       0,
			wantServerCount: new(0),
			json:            `{"controllerRef":"ctrl1","dnsProperties":{"acquisitionProperties":{"dnsAcquisitionType":"stat","dnsServers":[{"addressType":"ipv4","ipv4Address":"10.192.0.250"}]}}}`,
		},
		{
			name:            "dhcp: servers come from dhcpAcquiredDnsServers, not the empty static list",
			wantCount:       1,
			wantServerCount: new(1),
			wantKey:         "ctrl1_10.192.0.99",
			wantLabels: map[string]string{
				"dns_server":       "10.192.0.99",
				"acquisition_type": "dhcp",
			},
			json: `{"controllerRef":"ctrl1","networkSettings":{"dnsProperties":{"acquisitionProperties":{"dnsAcquisitionType":"dhcp","dnsServers":[]},"dhcpAcquiredDnsServers":[{"addressType":"ipv4","ipv4Address":"10.192.0.99"}]}}}`,
		},
		{
			name:            "dhcp: empty dhcpAcquiredDnsServers yields no instances",
			wantCount:       0,
			wantServerCount: new(0),
			json:            `{"controllerRef":"ctrl1","networkSettings":{"dnsProperties":{"acquisitionProperties":{"dnsAcquisitionType":"dhcp"},"dhcpAcquiredDnsServers":[]}}}`,
		},
		{
			name:            "disabled: no servers on either list",
			wantCount:       0,
			wantServerCount: new(0),
			json:            `{"controllerRef":"ctrl1","networkSettings":{"dnsProperties":{"acquisitionProperties":{"dnsAcquisitionType":"disabled","dnsServers":null}}}}`,
		},
		{
			name:            "unknown: populated static list is not exported (whitelist covers only stat and dhcp)",
			wantCount:       0,
			wantServerCount: new(0),
			json:            `{"controllerRef":"ctrl1","networkSettings":{"dnsProperties":{"acquisitionProperties":{"dnsAcquisitionType":"unknown","dnsServers":[{"addressType":"ipv4","ipv4Address":"10.192.0.250"}]}}}}`,
		},
		{
			name:            "static: three servers all count, uncapped (no maxItems in the swagger)",
			wantCount:       3,
			wantServerCount: new(3),
			json:            `{"controllerRef":"ctrl1","networkSettings":{"dnsProperties":{"acquisitionProperties":{"dnsAcquisitionType":"stat","dnsServers":[{"addressType":"ipv4","ipv4Address":"10.192.0.250"},{"addressType":"ipv4","ipv4Address":"10.193.0.250"},{"addressType":"ipv4","ipv4Address":"10.194.0.250"}]}}}}`,
		},
		{
			name:            "missing networkSettings.dnsProperties: no instances, count is 0 not absent",
			wantCount:       0,
			wantServerCount: new(0),
			json:            `{"controllerRef":"ctrl1","networkSettings":{}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHardware()
			controller := gjson.Parse(tt.json)
			controllerID := controller.Get("controllerRef").ClonedString()
			h.data[controllerMatrix].GetOrCreateInstance(controllerID)
			h.processDNSProperties(controller, controllerID, "A")

			instances := h.data[dnsPropertyMatrix].GetInstances()
			if len(instances) != tt.wantCount {
				t.Fatalf("instance count = %d, want %d", len(instances), tt.wantCount)
			}

			if tt.wantServerCount != nil {
				got, ok := h.data[controllerMatrix].GetMetric("dns_server_count").GetValueFloat64(h.data[controllerMatrix].GetInstance(controllerID))
				if !ok {
					t.Fatalf("dns_server_count not set")
				}
				if int(got) != *tt.wantServerCount {
					t.Errorf("dns_server_count = %v, want %d", got, *tt.wantServerCount)
				}
			}

			if tt.wantKey == "" {
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

func TestProcessNTPProperties(t *testing.T) {
	tests := []struct {
		name            string
		json            string
		wantCount       int
		wantServerCount *int // eseries_controller.ntp_server_count; nil skips the check
		wantKey         string
		wantLabels      map[string]string
	}{
		{
			name:            "static: two ipvx servers",
			wantCount:       2,
			wantServerCount: new(2),
			wantKey:         "ctrl1_10.192.88.175",
			wantLabels: map[string]string{
				"controller_id":    "ctrl1",
				"controller":       "A",
				"ntp_server":       "10.192.88.175",
				"address_type":     "ipv4",
				"acquisition_type": "Static",
			},
			json: `{"controllerRef":"ctrl1","networkSettings":{"ntpProperties":{"acquisitionProperties":{"ntpAcquisitionType":"stat","ntpServers":[{"addrType":"ipvx","ipvxAddress":{"addressType":"ipv4","ipv4Address":"10.192.88.175"}},{"addrType":"ipvx","ipvxAddress":{"addressType":"ipv4","ipv4Address":"10.192.88.150"}}]}}}}`,
		},
		{
			name:            "static: domainName server, no ipvxAddress",
			wantCount:       1,
			wantServerCount: new(1),
			wantKey:         "ctrl1_time-a-b.nist.gov",
			wantLabels: map[string]string{
				"ntp_server":       "time-a-b.nist.gov",
				"address_type":     "domainName",
				"acquisition_type": "Static",
			},
			json: `{"controllerRef":"ctrl1","networkSettings":{"ntpProperties":{"acquisitionProperties":{"ntpAcquisitionType":"stat","ntpServers":[{"addrType":"domainName","domainName":"time-a-b.nist.gov","ipvxAddress":null}]}}}}`,
		},
		{
			name:            "static: ntpServers is null",
			wantCount:       0,
			wantServerCount: new(0),
			json:            `{"controllerRef":"ctrl1","networkSettings":{"ntpProperties":{"acquisitionProperties":{"ntpAcquisitionType":"stat","ntpServers":null}}}}`,
		},
		{
			name:            "static: unrecognized addrType is skipped",
			wantCount:       0,
			wantServerCount: new(0),
			json:            `{"controllerRef":"ctrl1","networkSettings":{"ntpProperties":{"acquisitionProperties":{"ntpAcquisitionType":"stat","ntpServers":[{"addrType":"none"}]}}}}`,
		},
		{
			name:            "dhcp: servers come from dhcpAcquiredNtpServers",
			wantCount:       1,
			wantServerCount: new(1),
			wantKey:         "ctrl1_10.192.88.99",
			wantLabels: map[string]string{
				"ntp_server":       "10.192.88.99",
				"acquisition_type": "dhcp",
			},
			json: `{"controllerRef":"ctrl1","networkSettings":{"ntpProperties":{"acquisitionProperties":{"ntpAcquisitionType":"dhcp"},"dhcpAcquiredNtpServers":[{"addressType":"ipv4","ipv4Address":"10.192.88.99"}]}}}`,
		},
		{
			name:            "dhcp: empty dhcpAcquiredNtpServers yields no instances",
			wantCount:       0,
			wantServerCount: new(0),
			json:            `{"controllerRef":"ctrl1","networkSettings":{"ntpProperties":{"acquisitionProperties":{"ntpAcquisitionType":"dhcp"},"dhcpAcquiredNtpServers":[]}}}`,
		},
		{
			name:            "disabled: no servers on either list",
			wantCount:       0,
			wantServerCount: new(0),
			json:            `{"controllerRef":"ctrl1","networkSettings":{"ntpProperties":{"acquisitionProperties":{"ntpAcquisitionType":"disabled","ntpServers":null}}}}`,
		},
		{
			name:            "disabled: populated static list is not exported (whitelist covers only stat and dhcp)",
			wantCount:       0,
			wantServerCount: new(0),
			json:            `{"controllerRef":"ctrl1","networkSettings":{"ntpProperties":{"acquisitionProperties":{"ntpAcquisitionType":"disabled","ntpServers":[{"addrType":"ipvx","ipvxAddress":{"addressType":"ipv4","ipv4Address":"10.192.88.175"}}]}}}}`,
		},
		{
			name:            "duplicate address: one instance, no panic",
			wantCount:       1,
			wantServerCount: new(1),
			wantKey:         "ctrl1_10.192.88.175",
			json:            `{"controllerRef":"ctrl1","networkSettings":{"ntpProperties":{"acquisitionProperties":{"ntpAcquisitionType":"stat","ntpServers":[{"addrType":"ipvx","ipvxAddress":{"addressType":"ipv4","ipv4Address":"10.192.88.175"}},{"addrType":"ipvx","ipvxAddress":{"addressType":"ipv4","ipv4Address":"10.192.88.175"}}]}}}}`,
		},
		{
			name:            "missing networkSettings.ntpProperties: no instances, count is 0 not absent",
			wantCount:       0,
			wantServerCount: new(0),
			json:            `{"controllerRef":"ctrl1","networkSettings":{}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHardware()
			controller := gjson.Parse(tt.json)
			controllerID := controller.Get("controllerRef").ClonedString()
			h.data[controllerMatrix].GetOrCreateInstance(controllerID)
			h.processNTPProperties(controller, controllerID, "A")

			instances := h.data[ntpPropertyMatrix].GetInstances()
			if len(instances) != tt.wantCount {
				t.Fatalf("instance count = %d, want %d", len(instances), tt.wantCount)
			}

			if tt.wantServerCount != nil {
				got, ok := h.data[controllerMatrix].GetMetric("ntp_server_count").GetValueFloat64(h.data[controllerMatrix].GetInstance(controllerID))
				if !ok {
					t.Fatalf("ntp_server_count not set")
				}
				if int(got) != *tt.wantServerCount {
					t.Errorf("ntp_server_count = %v, want %d", got, *tt.wantServerCount)
				}
			}

			if tt.wantKey == "" {
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

func TestProcessNetInterfaces(t *testing.T) {
	tests := []struct {
		name       string
		properties string
		wantLabels map[string]string
	}{
		{
			name:       "static dns: servers in precedence order",
			properties: `"dnsProperties":{"acquisitionProperties":{"dnsAcquisitionType":"stat","dnsServers":[{"addressType":"ipv4","ipv4Address":"10.192.0.250"},{"addressType":"ipv4","ipv4Address":"10.193.0.250"}]}}`,
			wantLabels: map[string]string{
				"dns_config_method":  "Static",
				"primary_dns_server": "10.192.0.250",
				"backup_dns_server":  "10.193.0.250",
			},
		},
		{
			name:       "dhcp dns: servers come from dhcpAcquiredDnsServers, not the empty static list",
			properties: `"dnsProperties":{"acquisitionProperties":{"dnsAcquisitionType":"dhcp","dnsServers":[]},"dhcpAcquiredDnsServers":[{"addressType":"ipv4","ipv4Address":"10.192.0.99"},{"addressType":"ipv4","ipv4Address":"10.192.0.98"}]}`,
			wantLabels: map[string]string{
				"dns_config_method":  "dhcp",
				"primary_dns_server": "10.192.0.99",
				"backup_dns_server":  "10.192.0.98",
			},
		},
		{
			name:       "dhcp dns: empty acquired list leaves both addresses empty",
			properties: `"dnsProperties":{"acquisitionProperties":{"dnsAcquisitionType":"dhcp"},"dhcpAcquiredDnsServers":[]}`,
			wantLabels: map[string]string{
				"dns_config_method":  "dhcp",
				"primary_dns_server": "",
				"backup_dns_server":  "",
			},
		},
		{
			name:       "static dns: only the first two of three servers surface",
			properties: `"dnsProperties":{"acquisitionProperties":{"dnsAcquisitionType":"stat","dnsServers":[{"addressType":"ipv4","ipv4Address":"10.192.0.250"},{"addressType":"ipv4","ipv4Address":"10.193.0.250"},{"addressType":"ipv4","ipv4Address":"10.194.0.250"}]}}`,
			wantLabels: map[string]string{
				"primary_dns_server": "10.192.0.250",
				"backup_dns_server":  "10.193.0.250",
			},
		},
		{
			name:       "static ntp: two ipvx servers",
			properties: `"ntpProperties":{"acquisitionProperties":{"ntpAcquisitionType":"stat","ntpServers":[{"addrType":"ipvx","ipvxAddress":{"addressType":"ipv4","ipv4Address":"10.192.88.175"}},{"addrType":"ipvx","ipvxAddress":{"addressType":"ipv4","ipv4Address":"10.192.88.150"}}]}}`,
			wantLabels: map[string]string{
				"ntp_service":        "Static",
				"primary_ntp_server": "10.192.88.175",
				"backup_ntp_server":  "10.192.88.150",
			},
		},
		{
			name:       "static ntp: domainName server yields an FQDN with no backup",
			properties: `"ntpProperties":{"acquisitionProperties":{"ntpAcquisitionType":"stat","ntpServers":[{"addrType":"domainName","domainName":"time-a-b.nist.gov","ipvxAddress":null}]}}`,
			wantLabels: map[string]string{
				"ntp_service":        "Static",
				"primary_ntp_server": "time-a-b.nist.gov",
				"backup_ntp_server":  "",
			},
		},
		{
			name:       "static ntp: an unrecognized entry ahead of a valid one is skipped, not treated as primary",
			properties: `"ntpProperties":{"acquisitionProperties":{"ntpAcquisitionType":"stat","ntpServers":[{"addrType":"none"},{"addrType":"ipvx","ipvxAddress":{"addressType":"ipv4","ipv4Address":"10.192.88.175"}}]}}`,
			wantLabels: map[string]string{
				"primary_ntp_server": "10.192.88.175",
				"backup_ntp_server":  "",
			},
		},
		{
			name:       "dhcp ntp: servers come from dhcpAcquiredNtpServers in the flat ipvx shape",
			properties: `"ntpProperties":{"acquisitionProperties":{"ntpAcquisitionType":"dhcp"},"dhcpAcquiredNtpServers":[{"addressType":"ipv4","ipv4Address":"10.192.88.99"}]}`,
			wantLabels: map[string]string{
				"ntp_service":        "dhcp",
				"primary_ntp_server": "10.192.88.99",
				"backup_ntp_server":  "",
			},
		},
		{
			name:       "disabled ntp: null server list still records the method",
			properties: `"ntpProperties":{"acquisitionProperties":{"ntpAcquisitionType":"disabled","ntpServers":null}}`,
			wantLabels: map[string]string{
				"ntp_service":        "disabled",
				"primary_ntp_server": "",
				"backup_ntp_server":  "",
			},
		},
		{
			name:       "disabled ntp: populated server list still yields no exported addresses (whitelist covers only stat and dhcp)",
			properties: `"ntpProperties":{"acquisitionProperties":{"ntpAcquisitionType":"disabled","ntpServers":[{"addrType":"ipvx","ipvxAddress":{"addressType":"ipv4","ipv4Address":"10.192.88.175"}}]}}`,
			wantLabels: map[string]string{
				"ntp_service":        "disabled",
				"primary_ntp_server": "",
				"backup_ntp_server":  "",
			},
		},
		{
			name:       "unknown dns: populated server list still yields no exported addresses (whitelist covers only stat and dhcp)",
			properties: `"dnsProperties":{"acquisitionProperties":{"dnsAcquisitionType":"unknown","dnsServers":[{"addressType":"ipv4","ipv4Address":"10.192.0.250"}]}}`,
			wantLabels: map[string]string{
				"dns_config_method":  "unknown",
				"primary_dns_server": "",
				"backup_dns_server":  "",
			},
		},
		{
			name:       "missing dnsProperties and ntpProperties: empty labels, no panic",
			properties: `"linkStatus":"up"`,
			wantLabels: map[string]string{
				"dns_config_method":  "",
				"primary_dns_server": "",
				"backup_dns_server":  "",
				"ntp_service":        "",
				"primary_ntp_server": "",
				"backup_ntp_server":  "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHardware()
			h.processNetInterfaces(netInterfaceController(tt.properties), "ctrl1", "A")

			inst := h.data[netInterfaceMatrix].GetInstance("ctrl1_wan0")
			if inst == nil {
				t.Fatalf("instance %q not found", "ctrl1_wan0")
			}
			for label, want := range tt.wantLabels {
				if got := inst.GetLabel(label); got != want {
					t.Errorf("label %q = %q, want %q", label, got, want)
				}
			}
		})
	}
}

func TestIpvxAddress(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		wantAddress string
		wantType    string
	}{
		{"lowercase ipv4", `{"addressType":"ipv4","ipv4Address":"10.0.0.1"}`, "10.0.0.1", "ipv4"},
		{"mixed-case IPv4 is matched case-insensitively and normalized on output", `{"addressType":"IPv4","ipv4Address":"10.0.0.1"}`, "10.0.0.1", "ipv4"},
		{"uppercase IPV6 is matched case-insensitively and normalized on output", `{"addressType":"IPV6","ipv6Address":"::1"}`, "::1", "ipv6"},
		{"unrecognized addressType returns empty", `{"addressType":"bogus"}`, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAddress, gotType := ipvxAddress(gjson.Parse(tt.json))
			if gotAddress != tt.wantAddress || gotType != tt.wantType {
				t.Errorf("ipvxAddress(%s) = (%q, %q), want (%q, %q)", tt.json, gotAddress, gotType, tt.wantAddress, tt.wantType)
			}
		})
	}
}

func TestNtpServerAddress(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		wantAddress string
		wantType    string
	}{
		{"lowercase domainName", `{"addrType":"domainName","domainName":"time.nist.gov"}`, "time.nist.gov", "domainName"},
		{"mixed-case DomainName is matched case-insensitively", `{"addrType":"DomainName","domainName":"time.nist.gov"}`, "time.nist.gov", "domainName"},
		{"uppercase IPVX is matched case-insensitively", `{"addrType":"IPVX","ipvxAddress":{"addressType":"ipv4","ipv4Address":"10.0.0.1"}}`, "10.0.0.1", "ipv4"},
		{"ipvx branch normalizes the inner addressType too, not just its own addrType", `{"addrType":"ipvx","ipvxAddress":{"addressType":"IPv4","ipv4Address":"10.0.0.1"}}`, "10.0.0.1", "ipv4"},
		{"unrecognized addrType returns empty", `{"addrType":"none"}`, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAddress, gotType := ntpServerAddress(gjson.Parse(tt.json))
			if gotAddress != tt.wantAddress || gotType != tt.wantType {
				t.Errorf("ntpServerAddress(%s) = (%q, %q), want (%q, %q)", tt.json, gotAddress, gotType, tt.wantAddress, tt.wantType)
			}
		})
	}
}
