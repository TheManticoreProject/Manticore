package parameters_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
)

func Test_ParametersMarshalUnmarshalInvolution(t *testing.T) {
	params := parameters.NewParameters()
	params.WordCount = 3
	params.Words = []uint16{0x1234, 0x5678, 0x9ABC}

	marshalled, err := params.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal SMB parameters: %v", err)
	}

	t.Logf("Marshalled SMB parameters: %x", marshalled)

	// Verify the structure of marshalled data
	if marshalled[0] != 3 {
		t.Errorf("WordCount mismatch in marshalled data: %d != 3", marshalled[0])
	}

	if len(marshalled) != 1+3*2 {
		t.Errorf("Unexpected marshalled length: %d != 7", len(marshalled))
	}

	// Unmarshal and verify
	unmarshalledParams := parameters.NewParameters()
	_, err = unmarshalledParams.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Failed to unmarshal SMB parameters: %v", err)
	}

	if unmarshalledParams.WordCount != params.WordCount {
		t.Errorf("WordCount mismatch: %d != %d", unmarshalledParams.WordCount, params.WordCount)
	}

	if len(unmarshalledParams.Words) != len(params.Words) {
		t.Errorf("Words length mismatch: %d != %d", len(unmarshalledParams.Words), len(params.Words))
	}

	for i, word := range params.Words {
		if unmarshalledParams.Words[i] != word {
			t.Errorf("Word %d mismatch: %x != %x", i, unmarshalledParams.Words[i], word)
		}
	}
}

func Test_ParametersEmptyMarshalUnmarshal(t *testing.T) {
	params := parameters.NewParameters()
	params.WordCount = 0
	params.Words = []uint16{}

	marshalled, err := params.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal empty SMB parameters: %v", err)
	}

	if len(marshalled) != 1 {
		t.Errorf("Unexpected marshalled length for empty parameters: %d != 1", len(marshalled))
	}

	if marshalled[0] != 0 {
		t.Errorf("WordCount mismatch in marshalled data: %d != 0", marshalled[0])
	}

	// Unmarshal and verify
	unmarshalledParams := parameters.NewParameters()
	_, err = unmarshalledParams.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Failed to unmarshal empty SMB parameters: %v", err)
	}

	if unmarshalledParams.WordCount != 0 {
		t.Errorf("WordCount mismatch: %d != 0", unmarshalledParams.WordCount)
	}

	if len(unmarshalledParams.Words) != 0 {
		t.Errorf("Words should be empty, got length: %d", len(unmarshalledParams.Words))
	}
}

func Test_ParametersUnmarshalWithInsufficientData(t *testing.T) {
	// Create a buffer with WordCount = 3 but only enough data for 2 words
	buf := bytes.NewBuffer([]byte{})
	buf.WriteByte(3)              // WordCount = 3
	buf.Write([]byte{0x34, 0x12}) // First word: 0x1234
	buf.Write([]byte{0x78, 0x56}) // Second word: 0x5678
	// Missing third word

	params := parameters.NewParameters()
	_, err := params.Unmarshal(buf.Bytes())
	if err == nil {
		t.Errorf("Expected error when unmarshalling with insufficient data, got nil")
	}
}

func Test_ParametersUnmarshalEmptyData(t *testing.T) {
	params := parameters.NewParameters()
	_, err := params.Unmarshal([]byte{})

	if err == nil {
		t.Errorf("Expected error when unmarshalling empty data, got nil")
	}

	if params.WordCount != 0 {
		t.Errorf("Expected WordCount to be 0, got: %d", params.WordCount)
	}

	if len(params.Words) != 0 {
		t.Errorf("Expected Words to be empty, got length: %d", len(params.Words))
	}
}
