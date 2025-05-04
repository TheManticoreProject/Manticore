package llmnr_test

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/llmnr"
)

func TestHeaderSerialization(t *testing.T) {
	header := llmnr.Header{
		ID:      12345,
		Flags:   0x0100,
		QDCount: 1,
		ANCount: 0,
		NSCount: 0,
		ARCount: 0,
	}

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, header)
	if err != nil {
		t.Fatalf("Failed to serialize header: %v", err)
	}

	var deserializedHeader llmnr.Header
	err = binary.Read(buf, binary.BigEndian, &deserializedHeader)
	if err != nil {
		t.Fatalf("Failed to deserialize header: %v", err)
	}

	if header != deserializedHeader {
		t.Errorf("Deserialized header does not match original. Got %+v, want %+v", deserializedHeader, header)
	}
}

func TestMessageCreation(t *testing.T) {
	header := llmnr.Header{
		ID:      uint16(rand.Intn(65536)),
		Flags:   0x0100,
		QDCount: 1,
		ANCount: 0,
		NSCount: 0,
		ARCount: 0,
	}

	message := llmnr.Message{
		Header: header,
		Questions: []llmnr.Question{
			{
				Name:  "example.com",
				Type:  llmnr.TypeA,
				Class: llmnr.ClassIN,
			},
		},
	}

	if message.Header.ID != header.ID {
		t.Errorf("Message header ID mismatch. Got %d, want %d", message.Header.ID, header.ID)
	}

	if len(message.Questions) != 1 {
		t.Fatalf("Expected 1 question, got %d", len(message.Questions))
	}

	question := message.Questions[0]
	if question.Name != "example.com" {
		t.Errorf("Question name mismatch. Got %s, want %s", question.Name, "example.com")
	}
	if question.Type != llmnr.TypeA {
		t.Errorf("Question type mismatch. Got %d, want %d", question.Type, llmnr.TypeA)
	}
	if question.Class != llmnr.ClassIN {
		t.Errorf("Question class mismatch. Got %d, want %d", question.Class, llmnr.ClassIN)
	}
}

func TestMessageResponse(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	header := llmnr.Header{
		ID:      uint16(rand.Intn(65536)),
		Flags:   0x0100,
		QDCount: 1,
		ANCount: 0,
		NSCount: 0,
		ARCount: 0,
	}

	message := llmnr.Message{
		Header: header,
		Questions: []llmnr.Question{
			{
				Name:  "example.com",
				Type:  llmnr.TypeA,
				Class: llmnr.ClassIN,
			},
		},
	}

	response := llmnr.CreateResponseFromMessage(&message)
	if response.Header.ID != message.Header.ID {
		t.Errorf("Response header ID mismatch. Got %d, want %d", response.Header.ID, message.Header.ID)
	}

	if response.Header.Flags&0x8000 == 0 {
		t.Errorf("Response header QR flag not set")
	}

	if len(response.Questions) != 0 {
		t.Fatalf("Expected 0 question in response, got %d", len(response.Questions))
	}

	if len(response.Answers) != 0 {
		t.Fatalf("Expected 0 answer in response, got %d", len(response.Answers))
	}

	response.AddAnswerClassINTypeA("example.com", "127.0.0.1")

	question := response.Questions[0]
	if question.Name != "example.com" {
		t.Errorf("Response question name mismatch. Got %s, want %s", question.Name, "example.com")
	}
	if question.Type != llmnr.TypeA {
		t.Errorf("Response question type mismatch. Got %d, want %d", question.Type, llmnr.TypeA)
	}
	if question.Class != llmnr.ClassIN {
		t.Errorf("Response question class mismatch. Got %d, want %d", question.Class, llmnr.ClassIN)
	}
}
