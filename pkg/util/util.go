/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package util

func MinLen(elements [][]string) int {
	var min, i int
	min = len(elements[0])
	for i = 1; i < len(elements); i += 1 {
		if len(elements[i]) < min {
			min = len(elements[i])
		}
	}
	return min
}

func MaxLen(elements [][]string) int {
	var max, i int
	max = len(elements[0])
	for i = 1; i < len(elements); i += 1 {
		if len(elements[i]) > max {
			max = len(elements[i])
		}
	}
	return max
}

func AllSame(elements [][]string, k int) bool {
	var i int
	for i = 1; i < len(elements); i += 1 {
		if elements[i][k] != elements[0][k] {
			return false
		}
	}
	return true
}

func EqualStringSlice(a, b []string) bool {
	var i int
	if len(a) != len(b) {
		return false
	}
	for i = 0; i < len(a); i += 1 {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
