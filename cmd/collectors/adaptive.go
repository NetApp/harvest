package collectors

import (
	"github.com/netapp/harvest/v2/pkg/util"
	"math"
)

type QosAdaptive struct {
	PolicyGroup            string
	AbsoluteMinIOPS        string
	PeakIOPS               string
	PeakIOPSAllocation     string
	ExpectedIOPS           string
	ExpectedIOPSAllocation string
	BlockSize              string
	Svm                    string
}

func CalculateIOPS(policy *QosAdaptive, sizeTB, sizeUsedTB float64) (float64, error) {
	var absoluteMinIOPS, peakAllowedIOPS, expectedIOPS float64

	peakIOPS, err := util.ConvertStringToFloat64(policy.PeakIOPS)
	if err != nil {
		return 0, err
	}
	absoluteMinIOPS, err = util.ConvertStringToFloat64(policy.AbsoluteMinIOPS)
	if err != nil {
		return 0, err
	}

	expectedIOPS, err = util.ConvertStringToFloat64(policy.ExpectedIOPS)
	if err != nil {
		return 0, err
	}

	if policy.ExpectedIOPSAllocation == "used_space" {
		absoluteMinIOPS = max(absoluteMinIOPS, sizeUsedTB*expectedIOPS)
	} else {
		absoluteMinIOPS = max(absoluteMinIOPS, sizeTB*expectedIOPS)
	}

	if policy.PeakIOPSAllocation == "used_space" {
		peakAllowedIOPS = max(absoluteMinIOPS, sizeUsedTB*peakIOPS)
	} else {
		peakAllowedIOPS = max(absoluteMinIOPS, sizeTB*peakIOPS)
	}

	return math.Round(peakAllowedIOPS), nil
}

func CalculateThroughputPercent(totalDataBps, blockSize, peakIOPS float64) float64 {
	if totalDataBps == 0 {
		return 0
	}
	peakAllowedMbps := (peakIOPS * blockSize) / 1000
	return totalDataBps / (peakAllowedMbps * 1_000_000)
}

func CalculateUsedPercent(ops, peakAllowedIops float64) float64 {
	if ops == 0 {
		return 0
	}
	return (ops * 100) / peakAllowedIops
}
