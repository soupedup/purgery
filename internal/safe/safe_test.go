package safe

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompare(t *testing.T) {
	long := strings.Repeat("a", MaxCompareLen+1)

	cases := []struct {
		a   string
		b   string
		exp bool
	}{
		0: {exp: true},
		1: {"a", "a", true},
		2: {"a", "b", false},
		3: {"a", "aa", false},
		4: {"a", "long", false},
		5: {"long", "a", false},
		6: {long, long, false},
	}

	for caseIndex := range cases {
		kase := cases[caseIndex]

		t.Run(strconv.Itoa(caseIndex), func(t *testing.T) {
			assert.Equal(t, kase.exp, Compare(kase.a, kase.b))
		})
	}
}
