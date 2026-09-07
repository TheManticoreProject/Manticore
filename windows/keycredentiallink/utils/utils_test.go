package utils_test

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/source"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/utils"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/version"
)

func TestConvertFromBinaryIdentifier(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		version  version.KeyCredentialLinkVersion
		expected string
	}{
		{
			name:     "Version 0 hex encoding",
			input:    []byte{0x12, 0x34, 0x56},
			version:  version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_0},
			expected: "123456",
		},
		{
			name:     "Version 1 hex encoding",
			input:    []byte{0x12, 0x34, 0x56},
			version:  version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_1},
			expected: "123456",
		},
		{
			name:     "Version 2 base64 encoding",
			input:    []byte{0x12, 0x34, 0x56},
			version:  version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2},
			expected: base64.StdEncoding.EncodeToString([]byte{0x12, 0x34, 0x56}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := utils.ConvertFromBinaryIdentifier(tc.input, tc.version)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestConvertFromBinaryTime(t *testing.T) {
	// Create test timestamp (2022-03-15 12:00:03 UTC)
	testTimeBytes := []byte{0x80, 0xa3, 0x22, 0x34, 0x64, 0x38, 0xd8, 0x01}
	testTimeStruct := time.Date(2022, 3, 15, 12, 0, 3, 0, time.UTC)

	testCases := []struct {
		name     string
		input    []byte
		source   source.KeySource
		version  version.KeyCredentialLinkVersion
		expected time.Time
	}{
		{
			name:     "Version 0 AD source",
			input:    testTimeBytes,
			source:   source.KeySource{Value: source.KeySource_AD},
			version:  version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_0},
			expected: testTimeStruct,
		},
		{
			name:     "Version 1 AD source",
			input:    testTimeBytes,
			source:   source.KeySource{Value: source.KeySource_AD},
			version:  version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_1},
			expected: testTimeStruct,
		},
		{
			name:     "Version 2 AD source",
			input:    testTimeBytes,
			source:   source.KeySource{Value: source.KeySource_AD},
			version:  version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2},
			expected: testTimeStruct,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := utils.ConvertFromBinaryTime(tc.input, tc.source, tc.version)
			if !result.GetTime().Equal(tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result.GetTime())
			}
		})
	}
}

// The encoder has to emit the FILETIME the field holds. This test previously called
// only ConvertFromBinaryTime despite its name, which is why the encoder emitting
// Unix nanoseconds went unnoticed.
func TestConvertToBinaryTime(t *testing.T) {
	// 2022-03-15 12:00:03 UTC, and the bytes Windows writes for it.
	testTimeStruct := time.Date(2022, 3, 15, 12, 0, 3, 0, time.UTC)
	testTimeBytes := []byte{0x80, 0xa3, 0x22, 0x34, 0x64, 0x38, 0xd8, 0x01}

	testCases := []struct {
		name    string
		source  source.KeySource
		version version.KeyCredentialLinkVersion
	}{
		{
			name:    "Version 0 AD source",
			source:  source.KeySource{Value: source.KeySource_AD},
			version: version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_0},
		},
		{
			name:    "Version 1 AD source",
			source:  source.KeySource{Value: source.KeySource_AD},
			version: version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_1},
		},
		{
			name:    "Version 2 AD source",
			source:  source.KeySource{Value: source.KeySource_AD},
			version: version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded := utils.ConvertToBinaryTime(testTimeStruct, tc.source, tc.version)

			if !bytes.Equal(encoded, testTimeBytes) {
				t.Errorf("ConvertToBinaryTime() = %s, want %s (the FILETIME Windows writes)",
					hex.EncodeToString(encoded), hex.EncodeToString(testTimeBytes))
			}

			// The two functions are named as inverses and have to behave as such.
			decoded := utils.ConvertFromBinaryTime(encoded, tc.source, tc.version)
			if !decoded.GetTime().Equal(testTimeStruct) {
				t.Errorf("round-trip gave %v, want %v", decoded.GetTime(), testTimeStruct)
			}
		})
	}
}

func TestBinaryTimeInvolution(t *testing.T) {
	testCases := []struct {
		name    string
		time    time.Time
		source  source.KeySource
		version version.KeyCredentialLinkVersion
	}{
		{
			name:    "Version 0 AD source - Current time",
			time:    time.Now().UTC(),
			source:  source.KeySource{Value: source.KeySource_AD},
			version: version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_0},
		},
		{
			name:    "Version 1 AD source - Current time",
			time:    time.Now().UTC(),
			source:  source.KeySource{Value: source.KeySource_AD},
			version: version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_1},
		},
		{
			name:    "Version 2 AD source - Current time",
			time:    time.Now().UTC(),
			source:  source.KeySource{Value: source.KeySource_AD},
			version: version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2},
		},
		{
			name:    "Version 0 AD source - Past time",
			time:    time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
			source:  source.KeySource{Value: source.KeySource_AD},
			version: version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_0},
		},
		{
			name:    "Version 1 AD source - Past time",
			time:    time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
			source:  source.KeySource{Value: source.KeySource_AD},
			version: version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_1},
		},
		{
			name:    "Version 2 AD source - Past time",
			time:    time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
			source:  source.KeySource{Value: source.KeySource_AD},
			version: version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2},
		},
		{
			name:    "Version 0 AD source - Future time",
			time:    time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			source:  source.KeySource{Value: source.KeySource_AD},
			version: version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_0},
		},
		{
			name:    "Version 1 AD source - Future time",
			time:    time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			source:  source.KeySource{Value: source.KeySource_AD},
			version: version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_1},
		},
		{
			name:    "Version 2 AD source - Future time",
			time:    time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			source:  source.KeySource{Value: source.KeySource_AD},
			version: version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert time to binary
			dt := utils.NewDateTimeFromTime(tc.time)
			binary, err := dt.Marshal()
			if err != nil {
				t.Fatalf("Failed to marshal time: %v", err)
			}

			// Convert binary back to time
			result := utils.ConvertFromBinaryTime(binary, tc.source, tc.version)

			// Check if the final time matches the original
			if !result.GetTime().Equal(dt.GetTime()) {
				t.Errorf("Time involution failed.\n | Original time : %v\n | Final time    : %v",
					dt.GetTime(), result.GetTime())
			}

			// Additional check for exact tick matching
			resultBinary, err := result.Marshal()
			if err != nil {
				t.Fatalf("Failed to marshal result time: %v", err)
			}

			if !bytes.Equal(binary, resultBinary) {
				t.Errorf("Binary representation mismatch after involution.\n | Original: %v\n | Final: %v",
					binary, resultBinary)
			}
		})
	}
}

// A timestamp entry of eight zero bytes is a well-formed FILETIME denoting
// 1601-01-01, so decoding it must not consult the clock. It used to come back as
// the moment of the call, because the decoder went through NewDateTimeFromTicks,
// whose zero argument is overloaded to mean "now".
func TestConvertFromBinaryTime_ZeroIsTheEpochNotNow(t *testing.T) {
	raw := make([]byte, 8)

	for _, kcv := range []version.KeyCredentialLinkVersion{
		{Value: version.KeyCredentialLinkVersion_0},
		{Value: version.KeyCredentialLinkVersion_1},
		{Value: version.KeyCredentialLinkVersion_2},
	} {
		decoded := utils.ConvertFromBinaryTime(raw, source.KeySource{Value: source.KeySource_AD}, kcv)

		expected := time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)
		if !decoded.GetTime().UTC().Equal(expected) {
			t.Errorf("version %d: decoded %s, want %s", kcv.Value, decoded.GetTime().UTC(), expected)
		}

		if decoded.GetTicks() != 0 {
			t.Errorf("version %d: ticks = %d, want 0", kcv.Value, decoded.GetTicks())
		}

		if delta := time.Since(decoded.GetTime()); delta < time.Minute {
			t.Errorf("version %d: decoded a value %s from now, which means the clock was consulted", kcv.Value, delta)
		}
	}
}

// The constructor keeps its documented overload: callers that build a credential
// pass zero to mean the current time, and that must stay separate from decoding.
func TestNewDateTimeFromTicks_ZeroStillMeansNow(t *testing.T) {
	before := time.Now()
	dt := utils.NewDateTimeFromTicks(0)
	after := time.Now()

	if dt.GetTime().Before(before) || dt.GetTime().After(after) {
		t.Errorf("NewDateTimeFromTicks(0) = %s, want an instant between %s and %s", dt.GetTime(), before, after)
	}
}
