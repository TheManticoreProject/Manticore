package types_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

func TestOEM_STRING_NewOEM_STRING(t *testing.T) {
	oemString := types.NewOEM_STRING()

	if oemString.BufferFormat != types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING {
		t.Errorf("Expected BufferFormat to be %d, got %d", types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING, oemString.BufferFormat)
	}

	if oemString.Length != 0 {
		t.Errorf("Expected Length to be 0, got %d", oemString.Length)
	}

	if len(oemString.Buffer) != 0 {
		t.Errorf("Expected Buffer to be empty, got %v", oemString.Buffer)
	}
}

func TestOEM_STRING_Marshal(t *testing.T) {
	oemString := types.NewOEM_STRING()
	oemString.SetString("test")

	data, err := oemString.Marshal()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if data[0] != byte(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING) {
		t.Errorf("Expected first byte to be %d, got %d", types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING, data[0])
	}

	expected := []byte{'t', 'e', 's', 't', 0}
	if !bytes.Equal(data[1:], expected) {
		t.Errorf("Expected data[1:] to be %v, got %v", expected, data[1:])
	}
}

func TestOEM_STRING_Unmarshal(t *testing.T) {
	oemString := types.NewOEM_STRING()
	data := []byte{byte(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING), 't', 'e', 's', 't', 0}

	bytesRead, err := oemString.Unmarshal(data)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if bytesRead != 6 {
		t.Errorf("Expected bytesRead to be 6, got %d", bytesRead)
	}

	if oemString.BufferFormat != types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING {
		t.Errorf("Expected BufferFormat to be %d, got %d", types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING, oemString.BufferFormat)
	}

	expected := []byte{'t', 'e', 's', 't'}
	if !bytes.Equal(oemString.Buffer, expected) {
		t.Errorf("Expected Buffer to be %v, got %v", expected, oemString.Buffer)
	}
}

func TestOEM_STRING_GetString(t *testing.T) {
	oemString := types.NewOEM_STRING()
	oemString.SetString("test")

	result := oemString.GetString()

	if result != "test" {
		t.Errorf("Expected result to be 'test', got '%s'", result)
	}
}

func TestOEM_STRING_SetString(t *testing.T) {
	oemString := types.NewOEM_STRING()

	oemString.SetString("hello")

	if oemString.Length != 5 {
		t.Errorf("Expected Length to be 5, got %d", oemString.Length)
	}

	expected := []byte("hello")
	if !bytes.Equal(oemString.Buffer, expected) {
		t.Errorf("Expected Buffer to be %v, got %v", expected, oemString.Buffer)
	}
}
