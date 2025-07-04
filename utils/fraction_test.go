package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeFractionFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected Fraction
	}{
		{"0.(3)", MakeFraction(1, 3)},
		{"0.(6)", MakeFraction(2, 3)},
		{"1.2(3)", MakeFraction(37, 30)},
		{"0.25", MakeFraction(1, 4)},
		{"1.0(6)", MakeFraction(32, 30)},
		{"2.1(6)", MakeFraction(13, 6)},
		{"0.1(6)", MakeFraction(1, 6)},
	}

	for _, tt := range tests {
		got, err := MakeFractionFromString(tt.input)
		assert.NoError(t, err, "Erreur inattendue pour %q", tt.input)
		assert.Equal(t, tt.expected, got, "Résultat inattendu pour %q", tt.input)
	}
}
