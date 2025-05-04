package ip

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type TCPPortRange struct {
	Start uint16
	End   uint16
}

// String returns the string representation of the TCP port range.
//
// Returns:
//
//	string: The string representation of the TCP port range.
func (t *TCPPortRange) String() string {
	return fmt.Sprintf("%d-%d", t.Start, t.End)
}

// NewTCPPortRange creates a new TCP port range instance.
//
// Parameters:
//
//	start: The start port number.
//	end: The end port number.
//
// Returns:
//
//	*TCPPortRange: A new TCP port range instance.
func NewTCPPortRange(start, end uint16) *TCPPortRange {
	return &TCPPortRange{Start: start, End: end}
}

// NewTCPPortRangeFromString creates a new TCP port range instance from a string.
//
// Parameters:
//
//	s: The string representation of the TCP port range.
//
// Returns:
//
//	*TCPPortRange: A new TCP port range instance.
func NewTCPPortRangeFromString(s string) (*TCPPortRange, error) {
	var err error

	if !regexp.MustCompile(`^\s*(?:[0-9]|[1-9]\d{1,3}|[1-5]\d{4}|6[0-4]\d{3}|65[0-4]\d{2}|655[0-2]\d|6553[0-5])\s*-\s*(?:[0-9]|[1-9]\d{1,3}|[1-5]\d{4}|6[0-4]\d{3}|65[0-4]\d{2}|655[0-2]\d|6553[0-5])\s*$`).MatchString(s) {
		return nil, fmt.Errorf("invalid port range format")
	}

	parts := strings.Split(s, "-")

	if len(parts) == 2 {
		start := uint64(0)
		if len(parts[0]) > 0 {
			start, err = strconv.ParseUint(parts[0], 10, 16)
			if err != nil {
				return nil, err
			}

			if start > 65535 {
				return nil, fmt.Errorf("port range start must be between 0 and 65535")
			}
		}

		end := uint64(65535)
		if len(parts[1]) > 0 {
			end, err = strconv.ParseUint(parts[1], 10, 16)
			if err != nil {
				return nil, err
			}

			if end > 65535 {
				return nil, fmt.Errorf("port range end must be between 0 and 65535")
			}
		}

		return NewTCPPortRange(uint16(start), uint16(end)), nil
	}

	return nil, fmt.Errorf("invalid port range format")
}
