package scanner

import (
	"testing"
)

func TestScanCode(t *testing.T) {
	// literals
	t.Run("4 simple integer", func(t *testing.T) {
		{
			code := "18"
			got := ScanCode(code).String()
			expected := "18 "
			if got != expected {
				t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
			}
		}
		{
			code := "42"
			got := ScanCode(code).String()
			expected := "42 "
			if got != expected {
				t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
			}
		}
		{
			code := "23"
			got := ScanCode(code).String()
			expected := "23 "
			if got != expected {
				t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
			}
		}
		{
			code := "31"
			got := ScanCode(code).String()
			expected := "31 "
			if got != expected {
				t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
			}
		}

	})
	t.Run("simple literals value", func(t *testing.T) {
		code := `18; "coucou" 42.1(8); -0b01; +0xA4`
		got := ScanCode(code).String()
		expected := `18 ; "coucou" 42.1(8) ; -1 ; 20 `
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
}
