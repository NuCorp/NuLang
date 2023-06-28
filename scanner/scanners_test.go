package scanner

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
	"testing"
)

func TestScanCode(t *testing.T) {
	run := func(code, expected string, tokenList ...tokens.Token) func(t *testing.T) {
		return func(t *testing.T) {
			scanCode := ScanCode(code)
			got := scanCode.String()
			if got != expected {
				t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
			}
			if tokenList != nil {
				for i, token := range scanCode.TokenList() {
					if i >= len(tokenList) || tokenList[i] != token {
						t.Errorf("invalid token\ngot: %v\nexpected: %v\ndiff at %v: %v -> %v", scanCode.TokenList(), tokenList, i, token, tokenList[i])
					}
				}
			}
		}
	}
	// literals
	// integers
	t.Run("simple integer 1", func(t *testing.T) {
		code := "18"
		got := ScanCode(code).String()
		expected := "18 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("simple integer 2", func(t *testing.T) {
		code := "31"
		got := ScanCode(code).String()
		expected := "31 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("simple integer 3", func(t *testing.T) {
		code := "42"
		got := ScanCode(code).String()
		expected := "42 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("simple integer 4", func(t *testing.T) {
		code := "23"
		got := ScanCode(code).String()
		expected := "23 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("binary format", func(t *testing.T) {
		code := "0b010"
		got := ScanCode(code).String()
		expected := "2 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("octal format", func(t *testing.T) {
		code := "0o70"
		got := ScanCode(code).String()
		expected := "56 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("hex format", func(t *testing.T) {
		code := "0x0A0"
		got := ScanCode(code).String()
		expected := "160 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("negative integer", func(t *testing.T) {
		code := "-64"
		got := ScanCode(code).String()
		expected := "-64 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
	t.Run("negative hexadecimal integer", func(t *testing.T) {
		code := "-0x2A"
		got := ScanCode(code).String()
		expected := "-42 "
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})

	// floating point
	t.Run("simple float 1", run("18.02", "18.02 ", tokens.FLOAT))
	t.Run("simple float 2", run("-31.08", "-31.08 ", tokens.FLOAT))

	// fraction
	t.Run("simple fraction 1", run("1.0(3)", "1.0(3) ", tokens.FRACTION))
	t.Run("simple fraction 2", run("1.(3)", "1.(3) ", tokens.FRACTION))
	t.Run("simple fraction 3", run("-1.(53)", "-1.(53) ", tokens.FRACTION))

	t.Run("simple literals value", func(t *testing.T) {
		code := `18; "coucou" 42.1(8); -0b01; +0xA4`
		got := ScanCode(code).String()
		expected := `18 ; "coucou" 42.1(8) ; -1 ; 20 `
		if got != expected {
			t.Errorf("initial value: %v\ngot: %v\nexpected: %v", code, got, expected)
		}
	})
}
