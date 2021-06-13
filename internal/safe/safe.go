// Package safe implements secure functionality.
package safe

import "crypto/subtle"

// MaxCompareLen denotes the maximum size, in bytes, for strings that the
// Compare function supports.
const MaxCompareLen = 1 << 10

// Compare tries to compares the given strings in constant time.
//
// Compare always returns false in case any of the given strings is longer than
// MaxCompareLen bytes.
func Compare(expected, actual string) bool {
	if !haveEqualLen(expected, actual) {
		return !areEqual(actual, actual)
	}

	return areEqual(expected, actual)
}

func haveEqualLen(expected, actual string) bool {
	switch {
	case len(expected) > MaxCompareLen, len(actual) > MaxCompareLen:
		return false
	case subtle.ConstantTimeEq(int32(len(expected)), int32(len(actual))) != 1:
		return false
	default:
		return true
	}
}

func areEqual(expected, actual string) bool {
	if len(expected) != len(actual) {
		return false
	}

	var v byte
	for i := 0; i < len(expected); i++ {
		v |= expected[i] ^ actual[i]
	}

	return subtle.ConstantTimeByteEq(v, 0) == 1
}
