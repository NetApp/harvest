package power

import (
	"strconv"
	"strings"
)

// NeedsPsuPowerDrawnCorrection reports whether the PSU power-drawn value
// reported by ONTAP must be divided by 8 to obtain actual watts.
//
// ONTAP Bug 1372079: IOM12E shelves with firmware < 0240 and IOM12 shelves
// with firmware < 0290 report psu-power-drawn in 1/8-watt units instead of
// whole watts. The firmware fix converts the raw value before reporting it,
// so older firmware requires a client-side divide-by-8 correction.
// https://mysupport.netapp.com/site/bugs-online/product/ONTAP/BURT/1372079
//
// Affected hardware: IOM12E (embedded shelves on AFF A220/A200/C190,
// FAS2750/2720/2650/2620) and IOM12 (external shelf modules on DS224C,
// DS212C, DS460C).
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
