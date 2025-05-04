package dialects_test

import (
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/dialects"
)

func TestNewDialects(t *testing.T) {
	d := dialects.NewDialects()
	if d == nil {
		t.Fatal("NewDialects returned nil")
	}
	if len(d.Dialects) != 0 {
		t.Errorf("Expected empty dialects slice, got %v", d.Dialects)
	}
}

func TestAddDialect(t *testing.T) {
	d := dialects.NewDialects()

	// Test adding a single dialect
	d.AddDialect("NT LM 0.12")
	if len(d.Dialects) != 1 {
		t.Errorf("Expected 1 dialect, got %d", len(d.Dialects))
	}
	if d.Dialects[0] != "NT LM 0.12" {
		t.Errorf("Expected 'NT LM 0.12', got '%s'", d.Dialects[0])
	}

	// Test adding multiple dialects
	d.AddDialect("LANMAN2.1")
	d.AddDialect("DOS LANMAN2.1")

	if len(d.Dialects) != 3 {
		t.Errorf("Expected 3 dialects, got %d", len(d.Dialects))
	}

	expectedDialects := []string{"NT LM 0.12", "LANMAN2.1", "DOS LANMAN2.1"}
	for i, dialect := range expectedDialects {
		if d.Dialects[i] != dialect {
			t.Errorf("Expected dialect[%d] to be '%s', got '%s'", i, dialect, d.Dialects[i])
		}
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	testCases := []struct {
		name     string
		dialects []string
	}{
		{
			name:     "Empty dialects",
			dialects: []string{},
		},
		{
			name:     "Single dialect",
			dialects: []string{"NT LM 0.12"},
		},
		{
			name:     "Multiple dialects",
			dialects: []string{"NT LM 0.12", "LANMAN2.1", "DOS LANMAN2.1"},
		},
		{
			name:     "Dialects with special characters",
			dialects: []string{"NT LM 0.12", "PC NETWORK PROGRAM 1.0", "MICROSOFT NETWORKS 1.03"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create and populate dialects
			d := dialects.NewDialects()
			for _, dialect := range tc.dialects {
				d.AddDialect(dialect)
			}

			// Marshal
			marshalledData, err := d.Marshal()
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			// Unmarshal into a new Dialects struct
			d2 := dialects.NewDialects()
			bytesRead, err := d2.Unmarshal(marshalledData)
			if err != nil {
				t.Fatalf("Unmarshal failed with data %v: %v", marshalledData, err)
			}

			// Verify bytes read
			if bytesRead != len(marshalledData) {
				t.Errorf("Expected to read %d bytes, got %d", len(marshalledData), bytesRead)
			}

			// Verify unmarshalled data
			if len(tc.dialects) == 0 {
				// Special case for empty dialects
				if len(d2.Dialects) != 1 || d2.Dialects[0] != "" {
					t.Errorf("Expected empty string in dialects, got %v", d2.Dialects)
				}
			} else {
				// Check that all dialects are present
				joinedOriginal := strings.Join(tc.dialects, "\x00")
				joinedUnmarshalled := strings.Join(d2.Dialects, "\x00")

				if joinedOriginal != joinedUnmarshalled {
					t.Errorf("Unmarshalled dialects don't match original.\nExpected: %v\nGot: %v",
						tc.dialects, d2.Dialects)
				}
			}
		})
	}
}

func TestUnmarshalWithInvalidData(t *testing.T) {
	d := dialects.NewDialects()

	// Test with data that's too short
	_, err := d.Unmarshal([]byte{0x01})
	if err == nil {
		t.Error("Expected error when unmarshalling data that's too short, got nil")
	}

	// Test with invalid buffer format
	_, err = d.Unmarshal([]byte{0x03, 0x00})
	if err == nil {
		t.Error("Expected error when unmarshalling data with invalid buffer format, got nil")
	}
}
