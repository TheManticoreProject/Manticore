package structures_test

import (
	"encoding/binary"
	"errors"
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/qgy/qgyolaus/structures"
)

var codec, _ = ccsid.Lookup(ccsid.CCSID37)

func fixed(s string, n int) []byte {
	b, err := ccsid.EncodeFixed(codec, s, n)
	if err != nil {
		panic(err)
	}
	return b
}

// head builds offsets 0..11: name, user/group indicator, members indicator.
func head(name string, group, members byte) []byte {
	return append(fixed(name, 10), group, members)
}

// groups builds a BINARY(4) count and the 16-entry CHAR(10) array; slots
// past the listed names hold a sentinel that must be ignored.
func groups(count int32, names ...string) []byte {
	b := binary.BigEndian.AppendUint32(nil, uint32(count))
	for i := 0; i < structures.MaxGroupProfiles; i++ {
		name := "UNUSED"
		if i < len(names) {
			name = names[i]
		}
		b = append(b, fixed(name, 10)...)
	}
	return b
}

func TestAUTU0100(t *testing.T) {
	r, err := structures.DecodeAUTU0100(head("ALICE", 0xF0, 0xF0), codec)
	if err != nil || r != (structures.AUTU0100{Name: "ALICE"}) {
		t.Fatalf("user = %+v, %v", r, err)
	}
	r, err = structures.DecodeAUTU0100(head("DEVGRP", 0xF1, 0xF1), codec)
	if err != nil || !r.GroupProfile || !r.GroupHasMembers || r.Name != "DEVGRP" {
		t.Fatalf("group = %+v, %v", r, err)
	}
	// The full 10-byte name field is used.
	if r, _ := structures.DecodeAUTU0100(head("ABCDEFGHIJ", 0xF0, 0xF0), codec); r.Name != "ABCDEFGHIJ" {
		t.Errorf("10-character name = %q", r.Name)
	}
}

func TestAUTU0150(t *testing.T) {
	b := append(head("ALICE", 0xF0, 0xF0), fixed("Application developer", 50)...)
	r, err := structures.DecodeAUTU0150(b, codec)
	if err != nil || r.Name != "ALICE" || r.Description != "Application developer" {
		t.Fatalf("AUTU0150 = %+v, %v", r, err)
	}
	// Offset 61, the last description byte, is part of the description.
	last := append(head("A", 0xF0, 0xF0), fixed("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxZ", 50)...)
	if r, _ := structures.DecodeAUTU0150(last, codec); r.Description[49] != 'Z' {
		t.Errorf("description end = %q", r.Description)
	}
}

func TestAUTU0200(t *testing.T) {
	b := append(head("ALICE", 0xF0, 0xF0), groups(2, "DEV", "OPS")...)
	r, err := structures.DecodeAUTU0200(b, codec)
	if err != nil || !reflect.DeepEqual(r.GroupProfiles, []string{"DEV", "OPS"}) {
		t.Fatalf("AUTU0200 = %+v, %v", r, err)
	}
	all := []string{}
	for i := 0; i < 16; i++ {
		all = append(all, "G"+string(rune('A'+i)))
	}
	b = append(head("BOB", 0xF0, 0xF0), groups(16, all...)...)
	if r, err := structures.DecodeAUTU0200(b, codec); err != nil || len(r.GroupProfiles) != 16 || r.GroupProfiles[15] != "GP" {
		t.Errorf("16 groups = %v, %v", r.GroupProfiles, err)
	}
	if r, err := structures.DecodeAUTU0200(append(head("C", 0xF0, 0xF0), groups(0)...), codec); err != nil || len(r.GroupProfiles) != 0 {
		t.Errorf("no groups = %v, %v", r.GroupProfiles, err)
	}
}

func TestAUTU0250(t *testing.T) {
	b := append(head("ALICE", 0xF1, 0xF0), fixed("Lead", 50)...)
	b = append(b, 0xEE, 0xEE) // reserved bytes 62..63 are ignored
	b = append(b, groups(1, "DEV")...)
	r, err := structures.DecodeAUTU0250(b, codec)
	if err != nil || !r.GroupProfile || r.Description != "Lead" || !reflect.DeepEqual(r.GroupProfiles, []string{"DEV"}) {
		t.Fatalf("AUTU0250 = %+v, %v", r, err)
	}
}

func TestDecodersRejectWrongLength(t *testing.T) {
	for name, tc := range map[string]struct {
		length int
		decode func([]byte) error
	}{
		"AUTU0100": {12, func(b []byte) error { _, err := structures.DecodeAUTU0100(b, codec); return err }},
		"AUTU0150": {62, func(b []byte) error { _, err := structures.DecodeAUTU0150(b, codec); return err }},
		"AUTU0200": {176, func(b []byte) error { _, err := structures.DecodeAUTU0200(b, codec); return err }},
		"AUTU0250": {228, func(b []byte) error { _, err := structures.DecodeAUTU0250(b, codec); return err }},
	} {
		for _, n := range []int{0, tc.length - 1, tc.length + 1} {
			b := make([]byte, n)
			if err := tc.decode(b); !errors.Is(err, structures.ErrMalformed) {
				t.Errorf("%s with %d bytes: %v", name, n, err)
			}
		}
	}
	// A QSYLAUTU AUTU0200 entry (228 bytes) must not decode as QGYOLAUS
	// AUTU0200.
	if _, err := structures.DecodeAUTU0200(make([]byte, 228), codec); !errors.Is(err, structures.ErrMalformed) {
		t.Error("228-byte entry accepted as QGYOLAUS AUTU0200")
	}
}

func TestDecodersRejectBadValues(t *testing.T) {
	for name, b := range map[string][]byte{
		"ASCII indicator": head("A", '0', 0xF0),
		"blank indicator": head("A", 0xF0, 0x40),
		"indicator 2":     head("A", 0xF2, 0xF0),
	} {
		if _, err := structures.DecodeAUTU0100(b, codec); !errors.Is(err, structures.ErrMalformed) {
			t.Errorf("%s: %v", name, err)
		}
	}
	for _, count := range []int32{-1, 17} {
		b := append(head("A", 0xF0, 0xF0), groups(count)...)
		if _, err := structures.DecodeAUTU0200(b, codec); !errors.Is(err, structures.ErrMalformed) {
			t.Errorf("group count %d accepted", count)
		}
	}
}
