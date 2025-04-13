package parser

import (
	"github.com/LicorneSharing/GTL/optional"
	tassert "github.com/stretchr/testify/assert"
	"testing"
)

func Test_assert(t *testing.T) {
	type testcase struct {
		name     string
		cond     bool
		msg      string
		formats  []any
		panicMsg optional.Value[string]
	}

	for _, tt := range []testcase{
		{
			name:     "panic with simple message",
			cond:     false,
			msg:      "message",
			panicMsg: optional.Some(": message"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				var (
					recovered    = recover()
					panicMsg, ok = tt.panicMsg.LookupValue()
				)

				tassert.Equal(t, ok, recovered != nil)

				if !ok {
					return
				}

				tassert.Equal(t, "INVALID CALL TO FUNCTION"+panicMsg, recovered)
			}()

			assert(tt.cond, append([]any{tt.msg}, tt.formats...)...)
		})
	}
}
