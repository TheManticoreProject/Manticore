package llmnr_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/llmnr"
)

func TestNewMessage(t *testing.T) {
	msg := llmnr.NewMessage()

	if msg == nil {
		t.Fatal("NewMessage returned nil")
	}

	// Check default values
	if len(msg.Questions) != 0 {
		t.Errorf("expected empty Questions, got %d items", len(msg.Questions))
	}
	if len(msg.Answers) != 0 {
		t.Errorf("expected empty Answers, got %d items", len(msg.Answers))
	}
	if len(msg.Authority) != 0 {
		t.Errorf("expected empty Authority, got %d items", len(msg.Authority))
	}
	if len(msg.Additional) != 0 {
		t.Errorf("expected empty Additional, got %d items", len(msg.Additional))
	}
}

func TestMessageFlags(t *testing.T) {
	msg := llmnr.NewMessage()

	// Test Query/Response flags
	if !msg.IsQuery() {
		t.Error("new message should be a query by default")
	}
	if msg.IsResponse() {
		t.Error("new message should not be a response by default")
	}

	msg.SetResponse()
	if !msg.IsResponse() {
		t.Error("message should be a response after SetResponse")
	}
	if msg.IsQuery() {
		t.Error("message should not be a query after SetResponse")
	}

	msg.SetQuery()
	if !msg.IsQuery() {
		t.Error("message should be a query after SetQuery")
	}
	if msg.IsResponse() {
		t.Error("message should not be a response after SetQuery")
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "valid short name",
			input:   "host",
			wantErr: nil,
		},
		{
			name:    "valid domain name",
			input:   "host.local",
			wantErr: nil,
		},
		{
			name:    "valid long name",
			input:   "this.is.a.valid.domain.name",
			wantErr: nil,
		},
		{
			name:    "label too long",
			input:   "thisnameiswaytoolongforavaliddomainnameandshouldcauseanerrorwhentriedtobeusedintheprogram.com",
			wantErr: llmnr.ErrLabelTooLong,
		},
		{
			name:    "name too long",
			input:   "a.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.very.long.name",
			wantErr: llmnr.ErrNameTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := llmnr.ValidateDomainName(tt.input)
			if err != tt.wantErr {
				t.Errorf("ValidateDomainName() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddQuestion(t *testing.T) {
	msg := llmnr.NewMessage()

	// Test adding valid question
	err := msg.AddQuestion("host.local", llmnr.TypeA, llmnr.ClassIN)
	if err != nil {
		t.Errorf("AddQuestion() unexpected error: %v", err)
	}
	if len(msg.Questions) != 1 {
		t.Errorf("expected 1 question, got %d", len(msg.Questions))
	}
	if msg.QDCount != 1 {
		t.Errorf("expected QDCount=1, got %d", msg.QDCount)
	}

	// Test adding invalid question
	err = msg.AddQuestion(
		"thisnameiswaytoolongforavaliddomainnameandshouldcauseanerrorwhentriedtobeusedintheprogram.com",
		llmnr.TypeA,
		llmnr.ClassIN,
	)
	if err != llmnr.ErrLabelTooLong {
		t.Errorf("AddQuestion() error = %v, want %v", err, llmnr.ErrLabelTooLong)
	}
}

func TestAddAnswer(t *testing.T) {
	msg := llmnr.NewMessage()

	// Test adding valid answer
	rr := llmnr.ResourceRecord{
		Name:     "host.local",
		Type:     llmnr.TypeA,
		Class:    llmnr.ClassIN,
		TTL:      300,
		RDLength: 4,
		RData:    []byte{192, 168, 1, 1},
	}

	err := msg.AddAnswer(rr)
	if err != nil {
		t.Errorf("AddAnswer() unexpected error: %v", err)
	}
	if len(msg.Answers) != 1 {
		t.Errorf("expected 1 answer, got %d", len(msg.Answers))
	}
	if msg.ANCount != 1 {
		t.Errorf("expected ANCount=1, got %d", msg.ANCount)
	}

	// Test adding invalid answer
	rr.Name = "thisnameiswaytoolongforavaliddomainnameandshouldcauseanerrorwhentriedtobeusedintheprogram.com"
	err = msg.AddAnswer(rr)
	if err != llmnr.ErrLabelTooLong {
		t.Errorf("AddAnswer() error = %v, want %v", err, llmnr.ErrLabelTooLong)
	}
}

func TestValidate(t *testing.T) {
	msg := llmnr.NewMessage()

	// Test valid message
	err := msg.AddQuestion("host.local", llmnr.TypeA, llmnr.ClassIN)
	if err != nil {
		t.Fatalf("AddQuestion() unexpected error: %v", err)
	}

	rr := llmnr.ResourceRecord{
		Name:     "host.local",
		Type:     llmnr.TypeA,
		Class:    llmnr.ClassIN,
		TTL:      300,
		RDLength: 4,
		RData:    []byte{192, 168, 1, 1},
	}
	err = msg.AddAnswer(rr)
	if err != nil {
		t.Fatalf("AddAnswer() unexpected error: %v", err)
	}

	err = msg.Validate()
	if err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	// Test invalid message (mismatched counts)
	msg.QDCount = 2 // Manually break the count
	err = msg.Validate()
	if err != llmnr.ErrInvalidMessage {
		t.Errorf("Validate() error = %v, want %v", err, llmnr.ErrInvalidMessage)
	}
}

func TestMessageEncoding(t *testing.T) {
	tests := []struct {
		name    string
		message *llmnr.Message
		wantErr bool
	}{
		{
			name: "simple query",
			message: &llmnr.Message{
				Header: llmnr.Header{
					ID:      0x1234,
					Flags:   0,
					QDCount: 1,
				},
				Questions: []llmnr.Question{
					{
						Name:  "test.local",
						Type:  llmnr.TypeA,
						Class: llmnr.ClassIN,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "response with answer",
			message: &llmnr.Message{
				Header: llmnr.Header{
					ID:      0x1234,
					Flags:   llmnr.FlagQR,
					QDCount: 1,
					ANCount: 1,
				},
				Questions: []llmnr.Question{
					{
						Name:  "test.local",
						Type:  llmnr.TypeA,
						Class: llmnr.ClassIN,
					},
				},
				Answers: []llmnr.ResourceRecord{
					{
						Name:     "test.local",
						Type:     llmnr.TypeA,
						Class:    llmnr.ClassIN,
						TTL:      300,
						RDLength: 4,
						RData:    []byte{192, 168, 1, 1},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test encoding
			encoded, err := tt.message.Encode()
			if (err != nil) != tt.wantErr {
				t.Errorf("Message.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Test decoding
			decoded, err := llmnr.DecodeMessage(encoded)
			if err != nil {
				t.Errorf("DecodeMessage() error = %v", err)
				return
			}

			// Compare original and decoded messages
			if decoded.ID != tt.message.ID {
				t.Errorf("ID mismatch: got %v, want %v", decoded.ID, tt.message.ID)
			}
			if decoded.Flags != tt.message.Flags {
				t.Errorf("Flags mismatch: got %v, want %v", decoded.Flags, tt.message.Flags)
			}

			// Compare questions
			if len(decoded.Questions) != len(tt.message.Questions) {
				t.Errorf("Questions length mismatch: got %v, want %v", len(decoded.Questions), len(tt.message.Questions))
			} else {
				for i, q := range tt.message.Questions {
					if q.Name != decoded.Questions[i].Name {
						t.Errorf("Question[%d] Name mismatch: got %v, want %v", i, decoded.Questions[i].Name, q.Name)
					}
					if q.Type != decoded.Questions[i].Type {
						t.Errorf("Question[%d] Type mismatch: got %v, want %v", i, decoded.Questions[i].Type, q.Type)
					}
					if q.Class != decoded.Questions[i].Class {
						t.Errorf("Question[%d] Class mismatch: got %v, want %v", i, decoded.Questions[i].Class, q.Class)
					}
				}
			}

			// Compare answers
			if len(decoded.Answers) != len(tt.message.Answers) {
				t.Errorf("Answers length mismatch: got %v, want %v", len(decoded.Answers), len(tt.message.Answers))
			} else {
				for i, a := range tt.message.Answers {
					if a.Name != decoded.Answers[i].Name {
						t.Errorf("Answer[%d] Name mismatch: got %v, want %v", i, decoded.Answers[i].Name, a.Name)
					}
					if a.Type != decoded.Answers[i].Type {
						t.Errorf("Answer[%d] Type mismatch: got %v, want %v", i, decoded.Answers[i].Type, a.Type)
					}
					if a.Class != decoded.Answers[i].Class {
						t.Errorf("Answer[%d] Class mismatch: got %v, want %v", i, decoded.Answers[i].Class, a.Class)
					}
					if a.TTL != decoded.Answers[i].TTL {
						t.Errorf("Answer[%d] TTL mismatch: got %v, want %v", i, decoded.Answers[i].TTL, a.TTL)
					}
					if !bytes.Equal(a.RData, decoded.Answers[i].RData) {
						t.Errorf("Answer[%d] RData mismatch: got %v, want %v", i, decoded.Answers[i].RData, a.RData)
					}
				}
			}
		})
	}
}

func TestDomainNameEncoding(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "simple name",
			input:   "test.local",
			wantErr: false,
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: false,
		},
		{
			name:    "root",
			input:   ".",
			wantErr: false,
		},
		{
			name:    "multiple labels",
			input:   "www.test.local",
			wantErr: false,
		},
		{
			name:    "label too long",
			input:   strings.Repeat("a", 64) + ".local",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := llmnr.EncodeDomainName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("encodeDomainName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			decoded, _, err := llmnr.DecodeDomainName(encoded, 0)
			if err != nil {
				t.Errorf("decodeDomainName() error = %v", err)
				return
			}

			if tt.input == "" {
				tt.input = "."
			}
			if decoded != tt.input {
				t.Errorf("Domain name mismatch: got %v, want %v", decoded, tt.input)
			}
		})
	}
}
