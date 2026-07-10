package server_test

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/llmnr/class"
	"github.com/TheManticoreProject/Manticore/network/llmnr/llmnr_type"
	"github.com/TheManticoreProject/Manticore/network/llmnr/message"
	"github.com/TheManticoreProject/Manticore/network/llmnr/server"
)

// llmnrLiveResponseHex is the real LLMNR Type A response captured from the
// Windows Server 2016 lab host (owner name "TMP-W-2016", A=10.7.0.10). It is the
// same fixture exercised by the message package tests and is used here to
// confirm MessageToJson renders genuine traffic correctly.
//
//	Header:   ID=0x37c8, Flags=0x8000 (QR), QD=1, AN=1
//	Question: TMP-W-2016  A IN
//	Answer:   TMP-W-2016  A IN  TTL=30  RDATA=10.7.0.10
const llmnrLiveResponseHex = "37c8800000010001000000000a544d502d572d3230313600000100010a544d502d572d3230313600000100010000001e00040a07000a"

// TestMessageToJsonShape asserts the JSON shape produced by MessageToJson for a
// query and for a response carrying an A answer. The produced JSON is unmarshalled
// back into a generic map and the key fields (ID, QR, question name/type and the
// answer name/type/A-address) are asserted.
func TestMessageToJsonShape(t *testing.T) {
	// Build a simple Type A query for "printer.local".
	query := message.NewMessage()
	query.Header.Identifier = 0x1234
	if err := query.AddQuestion("printer.local", llmnr_type.TypeA, class.ClassIN); err != nil {
		t.Fatalf("AddQuestion failed: %v", err)
	}

	data, err := server.MessageToJson(nil, query)
	if err != nil {
		t.Fatalf("MessageToJson(query) failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("query JSON is not valid: %v\n%s", err, data)
	}

	header := decoded["header"].(map[string]interface{})
	if id := header["id"].(float64); uint16(id) != 0x1234 {
		t.Errorf("query header.id = %v, want %d", id, 0x1234)
	}
	flags := header["flags"].(map[string]interface{})
	if qr := flags["qr"].(bool); qr {
		t.Errorf("query flags.qr = true, want false (a query has QR clear)")
	}

	questions := decoded["questions"].([]interface{})
	if len(questions) != 1 {
		t.Fatalf("query questions = %d, want 1", len(questions))
	}
	q0 := questions[0].(map[string]interface{})
	if name := q0["name"].(string); name != "printer.local" {
		t.Errorf("query question name = %q, want %q", name, "printer.local")
	}
	if typ := q0["type"].(string); typ != llmnr_type.TypeA.String() {
		t.Errorf("query question type = %q, want %q", typ, llmnr_type.TypeA.String())
	}

	// Decode the real captured response and render it.
	raw, err := hex.DecodeString(llmnrLiveResponseHex)
	if err != nil {
		t.Fatalf("failed to decode fixture: %v", err)
	}
	resp := message.NewMessage()
	if _, err := resp.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal(fixture) failed: %v", err)
	}

	data, err = server.MessageToJson(nil, resp)
	if err != nil {
		t.Fatalf("MessageToJson(response) failed: %v", err)
	}

	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("response JSON is not valid: %v\n%s", err, data)
	}

	header = decoded["header"].(map[string]interface{})
	if id := header["id"].(float64); uint16(id) != 0x37c8 {
		t.Errorf("response header.id = %v, want %#x", id, 0x37c8)
	}
	flags = header["flags"].(map[string]interface{})
	if qr := flags["qr"].(bool); !qr {
		t.Errorf("response flags.qr = false, want true (a response has QR set)")
	}

	answers := decoded["answers"].([]interface{})
	if len(answers) != 1 {
		t.Fatalf("response answers = %d, want 1", len(answers))
	}
	a0 := answers[0].(map[string]interface{})
	if name := a0["name"].(string); name != "TMP-W-2016" {
		t.Errorf("response answer name = %q, want %q", name, "TMP-W-2016")
	}
	if typ := a0["type"].(string); typ != llmnr_type.TypeA.String() {
		t.Errorf("response answer type = %q, want %q", typ, llmnr_type.TypeA.String())
	}
	// The A record's typed RDATA must be the dotted-quad address.
	if addr, ok := a0["rdata"].(string); !ok || addr != "10.7.0.10" {
		t.Errorf("response answer rdata = %v, want %q", a0["rdata"], "10.7.0.10")
	}
	// The raw RDATA hex must always be present.
	if hexData, ok := a0["rdata_hex"].(string); !ok || hexData != "0a07000a" {
		t.Errorf("response answer rdata_hex = %v, want %q", a0["rdata_hex"], "0a07000a")
	}
}
