package commands

import "math"

// wireInt converts a 32-bit wire length or offset to int. On 32-bit platforms
// int(v) wraps values of 2 GiB and above to negative numbers, which slip past
// the "> 0" presence tests and upper-bound checks; such values are saturated
// to math.MaxInt instead, which every bounds check then rejects.
func wireInt(v uint32) int {
	if uint64(v) > math.MaxInt {
		return math.MaxInt
	}
	return int(v)
}
