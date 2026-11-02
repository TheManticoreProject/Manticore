package utils_test

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"strings"
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
			// Version 0 stores .NET DateTime.Ticks (since 0001-01-01).
			name:     "Version 0 AD source",
			input:    []byte{0x80, 0xa3, 0x99, 0x56, 0x7b, 0x06, 0xda, 0x08},
			source:   source.KeySource{Value: source.KeySource_AD},
			version:  version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_0},
			expected: testTimeStruct,
		},
		{
			// Version 1 stores DateTime.ToBinary; this one has DateTimeKind.Utc.
			name:     "Version 1 AD source",
			input:    []byte{0x80, 0xa3, 0x99, 0x56, 0x7b, 0x06, 0xda, 0x48},
			source:   source.KeySource{Value: source.KeySource_AD},
			version:  version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_1},
			expected: testTimeStruct,
		},
		{
			// DateTimeKind.Local: ToBinary stores the UTC ticks under the top bit.
			name:     "Version 1 local kind",
			input:    []byte{0x80, 0xa3, 0x99, 0x56, 0x7b, 0x06, 0xda, 0x88},
			source:   source.KeySource{Value: source.KeySource_AD},
			version:  version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_1},
			expected: testTimeStruct,
		},
		{
			// DateTimeKind.Unspecified carries no zone and is read as UTC.
			name:     "Version 1 unspecified kind",
			input:    []byte{0x80, 0xa3, 0x99, 0x56, 0x7b, 0x06, 0xda, 0x08},
			source:   source.KeySource{Value: source.KeySource_AD},
			version:  version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_1},
			expected: testTimeStruct,
		},
		{
			// Azure AD keys of version 2 store DateTime.ToBinary, not a FILETIME.
			name:     "Version 2 Azure AD source",
			input:    []byte{0x80, 0xa3, 0x99, 0x56, 0x7b, 0x06, 0xda, 0x48},
			source:   source.KeySource{Value: source.KeySource_AzureAD},
			version:  version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2},
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
			result, err := utils.ConvertFromBinaryTime(tc.input, tc.source, tc.version)
			if err != nil {
				t.Fatalf("ConvertFromBinaryTime() error = %v, want nil", err)
			}
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

	// The same instant as .NET DateTime.Ticks and as DateTime.ToBinary with
	// DateTimeKind.Utc.
	dotNetTicksBytes := []byte{0x80, 0xa3, 0x99, 0x56, 0x7b, 0x06, 0xda, 0x08}
	dotNetBinaryBytes := []byte{0x80, 0xa3, 0x99, 0x56, 0x7b, 0x06, 0xda, 0x48}

	testCases := []struct {
		name    string
		source  source.KeySource
		version version.KeyCredentialLinkVersion
		want    []byte
	}{
		{
			name:    "Version 0 AD source",
			source:  source.KeySource{Value: source.KeySource_AD},
			version: version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_0},
			want:    dotNetTicksBytes,
		},
		{
			name:    "Version 1 AD source",
			source:  source.KeySource{Value: source.KeySource_AD},
			version: version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_1},
			want:    dotNetBinaryBytes,
		},
		{
			name:    "Version 2 AD source",
			source:  source.KeySource{Value: source.KeySource_AD},
			version: version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2},
			want:    testTimeBytes,
		},
		{
			name:    "Version 2 Azure AD source",
			source:  source.KeySource{Value: source.KeySource_AzureAD},
			version: version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2},
			want:    dotNetBinaryBytes,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded := utils.ConvertToBinaryTime(testTimeStruct, tc.source, tc.version)

			if !bytes.Equal(encoded, tc.want) {
				t.Errorf("ConvertToBinaryTime() = %s, want %s",
					hex.EncodeToString(encoded), hex.EncodeToString(tc.want))
			}

			// The two functions are named as inverses and have to behave as such.
			decoded, err := utils.ConvertFromBinaryTime(encoded, tc.source, tc.version)
			if err != nil {
				t.Fatalf("ConvertFromBinaryTime() error = %v, want nil", err)
			}
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
			// Convert time to binary in the encoding of this version and source
			dt := utils.NewDateTimeFromTime(tc.time)
			binary := utils.ConvertToBinaryTime(tc.time, tc.source, tc.version)

			// Convert binary back to time
			result, err := utils.ConvertFromBinaryTime(binary, tc.source, tc.version)
			if err != nil {
				t.Fatalf("Failed to decode time: %v", err)
			}

			// Check if the final time matches the original
			if !result.GetTime().Equal(dt.GetTime()) {
				t.Errorf("Time involution failed.\n | Original time : %v\n | Final time    : %v",
					dt.GetTime(), result.GetTime())
			}

			// Additional check for exact byte matching
			resultBinary := utils.ConvertToBinaryTime(result.GetTime(), tc.source, tc.version)

			if !bytes.Equal(binary, resultBinary) {
				t.Errorf("Binary representation mismatch after involution.\n | Original: %v\n | Final: %v",
					binary, resultBinary)
			}
		})
	}
}

// A timestamp entry of eight zero bytes is a well-formed value denoting its
// encoding's epoch (1601-01-01 for a FILETIME, 0001-01-01 for .NET ticks), so
// decoding it must not consult the clock. It used to come back as the moment of
// the call, because the decoder went through NewDateTimeFromTicks, whose zero
// argument is overloaded to mean "now".
func TestConvertFromBinaryTime_ZeroIsTheEpochNotNow(t *testing.T) {
	raw := make([]byte, 8)

	for _, tc := range []struct {
		kcv      version.KeyCredentialLinkVersion
		expected time.Time
	}{
		{version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_0}, time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)},
		{version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_1}, time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)},
		{version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2}, time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)},
	} {
		kcv, expected := tc.kcv, tc.expected
		decoded, err := utils.ConvertFromBinaryTime(raw, source.KeySource{Value: source.KeySource_AD}, kcv)
		if err != nil {
			t.Fatalf("ConvertFromBinaryTime() error = %v, want nil", err)
		}

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

// The value of a timestamp entry is as long as the entry declared, so a blob can
// parse into one too short to hold the field. Reading it used to index past the end
// of the slice and panic.
func TestConvertFromBinaryTime_ShortBuffer(t *testing.T) {
	adSrc := source.KeySource{Value: source.KeySource_AD}
	kcv := version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2}

	for _, test := range []struct {
		name string
		data []byte
	}{
		{name: "nil", data: nil},
		{name: "empty", data: []byte{}},
		{name: "one byte", data: []byte{0x01}},
		{name: "seven bytes", data: []byte{1, 2, 3, 4, 5, 6, 7}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := utils.ConvertFromBinaryTime(test.data, adSrc, kcv); err == nil {
				t.Errorf("ConvertFromBinaryTime(%d bytes) returned no error, want one", len(test.data))
			}
		})
	}

	t.Run("eight bytes still decodes", func(t *testing.T) {
		data := []byte{0x80, 0xa3, 0x22, 0x34, 0x64, 0x38, 0xd8, 0x01}

		decoded, err := utils.ConvertFromBinaryTime(data, adSrc, kcv)
		if err != nil {
			t.Fatalf("ConvertFromBinaryTime() error = %v, want nil", err)
		}

		expected := time.Date(2022, 3, 15, 12, 0, 3, 0, time.UTC)
		if !decoded.GetTime().Equal(expected) {
			t.Errorf("decoded %s, want %s", decoded.GetTime(), expected)
		}
	})
}

// Base64 padding is one or two characters depending on the encoded length, so an
// identifier of any length has to survive the round trip. Rewriting the padding with
// a fixed single '=' only ever worked for lengths that are 2 mod 3 — which is what a
// 32-byte SHA-256 hash happens to be, hiding the defect on the normal path.
func TestConvertToBinaryIdentifier_RoundTripsEveryLength(t *testing.T) {
	for _, kcv := range []version.KeyCredentialLinkVersion{
		{Value: version.KeyCredentialLinkVersion_0},
		{Value: version.KeyCredentialLinkVersion_1},
		{Value: version.KeyCredentialLinkVersion_2},
	} {
		for size := 1; size <= 40; size++ {
			raw := make([]byte, size)
			for i := range raw {
				raw[i] = byte(i)
			}

			encoded := utils.ConvertFromBinaryIdentifier(raw, kcv)

			decoded, err := utils.ConvertToBinaryIdentifier(encoded, kcv)
			if err != nil {
				t.Errorf("version %d, %d bytes (%d mod 3): ConvertToBinaryIdentifier(%q) error = %v",
					kcv.Value, size, size%3, encoded, err)
				continue
			}

			if !bytes.Equal(decoded, raw) {
				t.Errorf("version %d, %d bytes: round trip gave %d bytes, want the original %d",
					kcv.Value, size, len(decoded), size)
			}
		}
	}
}

// An identifier whose padding has been stripped still decodes, which is what the
// original padding rewrite was reaching for.
func TestConvertToBinaryIdentifier_AcceptsUnpaddedInput(t *testing.T) {
	kcv := version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2}
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = byte(i)
	}

	unpadded := strings.TrimRight(utils.ConvertFromBinaryIdentifier(raw, kcv), "=")

	decoded, err := utils.ConvertToBinaryIdentifier(unpadded, kcv)
	if err != nil {
		t.Fatalf("ConvertToBinaryIdentifier(%q) error = %v, want nil", unpadded, err)
	}

	if !bytes.Equal(decoded, raw) {
		t.Errorf("decoded %d bytes, want the original 32", len(decoded))
	}
}

// TestEncodeBinaryTimeKeepsDecodedBytes checks a decoded timestamp is re-emitted
// byte for byte (here a DateTimeKind.Local value, which re-encoding from the
// instant would turn into Utc), and that changing it re-encodes for the version
// and source.
func TestEncodeBinaryTimeKeepsDecodedBytes(t *testing.T) {
	azure := source.KeySource{Value: source.KeySource_AzureAD}
	v2 := version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2}
	local := []byte{0x80, 0xa3, 0x99, 0x56, 0x7b, 0x06, 0xda, 0x88}

	dt, err := utils.ConvertFromBinaryTime(local, azure, v2)
	if err != nil {
		t.Fatal(err)
	}
	if got := utils.EncodeBinaryTime(dt, azure, v2); !bytes.Equal(got, local) {
		t.Errorf("EncodeBinaryTime = %x, want the decoded bytes %x", got, local)
	}

	dt.SetTime(dt.GetTime())
	if got, want := utils.EncodeBinaryTime(dt, azure, v2), []byte{0x80, 0xa3, 0x99, 0x56, 0x7b, 0x06, 0xda, 0x48}; !bytes.Equal(got, want) {
		t.Errorf("after SetTime: EncodeBinaryTime = %x, want the Utc encoding %x", got, want)
	}
}

// TestConvertToBinaryTimeClampsDotNetRange checks a time outside what a .NET
// DateTime holds is clamped instead of wrapping into another date or the kind
// bits.
func TestConvertToBinaryTimeClampsDotNetRange(t *testing.T) {
	v0 := version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_0}
	ad := source.KeySource{Value: source.KeySource_AD}
	for _, tc := range []struct {
		in   time.Time
		want uint64
	}{
		{time.Date(0, 6, 1, 0, 0, 0, 0, time.UTC), 0},
		{time.Date(15000, 1, 1, 0, 0, 0, 0, time.UTC), 3155378975999999999},
	} {
		got := binary.LittleEndian.Uint64(utils.ConvertToBinaryTime(tc.in, ad, v0))
		if got != tc.want {
			t.Errorf("%s: ticks %d, want %d", tc.in, got, tc.want)
		}
	}
}

// TestDateTimeStaysComparable pins that DateTime can still be compared with ==.
func TestDateTimeStaysComparable(t *testing.T) {
	a := utils.NewDateTimeFromTime(time.Date(2022, 3, 15, 12, 0, 3, 0, time.UTC))
	b := utils.NewDateTimeFromTime(time.Date(2022, 3, 15, 12, 0, 3, 0, time.UTC))
	if a != b {
		t.Error("equal DateTimes compare unequal")
	}
}
