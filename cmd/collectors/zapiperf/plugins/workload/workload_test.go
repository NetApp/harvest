package workload

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"testing"
)

func Test_calculateIOPS(t *testing.T) {
	testCases := []struct {
		policy       *collectors.QosAdaptive
		sizeTB       float64
		sizeUsedTB   float64
		expectedIOPS float64
	}{
		// test cases with expectedIOPSAllocation as "used_space" and peakIOPSAllocation as "allocated_space"
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "used_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "allocated_space", BlockSize: "any", Svm: "svm1"}, 0.001, 0, 75},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "used_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "allocated_space", BlockSize: "any", Svm: "svm1"}, 1, 0, 128},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "used_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "allocated_space", BlockSize: "any", Svm: "svm1"}, 1, 0.1, 128},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "used_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "allocated_space", BlockSize: "any", Svm: "svm1"}, 1, 0.2, 128},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "used_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "allocated_space", BlockSize: "any", Svm: "svm1"}, 1, 0.3, 154},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "used_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "allocated_space", BlockSize: "any", Svm: "svm1"}, 1, 0.5, 256},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "used_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "allocated_space", BlockSize: "any", Svm: "svm1"}, 1, 1, 512},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "used_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "allocated_space", BlockSize: "any", Svm: "svm1"}, 2, 2, 1024},

		// test cases with expectedIOPSAllocation as "allocated_space" and peakIOPSAllocation as "allocated_space"
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "allocated_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "allocated_space", BlockSize: "any", Svm: "svm1"}, 0.001, 0, 75},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "allocated_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "allocated_space", BlockSize: "any", Svm: "svm1"}, 1, 0, 512},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "allocated_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "allocated_space", BlockSize: "any", Svm: "svm1"}, 1, 0.1, 512},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "allocated_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "allocated_space", BlockSize: "any", Svm: "svm1"}, 1, 0.2, 512},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "allocated_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "allocated_space", BlockSize: "any", Svm: "svm1"}, 1, 0.3, 512},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "allocated_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "allocated_space", BlockSize: "any", Svm: "svm1"}, 1, 0.5, 512},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "allocated_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "allocated_space", BlockSize: "any", Svm: "svm1"}, 1, 1, 512},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "allocated_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "allocated_space", BlockSize: "any", Svm: "svm1"}, 2, 2, 1024},

		// test cases with expectedIOPSAllocation as "used_space" and peakIOPSAllocation as "used_space"
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "used_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "used_space", BlockSize: "any", Svm: "svm1"}, 0.001, 0, 75},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "used_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "used_space", BlockSize: "any", Svm: "svm1"}, 1, 0, 75},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "used_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "used_space", BlockSize: "any", Svm: "svm1"}, 1, 0.1, 75},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "used_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "used_space", BlockSize: "any", Svm: "svm1"}, 1, 0.2, 102},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "used_space", ExpectedIOPS: "50", ExpectedIOPSAllocation: "used_space", BlockSize: "any", Svm: "svm1"}, 1.5, 0.3, 154},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "175", PeakIOPS: "512", PeakIOPSAllocation: "used_space", ExpectedIOPS: "250", ExpectedIOPSAllocation: "used_space", BlockSize: "any", Svm: "svm1"}, 1, 0.3, 175},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "used_space", ExpectedIOPS: "250", ExpectedIOPSAllocation: "used_space", BlockSize: "any", Svm: "svm1"}, 1, 0.5, 256},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "used_space", ExpectedIOPS: "250", ExpectedIOPSAllocation: "used_space", BlockSize: "any", Svm: "svm1"}, 1, 1, 512},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "used_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "used_space", BlockSize: "any", Svm: "svm1"}, 2, 2, 1024},

		// test cases with expectedIOPSAllocation as "allocated_space" and peakIOPSAllocation as "used_space"
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "allocated_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "used_space", BlockSize: "any", Svm: "svm1"}, 0.001, 0, 75},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "allocated_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "used_space", BlockSize: "any", Svm: "svm1"}, 1, 0, 512},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "allocated_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "used_space", BlockSize: "any", Svm: "svm1"}, 1, 0.1, 512},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "allocated_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "used_space", BlockSize: "any", Svm: "svm1"}, 1, 0.2, 512},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "allocated_space", ExpectedIOPS: "250", ExpectedIOPSAllocation: "used_space", BlockSize: "any", Svm: "svm1"}, 1, 0.3, 512},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "allocated_space", ExpectedIOPS: "250", ExpectedIOPSAllocation: "used_space", BlockSize: "any", Svm: "svm1"}, 1, 0.5, 512},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "allocated_space", ExpectedIOPS: "250", ExpectedIOPSAllocation: "used_space", BlockSize: "any", Svm: "svm1"}, 1, 1, 512},
		{&collectors.QosAdaptive{PolicyGroup: "group1", AbsoluteMinIOPS: "75", PeakIOPS: "512", PeakIOPSAllocation: "allocated_space", ExpectedIOPS: "128", ExpectedIOPSAllocation: "used_space", BlockSize: "any", Svm: "svm1"}, 2, 2, 1024},
	}

	// Run test cases
	for _, tc := range testCases {
		peakAllowedIOPS, err := collectors.CalculateIOPS(tc.policy, tc.sizeTB, tc.sizeUsedTB)
		if err != nil {
			t.Errorf("calculateIOPS() with input %v returned an error: %v", *tc.policy, err)
		}
		if peakAllowedIOPS != tc.expectedIOPS {
			t.Errorf("calculateIOPS() with input PolicyGroup: %s, AbsoluteMinIOPS: %s, PeakIOPS: %s, PeakIOPSAllocation: %s, ExpectedIOPS: %s, ExpectedIOPSAllocation: %s, BlockSize: %s, SVM: %s, SizeTB: %f, SizeUsedTB: %f expected %f but got %f",
				tc.policy.PolicyGroup,
				tc.policy.AbsoluteMinIOPS,
				tc.policy.PeakIOPS,
				tc.policy.PeakIOPSAllocation,
				tc.policy.ExpectedIOPS,
				tc.policy.ExpectedIOPSAllocation,
				tc.policy.BlockSize,
				tc.policy.Svm,
				tc.sizeTB,
				tc.sizeUsedTB,
				tc.expectedIOPS,
				peakAllowedIOPS,
			)
		}
	}
}
