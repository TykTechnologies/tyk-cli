package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Verifies: SYS-REQ-031
func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"plain seconds", "60", 60},
		{"seconds suffix", "60s", 60},
		{"minutes", "1m", 60},
		{"five minutes", "5m", 300},
		{"one hour", "1h", 3600},
		{"twenty-four hours", "24h", 86400},
		{"one day", "1d", 86400},
		{"thirty days", "30d", 2592000},
		{"zero", "0", 0},
		{"zero suffix", "0s", 0},
		{"large seconds", "2592000", 2592000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Verifies: SYS-REQ-031
func TestParseDuration_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"alphabetic only", "abc"},
		{"negative", "-1"},
		{"negative with suffix", "-5m"},
		{"fractional", "1.5h"},
		{"mixed units", "1h30m"},
		{"empty string", ""},
		{"unsupported suffix", "30w"},
		{"unsupported suffix y", "1y"},
		{"spaces", "30 d"},
		{"suffix only", "m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseDuration(tt.input)
			assert.Error(t, err, "ParseDuration(%q) should return error", tt.input)
		})
	}
}

// Verifies: SYS-REQ-031
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int64
		expected string
	}{
		{"zero", 0, "0"},
		{"exact days", 2592000, "30d"},
		{"one day", 86400, "1d"},
		{"exact hours", 3600, "1h"},
		{"two hours", 7200, "2h"},
		{"exact minutes", 60, "1m"},
		{"five minutes", 300, "5m"},
		{"odd seconds", 45, "45s"},
		{"ninety seconds", 90, "90s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.seconds)
			assert.Equal(t, tt.expected, result)
		})
	}
}
