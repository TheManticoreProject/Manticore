package llmnr_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/llmnr"
)

func TestEncodeQuestions(t *testing.T) {
	tests := []struct {
		question llmnr.Question
		expected []byte
	}{
		{
			question: llmnr.Question{
				Name:  "example.com",
				Type:  llmnr.TypeA,
				Class: llmnr.ClassIN,
			},
			expected: []byte{
				7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0, 0, 1, 0, 1,
			},
		},
	}

	for _, test := range tests {
		t.Run("EncodeQuestions", func(t *testing.T) {
			var buf []byte
			encoded, err := llmnr.EncodeQuestion(test.question)
			if err != nil {
				t.Fatalf("failed to encode question: %v", err)
			}
			buf = append(buf, encoded...)
			if !bytes.Equal(buf, test.expected) {
				t.Errorf("EncodeQuestions = %v; want %v", buf, test.expected)
			}
		})
	}
}

func TestDecodeQuestions(t *testing.T) {
	tests := []struct {
		data     []byte
		expected llmnr.Question
	}{
		{
			data: []byte{
				7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0, 0, 1, 0, 1,
			},
			expected: llmnr.Question{
				Name:  "example.com",
				Type:  llmnr.TypeA,
				Class: llmnr.ClassIN,
			},
		},
	}

	for _, test := range tests {
		t.Run("DecodeQuestions", func(t *testing.T) {
			var offset int
			var question llmnr.Question
			question, _, err := llmnr.DecodeQuestion(test.data, offset)
			if err != nil {
				t.Fatalf("failed to decode question: %v", err)
			}
			if question != test.expected {
				t.Errorf("DecodeQuestions = %v; want %v", question, test.expected)
			}
		})
	}
}
