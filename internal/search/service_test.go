package search

import (
	"testing"
)

func TestCleanLocationQuery(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"bisa kamu searching, cuaca dibali hari ini", "bali"},
		{"cuaca di bali hari ini", "bali"},
		{"weather in tokyo today", "tokyo"},
	}

	for _, tt := range tests {
		got := cleanLocationQuery(tt.input)
		if got != tt.expected {
			t.Errorf("cleanLocationQuery(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}
