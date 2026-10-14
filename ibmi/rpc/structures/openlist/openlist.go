// Package openlist decodes the 80-byte open-list information structure that
// IBM i open-list APIs (QGYOLAUS, QGYGTLE and others) return.
//
// The decoder is structural only: it validates every documented field and
// maps the indicators to closed enums, but it does not require a particular
// program's record length or list status. Program packages add their own
// policies.
//
// Layout: IBM i "Open list information format",
// https://www.ibm.com/docs/ssw_ibm_i_74/apiref/oli.htm (retrieved
// 2026-10-14).
package openlist

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Length is the size of the list-information structure.
const Length = 80

// Handle is the opaque request handle of an open list. It is only valid on
// the server job that opened the list.
type Handle [4]byte

// Complete is the information-complete indicator.
type Complete uint8

const (
	completeUnknown Complete = iota
	// CompleteAccurate (C): all requested records were returned.
	CompleteAccurate
	// Incomplete (I): an interruption left the receiver incomplete.
	Incomplete
	// PartialAccurate (P): the receiver filled before all records fit.
	PartialAccurate
)

func (c Complete) String() string {
	return [...]string{"unknown", "complete", "incomplete", "partial"}[c]
}

// Status is the list status indicator.
type Status uint8

const (
	StatusPending     Status = iota // 0: building is pending
	StatusBuilding                  // 1: the list is being built
	StatusComplete                  // 2: completely built
	StatusError                     // 3: an error occurred while building
	StatusPrimed                    // 4: primed for asynchronous building
	StatusTooMuchData               // 5: incomplete; too much data
)

func (s Status) String() string {
	return [...]string{"pending", "building", "complete", "error", "primed", "too much data"}[s]
}

// Info is the decoded list information.
type Info struct {
	TotalRecords    int32
	RecordsReturned int32
	Handle          Handle
	RecordLength    int32
	Complete        Complete
	// Created is the raw 13-character creation timestamp (century digit,
	// YYMMDD, HHMMSS) in EBCDIC digits.
	Created     [13]byte
	Status      Status
	InfoLength  int32
	FirstRecord int32
}

// ErrMalformed is matched by every decoding failure.
var ErrMalformed = errors.New("openlist: malformed list information")

// Unmarshal decodes exactly Length bytes. The handle is copied without
// character conversion.
func Unmarshal(b []byte) (Info, error) {
	if len(b) != Length {
		return Info{}, fmt.Errorf("%w: %d bytes, want %d", ErrMalformed, len(b), Length)
	}
	i32 := func(off int) int32 { return int32(binary.BigEndian.Uint32(b[off:])) }
	info := Info{
		TotalRecords:    i32(0),
		RecordsReturned: i32(4),
		RecordLength:    i32(12),
		InfoLength:      i32(32),
		FirstRecord:     i32(36),
	}
	copy(info.Handle[:], b[8:12])
	copy(info.Created[:], b[17:30])
	for name, v := range map[string]int32{
		"total records": info.TotalRecords, "records returned": info.RecordsReturned,
		"record length": info.RecordLength, "length of information returned": info.InfoLength,
		"first record": info.FirstRecord,
	} {
		if v < 0 {
			return Info{}, fmt.Errorf("%w: negative %s %d", ErrMalformed, name, v)
		}
	}
	switch b[16] {
	case 0xC3: // EBCDIC C
		info.Complete = CompleteAccurate
	case 0xC9: // EBCDIC I
		info.Complete = Incomplete
	case 0xD7: // EBCDIC P
		info.Complete = PartialAccurate
	default:
		return Info{}, fmt.Errorf("%w: information-complete indicator 0x%02X", ErrMalformed, b[16])
	}
	if b[30] < 0xF0 || b[30] > 0xF5 {
		return Info{}, fmt.Errorf("%w: list status indicator 0x%02X", ErrMalformed, b[30])
	}
	info.Status = Status(b[30] - 0xF0)
	return info, nil
}

// RecoverHandle extracts the handle from a list-information buffer before
// full validation, so a caller can still close a list whose metadata is
// malformed. ok is false when the buffer is too short to hold it.
func RecoverHandle(b []byte) (h Handle, ok bool) {
	if len(b) < 12 {
		return Handle{}, false
	}
	copy(h[:], b[8:12])
	return h, true
}
