package num

import (
	"fmt"
	"math"
	"slices"
	"strconv"
)

func SumNumbers(s []float64) float64 {
	var total float64
	for _, num := range s {
		total += num
	}
	return total
}

// Max returns 0 when passed an empty slice, slices.Max panics if input is empty,
// This function can be removed once all callers are checked for empty slices
func Max(input []float64) float64 {
	if len(input) > 0 {
		return slices.Max(input)
	}
	return 0
}

// Min returns 0 when passed an empty slice, slices.Min panics if input is empty,
// This function can be removed once all callers are checked for empty slices
func Min(input []float64) float64 {
	if len(input) > 0 {
		return slices.Min(input)
	}
	return 0
}

func Avg(input []float64) float64 {
	if len(input) > 0 {
		return SumNumbers(input) / float64(len(input))
	}
	return 0
}

func AddIntString(input string, value int) string {
	i, _ := strconv.Atoi(input)
	i += value
	return strconv.FormatInt(int64(i), 10)
}

func SafeConvertToInt32(in int) (int32, error) {
	if in > math.MaxInt32 {
		return 0, fmt.Errorf("input %d is too large to convert to int32", in)
	}
	return int32(in), nil // #nosec G115
}
