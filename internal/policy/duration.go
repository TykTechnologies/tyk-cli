package policy

import (
	"fmt"
	"strconv"
	"strings"
)

// suffixMultipliers maps duration suffixes to their multiplier in seconds.
var suffixMultipliers = map[byte]int64{
	's': 1,
	'm': 60,
	'h': 3600,
	'd': 86400,
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
	if seconds%86400 == 0 {
		return fmt.Sprintf("%dd", seconds/86400)
	}
	if seconds%3600 == 0 {
		return fmt.Sprintf("%dh", seconds/3600)
	}
	if seconds%60 == 0 {
		return fmt.Sprintf("%dm", seconds/60)
	}
	return fmt.Sprintf("%ds", seconds)
}
