package hardware

import "testing"

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
