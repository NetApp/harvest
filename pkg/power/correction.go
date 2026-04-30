package power

import (
	"strconv"
	"strings"
)

// NeedsPsuPowerDrawnCorrection https://mysupport.netapp.com/site/bugs-online/product/ONTAP/BURT/1372079
func NeedsPsuPowerDrawnCorrection(moduleType, firmwareVersion string) bool {
	if firmwareVersion == "" {
		return false
	}
	fw, err := strconv.Atoi(strings.TrimSpace(firmwareVersion))
	if err != nil {
		return false
	}
	switch strings.ToUpper(strings.TrimSpace(moduleType)) {
	case "IOM12E":
		return fw < 240
	case "IOM12":
		return fw < 290
	default:
		return false
	}
}

// MinFirmwareVersion returns the numerically lower of two firmware version
// strings. If either is empty or not parseable as an integer, the other is
// returned. If both are unparseable, b is returned.
func MinFirmwareVersion(a, b string) string {
	aInt, aErr := strconv.Atoi(strings.TrimSpace(a))
	bInt, bErr := strconv.Atoi(strings.TrimSpace(b))
	switch {
	case aErr != nil:
		return b
	case bErr != nil:
		return a
	case aInt <= bInt:
		return a
	default:
		return b
	}
}
