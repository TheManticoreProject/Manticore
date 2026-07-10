package llmnr_test

import (
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/llmnr/class"
	"github.com/TheManticoreProject/Manticore/network/llmnr/llmnr_type"
	"github.com/TheManticoreProject/Manticore/network/llmnr/message"
)

// llmnrLiveQueryHex is the LLMNR Type A query for owner name "TMP-W-2016" that is
// the counterpart to the captured response fixture in message_test.go. It is the
// exact wire form a standard LLMNR client emits when resolving TMP-W-2016 (the
// Windows Server 2016 host in the lab, domain TMP-W-2016.local):
//
//	Header:   ID=0x37c8, Flags=0x0000 (standard query), QD=1
//	Question: TMP-W-2016  A IN
//
// It is the known-answer packet for the query path: parsing it must recover the
// question exactly, and re-marshalling the equivalent message must reproduce these
// bytes byte-for-byte.
const llmnrLiveQueryHex = "37c8000000010000000000000a544d502d572d323031360000010001"

// TestUnmarshalLiveQuery decodes the genuine LLMNR query fixture and asserts the
// header and the single question parse exactly.
func TestUnmarshalLiveQuery(t *testing.T) {
	data, err := hex.DecodeString(llmnrLiveQueryHex)
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

	if !msg.IsQuery() {
		t.Error("expected message to be a query")
	}
	if msg.Header.Identifier != 0x37c8 {
		t.Errorf("identifier = %#04x, want 0x37c8", msg.Header.Identifier)
	}
	if len(msg.Questions) != 1 {
		t.Fatalf("questions = %d, want 1", len(msg.Questions))
	}
	q := msg.Questions[0]
	if q.Name != "TMP-W-2016" {
		t.Errorf("question name = %q, want %q", q.Name, "TMP-W-2016")
	}
	if q.Type != llmnr_type.TypeA {
		t.Errorf("question type = %v, want A", q.Type)
	}
	if q.Class != class.ClassIN {
		t.Errorf("question class = %v, want IN", q.Class)
	}
	if len(msg.Answers) != 0 {
		t.Errorf("answers = %d, want 0", len(msg.Answers))
	}
}

// TestMarshalLiveQuery is the re-marshal side of the known-answer test: a query
// built the same way a client builds it must serialise to the exact fixture bytes.
func TestMarshalLiveQuery(t *testing.T) {
	want, err := hex.DecodeString(llmnrLiveQueryHex)
	if err != nil {
		t.Fatalf("failed to decode fixture: %v", err)
	}

	msg := message.NewMessage()
	msg.Header.Identifier = 0x37c8
	msg.SetQuery()
	if err := msg.AddQuestion("TMP-W-2016", llmnr_type.TypeA, class.ClassIN); err != nil {
		t.Fatalf("AddQuestion failed: %v", err)
	}

	got, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if hex.EncodeToString(got) != hex.EncodeToString(want) {
		t.Errorf("Marshal() = %s, want %s", hex.EncodeToString(got), hex.EncodeToString(want))
	}
}

// TestQueryRoundTripMultiQuestion exercises a multi-question query end-to-end: the
// second question's name shares no compression with the first, so both label
// sequences are encoded in full. Parsing the marshalled bytes must recover both
// questions with their distinct types.
func TestQueryRoundTripMultiQuestion(t *testing.T) {
	msg := message.NewMessage()
	msg.Header.Identifier = 0x1234
	msg.SetQuery()
	if err := msg.AddQuestion("wpad", llmnr_type.TypeA, class.ClassIN); err != nil {
		t.Fatalf("AddQuestion(wpad) failed: %v", err)
	}
	if err := msg.AddQuestion("printer.local", llmnr_type.TypeAAAA, class.ClassIN); err != nil {
		t.Fatalf("AddQuestion(printer.local) failed: %v", err)
	}

	encoded, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	decoded := message.NewMessage()
	if _, err := decoded.Unmarshal(encoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Header.QDCount != 2 || len(decoded.Questions) != 2 {
		t.Fatalf("QDCount=%d questions=%d, want 2/2", decoded.Header.QDCount, len(decoded.Questions))
	}
	if decoded.Questions[0].Name != "wpad" || decoded.Questions[0].Type != llmnr_type.TypeA {
		t.Errorf("question[0] = %v/%v, want wpad/A", decoded.Questions[0].Name, decoded.Questions[0].Type)
	}
	if decoded.Questions[1].Name != "printer.local" || decoded.Questions[1].Type != llmnr_type.TypeAAAA {
		t.Errorf("question[1] = %v/%v, want printer.local/AAAA", decoded.Questions[1].Name, decoded.Questions[1].Type)
	}
}
