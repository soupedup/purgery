// Package safe implements secure functionality.
package safe

import (
	"crypto/subtle"
	"math"
)

// Compare tries to compares the given strings in constant time.
//
// Compare always returns false in case any of the given strings is longer than
// math.MaxInt32 bytes.
func Compare(expected, actual string) bool {
	if !haveEqualLen(expected, actual) {
		return !areEqual(actual, actual)
	}

	return areEqual(expected, actual)
}

func haveEqualLen(expected, given string) bool {
	const max = math.MaxInt32

	switch {
	case len(expected) > max, len(given) > max:
		return false
	case subtle.ConstantTimeEq(int32(len(expected)), int32(len(given))) != 1:
		return false
	default:
		return true
	}
}

func areEqual(x, y string) bool {
	if len(x) != len(y) {
		return false
	}

	var v byte
	for i := 0; i < len(x); i++ {
		v |= x[i] ^ y[i]
	}

	return subtle.ConstantTimeByteEq(v, 0) == 1
}
