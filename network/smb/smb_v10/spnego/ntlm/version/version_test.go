package version_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/spnego/ntlm/version"
)

func TestDefaultVersion(t *testing.T) {
	v := version.DefaultVersion()

	if v.ProductMajorVersion != 10 {
		t.Errorf("Expected ProductMajorVersion to be 10, got %d", v.ProductMajorVersion)
	}

	if v.ProductMinorVersion != 0 {
		t.Errorf("Expected ProductMinorVersion to be 0, got %d", v.ProductMinorVersion)
	}

	if v.ProductBuild != 18362 {
		t.Errorf("Expected ProductBuild to be 18362, got %d", v.ProductBuild)
	}

	if v.NTLMRevision != version.NTLMSSP_REVISION_W2K3 {
		t.Errorf("Expected NTLMRevision to be %d, got %d", version.NTLMSSP_REVISION_W2K3, v.NTLMRevision)
	}
}

func TestNewVersion(t *testing.T) {
	v := version.NewVersion(6, 1, 7601, version.NTLMSSP_REVISION_W2K3)

	if v.ProductMajorVersion != 6 {
		t.Errorf("Expected ProductMajorVersion to be 6, got %d", v.ProductMajorVersion)
	}

	if v.ProductMinorVersion != 1 {
		t.Errorf("Expected ProductMinorVersion to be 1, got %d", v.ProductMinorVersion)
	}

	if v.ProductBuild != 7601 {
		t.Errorf("Expected ProductBuild to be 7601, got %d", v.ProductBuild)
	}

	if v.NTLMRevision != version.NTLMSSP_REVISION_W2K3 {
		t.Errorf("Expected NTLMRevision to be %d, got %d", version.NTLMSSP_REVISION_W2K3, v.NTLMRevision)
	}

	// Check that Reserved is initialized to zeros
	for i, b := range v.Reserved {
		if b != 0 {
			t.Errorf("Expected Reserved[%d] to be 0, got %d", i, b)
		}
	}
}

func TestVersionString(t *testing.T) {
	v := version.NewVersion(10, 0, 18362, version.NTLMSSP_REVISION_W2K3)
	expected := "Version 10.0 (Build 18362) NTLM Revision 15"

	if v.String() != expected {
		t.Errorf("Expected String() to return %q, got %q", expected, v.String())
	}
}

func TestVersionMarshalUnmarshal(t *testing.T) {
	testCases := []struct {
		name    string
		version version.Version
	}{
		{
			name:    "Default Windows 10",
			version: version.DefaultVersion(),
		},
		{
			name:    "Windows 7",
			version: version.NewVersion(6, 1, 7601, version.NTLMSSP_REVISION_W2K3),
		},
		{
			name:    "Windows XP",
			version: version.NewVersion(5, 1, 2600, version.NTLMSSP_REVISION_W2K3),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Marshal
			data, err := tc.version.Marshal()
			if err != nil {
				t.Fatalf("Marshal() returned error: %v", err)
			}

			// Verify length
			if len(data) != 8 {
				t.Errorf("Expected marshaled data length to be 8, got %d", len(data))
			}

			// Unmarshal
			var unmarshaled version.Version
			bytesRead, err := unmarshaled.Unmarshal(data)
			if err != nil {
				t.Fatalf("UnmarshalVersion() returned error: %v", err)
			}

			// Verify bytes read
			if bytesRead != 8 {
				t.Errorf("Expected UnmarshalVersion to read 8 bytes, got %d", bytesRead)
			}

			// Verify fields match
			if unmarshaled.ProductMajorVersion != tc.version.ProductMajorVersion {
				t.Errorf("ProductMajorVersion mismatch: expected %d, got %d",
					tc.version.ProductMajorVersion, unmarshaled.ProductMajorVersion)
			}
			if unmarshaled.ProductMinorVersion != tc.version.ProductMinorVersion {
				t.Errorf("ProductMinorVersion mismatch: expected %d, got %d",
					tc.version.ProductMinorVersion, unmarshaled.ProductMinorVersion)
			}
			if unmarshaled.ProductBuild != tc.version.ProductBuild {
				t.Errorf("ProductBuild mismatch: expected %d, got %d",
					tc.version.ProductBuild, unmarshaled.ProductBuild)
			}
			if !bytes.Equal(unmarshaled.Reserved[:], tc.version.Reserved[:]) {
				t.Errorf("Reserved mismatch: expected %v, got %v",
					tc.version.Reserved, unmarshaled.Reserved)
			}
			if unmarshaled.NTLMRevision != tc.version.NTLMRevision {
				t.Errorf("NTLMRevision mismatch: expected %d, got %d",
					tc.version.NTLMRevision, unmarshaled.NTLMRevision)
			}
		})
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	testCases := []struct {
		name    string
		version version.Version
	}{
		{
			name: "Windows 10",
			version: version.Version{
				ProductMajorVersion: 10,
				ProductMinorVersion: 0,
				ProductBuild:        18362,
				Reserved:            [3]byte{0, 0, 0},
				NTLMRevision:        version.NTLMSSP_REVISION_W2K3,
			},
		},
		{
			name: "Windows 7",
			version: version.Version{
				ProductMajorVersion: 6,
				ProductMinorVersion: 1,
				ProductBuild:        7601,
				Reserved:            [3]byte{0, 0, 0},
				NTLMRevision:        version.NTLMSSP_REVISION_W2K3,
			},
		},
		{
			name: "Windows Server 2019",
			version: version.Version{
				ProductMajorVersion: 10,
				ProductMinorVersion: 0,
				ProductBuild:        17763,
				Reserved:            [3]byte{0, 0, 0},
				NTLMRevision:        version.NTLMSSP_REVISION_W2K3,
			},
		},
		{
			name: "Custom Version",
			version: version.Version{
				ProductMajorVersion: 12,
				ProductMinorVersion: 5,
				ProductBuild:        9999,
				Reserved:            [3]byte{1, 2, 3}, // Non-standard reserved values
				NTLMRevision:        20,               // Non-standard revision
			},
		},
		{
			name:    "Default Version",
			version: version.DefaultVersion(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Marshal the version
			data, err := tc.version.Marshal()
			if err != nil {
				t.Fatalf("Failed to marshal version: %v", err)
			}

			// Unmarshal into a new version object
			var unmarshaled version.Version
			bytesRead, err := unmarshaled.Unmarshal(data)
			if err != nil {
				t.Fatalf("Failed to unmarshal version: %v", err)
			}

			// Verify correct number of bytes read
			if bytesRead != 8 {
				t.Errorf("Expected to read 8 bytes, got %d", bytesRead)
			}

			// Compare the original and unmarshaled versions
			if unmarshaled.ProductMajorVersion != tc.version.ProductMajorVersion {
				t.Errorf("ProductMajorVersion mismatch: expected %d, got %d",
					tc.version.ProductMajorVersion, unmarshaled.ProductMajorVersion)
			}
			if unmarshaled.ProductMinorVersion != tc.version.ProductMinorVersion {
				t.Errorf("ProductMinorVersion mismatch: expected %d, got %d",
					tc.version.ProductMinorVersion, unmarshaled.ProductMinorVersion)
			}
			if unmarshaled.ProductBuild != tc.version.ProductBuild {
				t.Errorf("ProductBuild mismatch: expected %d, got %d",
					tc.version.ProductBuild, unmarshaled.ProductBuild)
			}
			if !bytes.Equal(unmarshaled.Reserved[:], tc.version.Reserved[:]) {
				t.Errorf("Reserved mismatch: expected %v, got %v",
					tc.version.Reserved, unmarshaled.Reserved)
			}
			if unmarshaled.NTLMRevision != tc.version.NTLMRevision {
				t.Errorf("NTLMRevision mismatch: expected %d, got %d",
					tc.version.NTLMRevision, unmarshaled.NTLMRevision)
			}

			// Marshal the unmarshaled version and compare with original marshaled data
			remarshaled, err := unmarshaled.Marshal()
			if err != nil {
				t.Fatalf("Failed to marshal unmarshaled version: %v", err)
			}
			if !bytes.Equal(data, remarshaled) {
				t.Errorf("Remarshaled data doesn't match original: expected %v, got %v", data, remarshaled)
			}
		})
	}
}

func TestUnmarshalVersionWithShortData(t *testing.T) {
	var v version.Version
	data := []byte{10, 0, 0, 0} // Only 4 bytes, should be 8

	_, err := v.Unmarshal(data)
	if err == nil {
		t.Error("Expected UnmarshalVersion to return error with short data, got nil")
	}
}
