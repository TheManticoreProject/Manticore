package client

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/pdu"
)

func TestFragmentRequestSingle(t *testing.T) {
	tmpl := pdu.NewHeader(pdu.PacketTypeRequest)
	stub := []byte{1, 2, 3, 4}
	frags, err := fragmentRequest(tmpl, stub, 4096)
	if err != nil {
		t.Fatalf("fragmentRequest: %v", err)
	}
	if len(frags) != 1 {
		t.Fatalf("got %d fragments, want 1", len(frags))
	}
	f := frags[0]
	if f.Header.Flags1.Has(pdu.Flags1Frag) || f.Header.Flags1.Has(pdu.Flags1LastFrag) {
		t.Errorf("single fragment must set neither frag nor lastfrag, got %s", f.Header.Flags1)
	}
	if f.Header.FragmentNumber != 0 {
		t.Errorf("fragnum = %d, want 0", f.Header.FragmentNumber)
	}
	if !bytes.Equal(f.Body, stub) {
		t.Errorf("body = % x, want % x", f.Body, stub)
	}
}

func TestFragmentRequestEmptyStub(t *testing.T) {
	frags, err := fragmentRequest(pdu.NewHeader(pdu.PacketTypeRequest), nil, 4096)
	if err != nil {
		t.Fatalf("fragmentRequest: %v", err)
	}
	if len(frags) != 1 {
		t.Fatalf("got %d fragments, want 1", len(frags))
	}
	if len(frags[0].Body) != 0 {
		t.Errorf("empty stub should yield empty body, got %d bytes", len(frags[0].Body))
	}
}

func TestFragmentRequestMulti(t *testing.T) {
	stub := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	// maxBody = maxPDU - HeaderSize = 4 -> 3 fragments (4, 4, 2 bytes).
	frags, err := fragmentRequest(pdu.NewHeader(pdu.PacketTypeRequest), stub, pdu.HeaderSize+4)
	if err != nil {
		t.Fatalf("fragmentRequest: %v", err)
	}
	if len(frags) != 3 {
		t.Fatalf("got %d fragments, want 3", len(frags))
	}
	var reassembled []byte
	for i, f := range frags {
		if f.Header.FragmentNumber != uint16(i) {
			t.Errorf("fragment %d has fragnum %d", i, f.Header.FragmentNumber)
		}
		if !f.Header.Flags1.Has(pdu.Flags1Frag) {
			t.Errorf("fragment %d missing frag flag", i)
		}
		if len(f.Body) > 4 {
			t.Errorf("fragment %d body %d bytes exceeds maxBody 4", i, len(f.Body))
		}
		isLast := i == len(frags)-1
		if got := f.Header.Flags1.Has(pdu.Flags1LastFrag); got != isLast {
			t.Errorf("fragment %d lastfrag = %v, want %v", i, got, isLast)
		}
		reassembled = append(reassembled, f.Body...)
	}
	if !bytes.Equal(reassembled, stub) {
		t.Errorf("reassembled bodies = % x, want % x", reassembled, stub)
	}
}

func TestFragmentRequestPreservesTemplateFlags(t *testing.T) {
	tmpl := pdu.NewHeader(pdu.PacketTypeRequest)
	tmpl.Flags1 |= pdu.Flags1Idempotent
	frags, err := fragmentRequest(tmpl, make([]byte, 9), pdu.HeaderSize+4)
	if err != nil {
		t.Fatalf("fragmentRequest: %v", err)
	}
	if len(frags) < 2 {
		t.Fatalf("expected multiple fragments, got %d", len(frags))
	}
	for i, f := range frags {
		if !f.Header.Flags1.Has(pdu.Flags1Idempotent) {
			t.Errorf("fragment %d lost the idempotent flag", i)
		}
	}
}

func TestFragmentRequestRejectsTinyMaxPDU(t *testing.T) {
	if _, err := fragmentRequest(pdu.NewHeader(pdu.PacketTypeRequest), []byte{1}, pdu.HeaderSize); err == nil {
		t.Fatal("expected error when maxPDU leaves no room for a body, got nil")
	}
}

func TestReassemblerSingle(t *testing.T) {
	var r responseReassembler
	h := pdu.NewHeader(pdu.PacketTypeResponse) // neither frag nor lastfrag
	r.add(h, []byte("whole response"))
	if !r.complete() {
		t.Fatal("single unfragmented response should be complete")
	}
	if got := r.assemble(); string(got) != "whole response" {
		t.Fatalf("assemble = %q", got)
	}
}

func TestReassemblerMultiOutOfOrder(t *testing.T) {
	var r responseReassembler

	last := pdu.NewHeader(pdu.PacketTypeResponse)
	last.Flags1 |= pdu.Flags1Frag | pdu.Flags1LastFrag
	last.FragmentNumber = 1
	r.add(last, []byte("WORLD"))
	if r.complete() {
		t.Fatal("must not be complete before fragnum 0 arrives")
	}

	first := pdu.NewHeader(pdu.PacketTypeResponse)
	first.Flags1 |= pdu.Flags1Frag
	first.FragmentNumber = 0
	r.add(first, []byte("HELLO"))

	if !r.complete() {
		t.Fatal("should be complete once both fragments are present")
	}
	if got := r.assemble(); string(got) != "HELLOWORLD" {
		t.Fatalf("assemble = %q, want HELLOWORLD", got)
	}
}

func TestReassemblerIncomplete(t *testing.T) {
	var r responseReassembler
	mid := pdu.NewHeader(pdu.PacketTypeResponse)
	mid.Flags1 |= pdu.Flags1Frag // not last
	mid.FragmentNumber = 0
	r.add(mid, []byte("partial"))
	if r.complete() {
		t.Fatal("a fragment stream with no lastfrag must not be complete")
	}
}
