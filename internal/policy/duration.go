package policy

import (
	"fmt"
	"strconv"
	"strings"
)

// Duration unit constants in seconds.
const (
	secondsPerMinute = 60
	secondsPerHour   = 3600
	secondsPerDay    = 86400
)

// suffixMultipliers maps duration suffixes to their multiplier in seconds.
var suffixMultipliers = map[byte]int64{
	's': 1,
	'm': secondsPerMinute,
	'h': secondsPerHour,
	'd': secondsPerDay,
}

// ParseDuration parses a duration string into seconds.
// Accepted formats: plain integer ("60"), or integer with suffix s/m/h/d ("30d", "1h").
// Rejects negative values, fractional values, mixed units, and unsupported suffixes.
func ParseDuration(s string) (int64, error) {
	if s == "" {
		return 0, fmt.Errorf("invalid duration %q: empty string", s)
	}

	// Check for spaces
	if strings.ContainsAny(s, " \t") {
		return 0, fmt.Errorf("invalid duration %q: must not contain spaces", s)
	}

	// Check for negative
	if s[0] == '-' {
		return 0, fmt.Errorf("invalid duration %q: negative values not allowed", s)
	}

	// Check for fractional (contains '.')
	if strings.Contains(s, ".") {
		return 0, fmt.Errorf("invalid duration %q: fractional values not allowed", s)
	}

	last := s[len(s)-1]
	multiplier, hasSuffix := suffixMultipliers[last]

	if hasSuffix {
		numPart := s[:len(s)-1]
		if numPart == "" {
			return 0, fmt.Errorf("invalid duration %q: missing numeric value", s)
		}
		n, err := strconv.ParseInt(numPart, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q: %w", s, err)
		}
		return n * multiplier, nil
	}

	// No recognized suffix -- try plain integer
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: expected integer or NNs/NNm/NNh/NNd", s)
	}
	return n, nil
}

// FormatDuration converts seconds to the largest clean human-readable unit.
// 0 -> "0", 86400 -> "1d", 3600 -> "1h", 60 -> "1m", 45 -> "45s".
func FormatDuration(seconds int64) string {
	if seconds == 0 {
		return "0"
	}
	if seconds%secondsPerDay == 0 {
		return fmt.Sprintf("%dd", seconds/secondsPerDay)
	}
	if seconds%secondsPerHour == 0 {
		return fmt.Sprintf("%dh", seconds/secondsPerHour)
	}
	if seconds%secondsPerMinute == 0 {
		return fmt.Sprintf("%dm", seconds/secondsPerMinute)
	}
	return fmt.Sprintf("%ds", seconds)
}
