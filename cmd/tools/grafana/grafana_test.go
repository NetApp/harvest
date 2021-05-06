package main

import (
	"testing"
)

func TestCheckVersion(t *testing.T) {

	inputVersion := []string{"7.2.3.4", "abc.1.3", "4.5.4", "7.1.0", "7.5.5"}
	expectedOutPut := []bool{true, false, false, true, true}
	// version length greater than 3

	for i, s := range inputVersion {
		c := checkVersion(s)
		if c != expectedOutPut[i] {
			t.Errorf("Expected %t but got %t for input %s", expectedOutPut[i], c, inputVersion[i])
		}
	}
}
