package scanner

import (
	"testing"
)

func TestScanCode(t *testing.T) {
	t.Run("literals value", func(t *testing.T) {
		got := ScanCode(`18; "coucou" 42.1(8); -0b01; +0xA4`).String()
		expected := `18 ; "coucou" 42.1(8) ; -1 ; 20`
		if got != expected {
			t.Errorf("got: %v\nexpected: %v", got, expected)
		}
	})
}
