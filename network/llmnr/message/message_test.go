package message_test

import (
	"encoding/hex"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/llmnr/class"
	"github.com/TheManticoreProject/Manticore/network/llmnr/domain_name"
	"github.com/TheManticoreProject/Manticore/network/llmnr/errors"
	"github.com/TheManticoreProject/Manticore/network/llmnr/llmnr_type"
	"github.com/TheManticoreProject/Manticore/network/llmnr/message"
	"github.com/TheManticoreProject/Manticore/network/llmnr/message/header"
	"github.com/TheManticoreProject/Manticore/network/llmnr/question"
	"github.com/TheManticoreProject/Manticore/network/llmnr/resourcerecord"
)

func TestNewMessage(t *testing.T) {
	msg := message.NewMessage()

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

func TestMessageCreation(t *testing.T) {
	header := header.Header{
		Identifier: uint16(rand.Intn(65536)),
		Flags:      0x0100,
		QDCount:    1,
		ANCount:    0,
		NSCount:    0,
		ARCount:    0,
	}

	message := message.Message{
		Header: header,
		Questions: []question.Question{
			{
				Name:  "example.com",
				Type:  llmnr_type.TypeA,
				Class: class.ClassIN,
			},
		},
	}

	if message.Header.Identifier != header.Identifier {
		t.Errorf("Message header Identifier mismatch. Got %d, want %d", message.Header.Identifier, header.Identifier)
	}

	if len(message.Questions) != 1 {
		t.Fatalf("Expected 1 question, got %d", len(message.Questions))
	}

	question := message.Questions[0]
	if question.Name != "example.com" {
		t.Errorf("Question name mismatch. Got %s, want %s", question.Name, "example.com")
	}
	if question.Type != llmnr_type.TypeA {
		t.Errorf("Question type mismatch. Got %d, want %d", question.Type, llmnr_type.TypeA)
	}
	if question.Class != class.ClassIN {
		t.Errorf("Question class mismatch. Got %d, want %d", question.Class, class.ClassIN)
	}
}

func TestMessageResponse(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	header := header.Header{
		Identifier: uint16(rand.Intn(65536)),
		Flags:      0x0100,
		QDCount:    1,
		ANCount:    0,
		NSCount:    0,
		ARCount:    0,
	}

	m := message.Message{
		Header: header,
		Questions: []question.Question{
			{
				Name:  "example.com",
				Type:  llmnr_type.TypeA,
				Class: class.ClassIN,
			},
		},
	}

	response := message.CreateResponseFromMessage(&m)
	if response.Header.Identifier != m.Header.Identifier {
		t.Errorf("Response header Identifier mismatch. Got %d, want %d", response.Header.Identifier, m.Header.Identifier)
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
	if question.Type != llmnr_type.TypeA {
		t.Errorf("Response question type mismatch. Got %d, want %d", question.Type, llmnr_type.TypeA)
	}
	if question.Class != class.ClassIN {
		t.Errorf("Response question class mismatch. Got %d, want %d", question.Class, class.ClassIN)
	}
}

func TestAddQuestion(t *testing.T) {
	msg := message.NewMessage()

	// Test adding valid question
	err := msg.AddQuestion("host.local", llmnr_type.TypeA, class.ClassIN)
	if err != nil {
		t.Errorf("AddQuestion() unexpected error: %v", err)
	}
	if len(msg.Questions) != 1 {
		t.Errorf("expected 1 question, got %d", len(msg.Questions))
	}
	if msg.Header.QDCount != 1 {
		t.Errorf("expected QDCount=1, got %d", msg.Header.QDCount)
	}

	// Test adding invalid question
	err = msg.AddQuestion(
		"thisnameiswaytoolongforavaliddomainnameandshouldcauseanerrorwhentriedtobeusedintheprogram.com",
		llmnr_type.TypeA,
		class.ClassIN,
	)
	if err != errors.ErrLabelTooLong {
		t.Errorf("AddQuestion() error = %v, want %v", err, errors.ErrLabelTooLong)
	}
}

func TestAddAnswer(t *testing.T) {
	msg := message.NewMessage()

	// Test adding valid answer
	rr := resourcerecord.ResourceRecord{
		Name:     "host.local",
		Type:     llmnr_type.TypeA,
		Class:    class.ClassIN,
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
	if msg.Header.ANCount != 1 {
		t.Errorf("expected ANCount=1, got %d", msg.Header.ANCount)
	}

	// Test adding invalid answer
	rr.Name = "thisnameiswaytoolongforavaliddomainnameandshouldcauseanerrorwhentriedtobeusedintheprogram.com"
	err = msg.AddAnswer(rr)
	if err != errors.ErrLabelTooLong && err != errors.ErrNameTooLong {
		t.Errorf("AddAnswer() error = %v, want %v", err, errors.ErrLabelTooLong)
	}
}

// TestUnmarshalMultipleRecords exercises the Marshal -> Unmarshal round trip
// with more than one question and at least one answer. The prior
// implementation of Message.Unmarshal overwrote its cumulative offset on
// every nested Unmarshal, so the second question's bytes were parsed from
// the wrong position and later sections got corrupted. This test fails on
// the broken code and passes once bytesRead accumulates correctly.
func TestUnmarshalMultipleRecords(t *testing.T) {
	orig := message.NewMessage()
	orig.Header.Identifier = 0x1234
	orig.Header.Flags = 0x0100

	if err := orig.AddQuestion("host.local", llmnr_type.TypeA, class.ClassIN); err != nil {
		t.Fatalf("AddQuestion 1 failed: %v", err)
	}
	if err := orig.AddQuestion("other.local", llmnr_type.TypeAAAA, class.ClassIN); err != nil {
		t.Fatalf("AddQuestion 2 failed: %v", err)
	}
	if err := orig.AddAnswer(resourcerecord.ResourceRecord{
		Name:     "host.local",
		Type:     llmnr_type.TypeA,
		Class:    class.ClassIN,
		TTL:      300,
		RDLength: 4,
		RData:    []byte{192, 168, 1, 1},
	}); err != nil {
		t.Fatalf("AddAnswer failed: %v", err)
	}

	wire, err := orig.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	parsed := message.NewMessage()
	n, err := parsed.Unmarshal(wire)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if n != len(wire) {
		t.Errorf("Unmarshal returned %d bytes read, want %d", n, len(wire))
	}

	if len(parsed.Questions) != 2 {
		t.Fatalf("expected 2 questions, got %d", len(parsed.Questions))
	}
	if parsed.Questions[0].Name != "host.local" {
		t.Errorf("question 0 name mismatch: got %q", parsed.Questions[0].Name)
	}
	if parsed.Questions[1].Name != "other.local" {
		t.Errorf("question 1 name mismatch: got %q", parsed.Questions[1].Name)
	}
	if len(parsed.Answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(parsed.Answers))
	}
	if parsed.Answers[0].Name != "host.local" {
		t.Errorf("answer name mismatch: got %q", parsed.Answers[0].Name)
	}
}

// llmnrLiveResponseHex is a real LLMNR response captured from a Windows Server
// 2016 host (owner name "TMP-W-2016") replying to a Type A query. This host
// does not compress the answer NAME (it repeats the full name), so it exercises
// the uncompressed path end-to-end against genuine traffic.
//
//	Header:   ID=0x37c8, Flags=0x8000 (QR), QD=1, AN=1
//	Question: TMP-W-2016  A IN
//	Answer:   TMP-W-2016  A IN  TTL=30  RDATA=10.7.0.10
const llmnrLiveResponseHex = "37c8800000010001000000000a544d502d572d3230313600000100010a544d502d572d3230313600000100010000001e00040a07000a"

// llmnrCompressedResponseHex is a hand-built LLMNR response modelled on the live
// capture above, but with the answer NAME replaced by a 0xC0 compression pointer
// (0xC00C) back to the question name at offset 12. It is the known-answer packet
// for the compression path: the prior code, which fed a sub-slice into the name
// decoder, resolves this pointer against the wrong origin and produces garbage.
//
//	Answer NAME = 0xC0 0x0C -> offset 12 (the question name "TMP-W-2016").
const llmnrCompressedResponseHex = "37c8800000010001000000000a544d502d572d323031360000010001c00c000100010000001e00040a07000a"

// TestUnmarshalLiveWindowsResponse decodes the real captured Windows LLMNR
// response and confirms the full message (question + answer RR) is parsed
// correctly, including the owner name and the A record.
func TestUnmarshalLiveWindowsResponse(t *testing.T) {
	data, err := hex.DecodeString(llmnrLiveResponseHex)
	if err != nil {
		t.Fatalf("failed to decode fixture: %v", err)
	}

	msg := message.NewMessage()
	n, err := msg.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Unmarshal read %d bytes, want %d", n, len(data))
	}

	if !msg.IsResponse() {
		t.Errorf("expected message to be a response")
	}
	if len(msg.Questions) != 1 || msg.Questions[0].Name != "TMP-W-2016" {
		t.Fatalf("unexpected questions: %+v", msg.Questions)
	}
	if len(msg.Answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(msg.Answers))
	}
	if msg.Answers[0].Name != "TMP-W-2016" {
		t.Errorf("answer name = %q; want %q", msg.Answers[0].Name, "TMP-W-2016")
	}
	if msg.Answers[0].Type != llmnr_type.TypeA {
		t.Errorf("answer type = %v; want A", msg.Answers[0].Type)
	}
	if got := net.IP(msg.Answers[0].RData).String(); got != "10.7.0.10" {
		t.Errorf("answer RDATA = %q; want %q", got, "10.7.0.10")
	}
}

// TestUnmarshalCompressedAnswerName is the known-answer test for a message whose
// answer NAME is a 0xC0 compression pointer back to the question name. It must
// decode to the same owner name and A record as the uncompressed capture.
func TestUnmarshalCompressedAnswerName(t *testing.T) {
	data, err := hex.DecodeString(llmnrCompressedResponseHex)
	if err != nil {
		t.Fatalf("failed to decode fixture: %v", err)
	}

	msg := message.NewMessage()
	n, err := msg.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Unmarshal read %d bytes, want %d", n, len(data))
	}

	if len(msg.Questions) != 1 || msg.Questions[0].Name != "TMP-W-2016" {
		t.Fatalf("unexpected questions: %+v", msg.Questions)
	}
	if len(msg.Answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(msg.Answers))
	}
	// The compressed answer NAME must resolve to the question name.
	if msg.Answers[0].Name != "TMP-W-2016" {
		t.Errorf("compressed answer name = %q; want %q", msg.Answers[0].Name, "TMP-W-2016")
	}
	if msg.Answers[0].Type != llmnr_type.TypeA {
		t.Errorf("answer type = %v; want A", msg.Answers[0].Type)
	}
	if got := net.IP(msg.Answers[0].RData).String(); got != "10.7.0.10" {
		t.Errorf("answer RDATA = %q; want %q", got, "10.7.0.10")
	}
}

// TestUnmarshalPointerLoopMessage ensures a message carrying a malformed
// compression pointer loop is rejected with an error instead of hanging or
// panicking.
func TestUnmarshalPointerLoopMessage(t *testing.T) {
	// Header (QD=1, AN=0) followed by a question whose name at offset 12 is a
	// self-referential pointer (0xC0 0x0C -> offset 12), then type A / class IN.
	data := []byte{
		0x37, 0xc8, // ID
		0x00, 0x00, // Flags
		0x00, 0x01, // QDCount
		0x00, 0x00, // ANCount
		0x00, 0x00, // NSCount
		0x00, 0x00, // ARCount
		0xC0, 0x0C, // name: self pointer to offset 12
		0x00, 0x01, // type A
		0x00, 0x01, // class IN
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		msg := message.NewMessage()
		if _, err := msg.Unmarshal(data); err == nil {
			t.Errorf("expected error for pointer loop, got nil")
		}
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("Unmarshal did not terminate on pointer loop")
	}
}

func TestValidate(t *testing.T) {
	msg := message.NewMessage()

	// Test valid message
	err := msg.AddQuestion("host.local", llmnr_type.TypeA, class.ClassIN)
	if err != nil {
		t.Fatalf("AddQuestion() unexpected error: %v", err)
	}

	rr := resourcerecord.ResourceRecord{
		Name:     "host.local",
		Type:     llmnr_type.TypeA,
		Class:    class.ClassIN,
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
	msg.Header.QDCount = 2 // Manually break the count
	err = msg.Validate()
	if err != errors.ErrInvalidMessage {
		t.Errorf("Validate() error = %v, want %v", err, errors.ErrInvalidMessage)
	}
}

func TestValidateChecksAuthorityAndAdditional(t *testing.T) {
	// Label of 64 'a' chars — exceeds MaxLabelLength (63).
	tooLongLabel := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	mkRR := func(name string) resourcerecord.ResourceRecord {
		return resourcerecord.ResourceRecord{
			Name:     domain_name.DomainName(name),
			Type:     llmnr_type.TypeA,
			Class:    class.ClassIN,
			TTL:      30,
			RDLength: 4,
			RData:    []byte{127, 0, 0, 1},
		}
	}

	// Invalid name in Authority section.
	msgAuth := message.NewMessage()
	msgAuth.Authority = append(msgAuth.Authority, mkRR(tooLongLabel))
	msgAuth.Header.NSCount = 1
	if err := msgAuth.Validate(); err != errors.ErrLabelTooLong {
		t.Errorf("Validate() returned %v for invalid Authority name, want %v", err, errors.ErrLabelTooLong)
	}

	// Invalid name in Additional section.
	msgAdd := message.NewMessage()
	msgAdd.Additional = append(msgAdd.Additional, mkRR(tooLongLabel))
	msgAdd.Header.ARCount = 1
	if err := msgAdd.Validate(); err != errors.ErrLabelTooLong {
		t.Errorf("Validate() returned %v for invalid Additional name, want %v", err, errors.ErrLabelTooLong)
	}
}
