package data_structures_test

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_structures"
)

func TestFILETIME_ToInt64(t *testing.T) {
	testCases := []struct {
		name              string
		hexData           string
		expectedTimestamp int64
	}{
		{
			name:              "Aug 27, 2011 23:21:49.781250000 CEST",
			hexData:           "148a5255ff64cc01",
			expectedTimestamp: int64(0x01cc64ff55528a14),
		},
		{
			name:              "Jan 1, 2000 00:00:00 UTC",
			hexData:           "0000e1f505e0cd01",
			expectedTimestamp: int64(0x01cde005f5e10000),
		},
		{
			name:              "Dec 31, 2020 23:59:59 UTC",
			hexData:           "00a0bf5ed095e501",
			expectedTimestamp: int64(0x01e595d05ebfa000),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			smbTime := &data_structures.FILETIME{}

			data, err := hex.DecodeString(tc.hexData)
			if err != nil {
				t.Fatalf("Failed to decode hex string: %v", err)
			}
			smbTime.Unmarshal(data)

			if timestamp := smbTime.ToInt64(); timestamp != tc.expectedTimestamp {
				t.Errorf("Expected timestamp 0x%016x, got 0x%016x", tc.expectedTimestamp, timestamp)
			}
		})
	}
}

func TestFILETIME_GetUnixTimestamp(t *testing.T) {
	// Test case: 148a5255ff64cc01 should be Aug 27, 2011 23:21:49.781250000 CEST
	smbTime := &data_structures.FILETIME{}

	hexData := "148a5255ff64cc01"
	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatalf("Failed to decode hex string: %v", err)
	}
	smbTime.Unmarshal(data)

	// Expected timestamp for Aug 27, 2011 23:21:49 UTC
	expectedTime := time.Date(2011, time.August, 27, 21, 21, 49, 781250000, time.UTC)
	expectedTimestamp := expectedTime.Unix()

	if timestamp := smbTime.GetUnixTimestamp(); timestamp != expectedTimestamp {
		t.Errorf("Expected timestamp %d, got %d", expectedTimestamp, timestamp)
		t.Errorf("Expected timestamp 0x%08x, got 0x%08x", expectedTimestamp, timestamp)
	}
}

func TestFILETIME_GetTime(t *testing.T) {
	// Test case: 148a5255ff64cc01 should be Aug 27, 2011 23:21:49.781250000 CEST
	smbTime := &data_structures.FILETIME{}

	hexData := "148a5255ff64cc01"
	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatalf("Failed to decode hex string: %v", err)
	}
	smbTime.Unmarshal(data)

	// Expected time for Aug 27, 2011 23:21:49 UTC
	expectedTime := time.Date(2011, time.August, 27, 21, 21, 49, 781250000, time.UTC)

	if gotTime := smbTime.GetTime(); !gotTime.Equal(expectedTime) {
		t.Errorf("Expected time %s, got %s",
			expectedTime.Format(time.RFC3339),
			gotTime.Format(time.RFC3339))
	}
}

func TestFILETIME_String(t *testing.T) {
	// Test case: 148a5255ff64cc01 should be Aug 27, 2011 23:21:49.781250000 CEST
	smbTime := &data_structures.FILETIME{}

	hexData := "148a5255ff64cc01"
	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatalf("Failed to decode hex string: %v", err)
	}
	smbTime.Unmarshal(data)

	// Expected string format for Aug 27, 2011 21:21:49 UTC
	expectedString := "2011-08-27 21:21:49.78125"

	if str := smbTime.String(); str != expectedString {
		t.Errorf("Expected string %s, got %s", expectedString, str)
	}
}

func TestFILETIME_Unmarshal(t *testing.T) {
	smbTime := &data_structures.FILETIME{}

	hexData := "148a5255ff64cc01"
	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatalf("Failed to decode hex string: %v", err)
	}
	smbTime.Unmarshal(data)

	expectedHighPart := uint32(0x01cc64ff)
	expectedLowPart := uint32(0x55528a14)

	if smbTime.DwHighDateTime != expectedHighPart || smbTime.DwLowDateTime != expectedLowPart {
		t.Errorf("Expected HighPart = 0x%08x got HighPart = 0x%08x", expectedHighPart, smbTime.DwHighDateTime)
		t.Errorf("Expected LowPart  = 0x%08x got LowPart  = 0x%08x", expectedLowPart, smbTime.DwLowDateTime)
	}

	// Test with insufficient data
	shortData := make([]byte, 4)
	originalHighPart := smbTime.DwHighDateTime
	originalLowPart := smbTime.DwLowDateTime

	smbTime.Unmarshal(shortData)

	if smbTime.DwHighDateTime != originalHighPart || smbTime.DwLowDateTime != originalLowPart {
		t.Errorf("Expected no change with insufficient data, but values were modified")
	}
}

func TestNewFILETIMEFromTime(t *testing.T) {
	testCases := []struct {
		name          string
		inputTime     time.Time
		expectedInt64 int64
		expectedLow   uint32
		expectedHigh  uint32
	}{
		{
			name:          "Aug 27, 2011 23:21:49.781250000 UTC",
			inputTime:     time.Date(2011, time.August, 27, 23, 21, 49, 781250000, time.UTC),
			expectedInt64: int64(0x01cc651018db5a14),
			expectedLow:   uint32(0x18db5a14),
			expectedHigh:  uint32(0x01cc6510),
		},
		{
			name:          "Jan 1, 2000 00:00:00 UTC",
			inputTime:     time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
			expectedInt64: int64(0x01bf53eb256d4000),
			expectedLow:   uint32(0x256d4000),
			expectedHigh:  uint32(0x01bf53eb),
		},
		{
			name:          "Dec 31, 2020 23:59:59 UTC",
			inputTime:     time.Date(2020, time.December, 31, 23, 59, 59, 0, time.UTC),
			expectedInt64: int64(0x01d6dfd10b9ce980),
			expectedLow:   uint32(0x0b9ce980),
			expectedHigh:  uint32(0x01d6dfd1),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fileTime := data_structures.NewFILETIMEFromTime(tc.inputTime)

			// Check that the int64 representation matches
			if got := fileTime.ToInt64(); got != tc.expectedInt64 {
				t.Errorf("ToInt64() = 0x%016x, want 0x%016x", got, tc.expectedInt64)
			}

			// Check that the high and low parts match
			if fileTime.DwLowDateTime != tc.expectedLow {
				t.Errorf("DwLowDateTime = 0x%08x, want 0x%08x", fileTime.DwLowDateTime, tc.expectedLow)
			}

			if fileTime.DwHighDateTime != tc.expectedHigh {
				t.Errorf("DwHighDateTime = 0x%08x, want 0x%08x", fileTime.DwHighDateTime, tc.expectedHigh)
			}

			// Check that converting back to time.Time works correctly
			if gotTime := fileTime.GetTime(); !gotTime.Equal(tc.inputTime) {
				t.Errorf("GetTime() = %v, want %v", gotTime, tc.inputTime)
			}
		})
	}
}

func TestFILETIME_MarshalUnmarshalInvolution(t *testing.T) {
	// Create an FILETIME instance with known values
	originalTime := &data_structures.FILETIME{
		DwLowDateTime:  0x55528a14,
		DwHighDateTime: 0x01cc64ff,
	}

	// Marshal the FILETIME structure to bytes
	marshalledData, err := originalTime.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal FILETIME: %v", err)
	}

	// Create a new FILETIME instance to unmarshal into
	unmarshalledTime := &data_structures.FILETIME{}

	// Unmarshal the bytes back into the new instance
	unmarshalledTime.Unmarshal(marshalledData)

	// Verify that the original and unmarshalled instances have the same values
	if originalTime.DwLowDateTime != unmarshalledTime.DwLowDateTime {
		t.Errorf("LowPart values don't match: original=0x%x, unmarshalled=0x%x",
			originalTime.DwLowDateTime, unmarshalledTime.DwLowDateTime)
	}

	if originalTime.DwHighDateTime != unmarshalledTime.DwHighDateTime {
		t.Errorf("HighPart values don't match: original=0x%x, unmarshalled=0x%x",
			originalTime.DwHighDateTime, unmarshalledTime.DwHighDateTime)
	}

	// Also verify that the int64 representation is preserved
	if originalTime.ToInt64() != unmarshalledTime.ToInt64() {
		t.Errorf("Int64 representations don't match: original=%d, unmarshalled=%d",
			originalTime.ToInt64(), unmarshalledTime.ToInt64())
	}

	// And verify that the time representation is preserved
	if originalTime.GetTime().Unix() != unmarshalledTime.GetTime().Unix() {
		t.Errorf("Time representations don't match: original=%s, unmarshalled=%s",
			originalTime.GetTime().String(), unmarshalledTime.GetTime().String())
	}
}
