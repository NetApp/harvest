package power

import "strings"

// Rail identifies a power rail classification inferred from sensor labels.
type Rail uint8

const (
	RailUnknown Rail = iota
	RailInput
	RailOutput
)

// RailFromLabel infers a rail type from a single label string.
func RailFromLabel(label string) Rail {
	if label == "" {
		return RailUnknown
	}
	lower := strings.ToLower(label)
	hasInput := strings.Contains(lower, "input") || strings.Contains(lower, "vin") || strings.Contains(lower, "iin")
	hasOutput := strings.Contains(lower, "output") || strings.Contains(lower, "vout") || strings.Contains(lower, "iout")
	switch {
	case hasInput && hasOutput:
		return RailUnknown
	case hasInput:
		return RailInput
	case hasOutput:
		return RailOutput
	default:
		return RailUnknown
	}
}

// ClassifyRailFromLabels returns the first non-unknown rail from the provided labels.
func ClassifyRailFromLabels(labels ...string) Rail {
	for _, label := range labels {
		if rail := RailFromLabel(label); rail != RailUnknown {
			return rail
		}
	}
	return RailUnknown
}

// ResolveRail reconciles voltage/current rail classifications into a single rail.
func ResolveRail(voltageRail, currentRail Rail) Rail {
	if voltageRail != RailUnknown && currentRail != RailUnknown && voltageRail != currentRail {
		return RailUnknown
	}
	if voltageRail != RailUnknown {
		return voltageRail
	}
	if currentRail != RailUnknown {
		return currentRail
	}
	return RailUnknown
}
