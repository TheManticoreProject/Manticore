package messages

import (
	"encoding/asn1"
	"math"
	"testing"
	"time"
)

// TestKRBApplicationBodyEncodesZeroUsecAndSeqNumber checks that usec and
// seq-number set to 0 are encoded (and decoded as present), while unset ones
// are omitted, for both KRB-SAFE and the KRB-PRIV enc-part.
func TestKRBApplicationBodyEncodesZeroUsecAndSeqNumber(t *testing.T) {
	zeroUsec, zeroSeq := 0, 0
	set := KRBApplicationBody{
		UserData:       []byte("data"),
		Timestamp:      time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC),
		Usec:           &zeroUsec,
		SequenceNumber: &zeroSeq,
		SenderAddress:  HostAddress{AddrType: 2, Address: []byte{192, 0, 2, 1}},
	}
	unset := set
	unset.Usec, unset.SequenceNumber = nil, nil

	safeWire, err := (&KRBSafe{SafeBody: set, Checksum: Checksum{CKSumType: 16, Checksum: []byte{1}}}).Marshal()
	if err != nil {
		t.Fatalf("KRBSafe.Marshal: %v", err)
	}
	var safe KRBSafe
	if _, err := safe.Unmarshal(safeWire); err != nil {
		t.Fatalf("KRBSafe.Unmarshal: %v", err)
	}
	privWire, err := (*EncKRBPrivPart)(&set).Marshal()
	if err != nil {
		t.Fatalf("EncKRBPrivPart.Marshal: %v", err)
	}
	var priv EncKRBPrivPart
	if _, err := priv.Unmarshal(privWire); err != nil {
		t.Fatalf("EncKRBPrivPart.Unmarshal: %v", err)
	}
	for name, got := range map[string]KRBApplicationBody{"KRB-SAFE": safe.SafeBody, "KRB-PRIV": KRBApplicationBody(priv)} {
		if got.Usec == nil || *got.Usec != 0 {
			t.Errorf("%s usec = %v, want present 0", name, got.Usec)
		}
		if got.SequenceNumber == nil || *got.SequenceNumber != 0 {
			t.Errorf("%s seq-number = %v, want present 0", name, got.SequenceNumber)
		}
	}

	der, err := unset.Marshal()
	if err != nil {
		t.Fatalf("Marshal unset: %v", err)
	}
	var back KRBApplicationBody
	if _, err := back.Unmarshal(der); err != nil {
		t.Fatalf("Unmarshal unset: %v", err)
	}
	if back.Usec != nil || back.SequenceNumber != nil {
		t.Errorf("unset usec/seq-number decoded as present: %v / %v", back.Usec, back.SequenceNumber)
	}
}

// TestKRBSafeKeepsReceivedSafeBodyBytes checks that Unmarshal exposes the
// safe-body exactly as received.
func TestKRBSafeKeepsReceivedSafeBodyBytes(t *testing.T) {
	seq := 7
	body := KRBApplicationBody{UserData: []byte("x"), SequenceNumber: &seq, SenderAddress: HostAddress{AddrType: 2, Address: []byte{10, 0, 0, 1}}}
	want, err := body.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	wire, err := (&KRBSafe{SafeBody: body, Checksum: Checksum{CKSumType: 16, Checksum: []byte{1}}}).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var safe KRBSafe
	if _, err := safe.Unmarshal(wire); err != nil {
		t.Fatal(err)
	}
	if string(safe.SafeBodyRaw) != string(want) {
		t.Fatalf("SafeBodyRaw = %x, want %x", safe.SafeBodyRaw, want)
	}
}

// TestKRBApplicationBodyRejectsOutOfRangeValues checks that a usec outside
// Microseconds and a seq-number outside UInt32 are refused on decode rather
// than converted, which on a 32-bit build would wrap the sequence number.
func TestKRBApplicationBodyRejectsOutOfRangeValues(t *testing.T) {
	base := krbApplicationBodyWire{
		UserData:       []byte("data"),
		Timestamp:      time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC),
		Usec:           krbApplicationBodyAbsent,
		SequenceNumber: krbApplicationBodyAbsent,
		SenderAddress:  HostAddress{AddrType: 2, Address: []byte{192, 0, 2, 1}},
	}
	cases := map[string]func(*krbApplicationBodyWire){
		"usec 1000000":    func(w *krbApplicationBodyWire) { w.Usec = 1000000 },
		"usec -1":         func(w *krbApplicationBodyWire) { w.Usec = -1 },
		"seq-number 2^32": func(w *krbApplicationBodyWire) { w.SequenceNumber = 1 << 32 },
		"seq-number -1":   func(w *krbApplicationBodyWire) { w.SequenceNumber = -1 },
	}
	for name, mutate := range cases {
		w := base
		mutate(&w)
		data, err := asn1.Marshal(w)
		if err != nil {
			t.Fatalf("%s: marshal: %v", name, err)
		}
		var b KRBApplicationBody
		if _, err := b.Unmarshal(data); err == nil {
			t.Errorf("%s: Unmarshal accepted it, want an error", name)
		}
	}

	w := base
	w.Usec, w.SequenceNumber = 999999, 1<<32-1
	data, err := asn1.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	var b KRBApplicationBody
	if _, err := b.Unmarshal(data); err != nil {
		t.Fatalf("Unmarshal of the largest valid values: %v", err)
	}
	if *b.Usec != 999999 || uint32(*b.SequenceNumber) != 1<<32-1 {
		t.Errorf("decoded usec=%d seq=%d", *b.Usec, *b.SequenceNumber)
	}
}

// TestKRBApplicationBodySeqNumberHighBitRoundTrips checks that a sequence number
// with the top bit set round-trips. On a 32-bit build such a value is held as a
// negative int, and must still go on the wire as the UInt32 it stands for.
func TestKRBApplicationBodySeqNumberHighBitRoundTrips(t *testing.T) {
	high := uint32(math.MaxInt32) + 2
	seq := int(int32(high)) // negative on every build; what a 32-bit caller holds
	body := KRBApplicationBody{
		UserData:       []byte("data"),
		Timestamp:      time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC),
		SequenceNumber: &seq,
		SenderAddress:  HostAddress{AddrType: 2, Address: []byte{192, 0, 2, 1}},
	}
	data, err := body.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var w krbApplicationBodyWire
	if _, err := asn1.Unmarshal(data, &w); err != nil {
		t.Fatal(err)
	}
	if w.SequenceNumber != int64(high) {
		t.Fatalf("seq-number on the wire = %d, want %d", w.SequenceNumber, high)
	}
	var out KRBApplicationBody
	if _, err := out.Unmarshal(data); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if uint32(*out.SequenceNumber) != high {
		t.Errorf("decoded seq-number = %d, want %d as a uint32", *out.SequenceNumber, high)
	}
}
