package data_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
)

func Test_DataMarshalUnmarshal(t *testing.T) {
	d := data.NewData()
	d.ByteCount = 5
	d.Bytes = []byte{0x01, 0x02, 0x03, 0x04, 0x05}

	marshalled, err := d.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal SMB data: %v", err)
	}

	if marshalled[0] != 5 {
		t.Errorf("ByteCount mismatch in marshalled data: %d != 5", marshalled[0])
	}

	if len(marshalled) != 2+len(d.Bytes) {
		t.Errorf("Unexpected marshalled length: %d != %d", len(marshalled), 2+len(d.Bytes))
	}

	// Unmarshal and verify
	unmarshalledData := data.NewData()
	_, err = unmarshalledData.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Failed to unmarshal SMB data: %v", err)
	}

	if unmarshalledData.ByteCount != d.ByteCount {
		t.Errorf("ByteCount mismatch: %d != %d", unmarshalledData.ByteCount, d.ByteCount)
	}

	if len(unmarshalledData.Bytes) != len(d.Bytes) {
		t.Errorf("Bytes length mismatch: %d != %d", len(unmarshalledData.Bytes), len(d.Bytes))
	}

	for i, b := range d.Bytes {
		if unmarshalledData.Bytes[i] != b {
			t.Errorf("Byte %d mismatch: %x != %x", i, unmarshalledData.Bytes[i], b)
		}
	}
}

func Test_DataEmptyMarshalUnmarshal(t *testing.T) {
	d := data.NewData()
	d.ByteCount = 0
	d.Bytes = []byte{}

	marshalled, err := d.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal empty SMB data: %v", err)
	}

	if len(marshalled) != 2 {
		t.Errorf("Unexpected marshalled length for empty data: %d != 2", len(marshalled))
	}

	if marshalled[0] != 0 {
		t.Errorf("ByteCount mismatch in marshalled data: %d != 0", marshalled[0])
	}

	// Unmarshal and verify
	unmarshalledData := data.NewData()
	_, err = unmarshalledData.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Failed to unmarshal empty SMB data: %v", err)
	}

	if unmarshalledData.ByteCount != 0 {
		t.Errorf("ByteCount mismatch: %d != 0", unmarshalledData.ByteCount)
	}

	if len(unmarshalledData.Bytes) != 0 {
		t.Errorf("Bytes should be empty, got length: %d", len(unmarshalledData.Bytes))
	}
}

func Test_DataUnmarshalWithInsufficientData(t *testing.T) {
	// Create a buffer with ByteCount = 5 but only enough data for 3 bytes
	buf := bytes.NewBuffer([]byte{})
	buf.Write([]byte{0x05, 0x00})       // ByteCount = 5
	buf.Write([]byte{0x01, 0x02, 0x03}) // Only 3 bytes of data

	data := data.NewData()
	_, err := data.Unmarshal(buf.Bytes())

	if err == nil {
		t.Errorf("Expected error when unmarshalling with insufficient data, got nil")
	}
}

func Test_DataUnmarshalEmptyData(t *testing.T) {
	data := data.NewData()
	_, err := data.Unmarshal([]byte{})
	if err == nil {
		t.Errorf("Expected error when unmarshalling empty data, got nil")
	}

	if data.ByteCount != 0 {
		t.Errorf("Expected ByteCount to be 0, got: %d", data.ByteCount)
	}

	if len(data.Bytes) != 0 {
		t.Errorf("Expected Bytes to be empty, got length: %d", len(data.Bytes))
	}
}
