package pdu

import "strings"

// Flags1 is the first set of connectionless PDU flags (the flags1 field of the
// common header).
//
// References:
//   - [C706] section 12.6.3.2 (flags1):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
type Flags1 uint8

const (
	// Flags1Reserved01 is reserved for implementations; must be ignored.
	Flags1Reserved01 Flags1 = 0x01
	// Flags1LastFrag marks the last fragment of a multi-PDU transmission.
	Flags1LastFrag Flags1 = 0x02
	// Flags1Frag marks a fragment of a multi-PDU transmission.
	Flags1Frag Flags1 = 0x04
	// Flags1NoFack tells the receiver not to send a fack for this fragment.
	Flags1NoFack Flags1 = 0x08
	// Flags1Maybe marks a "maybe" request (no response expected). Client-to-server.
	Flags1Maybe Flags1 = 0x10
	// Flags1Idempotent marks an idempotent request. Client-to-server.
	Flags1Idempotent Flags1 = 0x20
	// Flags1Broadcast marks a broadcast request. Client-to-server.
	Flags1Broadcast Flags1 = 0x40
	// Flags1Reserved80 is reserved for implementations; must be ignored.
	Flags1Reserved80 Flags1 = 0x80
)

// Has reports whether all bits in flag are set.
func (f Flags1) Has(flag Flags1) bool { return f&flag == flag }

// String returns a "|"-joined list of set flag names, or "none".
func (f Flags1) String() string {
	if f == 0 {
		return "none"
	}
	names := []struct {
		flag Flags1
		name string
	}{
		{Flags1Reserved01, "reserved_01"},
		{Flags1LastFrag, "lastfrag"},
		{Flags1Frag, "frag"},
		{Flags1NoFack, "nofack"},
		{Flags1Maybe, "maybe"},
		{Flags1Idempotent, "idempotent"},
		{Flags1Broadcast, "broadcast"},
		{Flags1Reserved80, "reserved_80"},
	}
	var set []string
	for _, n := range names {
		if f.Has(n.flag) {
			set = append(set, n.name)
		}
	}
	return strings.Join(set, "|")
}

// Flags2 is the second set of connectionless PDU flags (the flags2 field of the
// common header). Only one bit is defined; the rest are reserved and must be zero.
//
// References:
//   - [C706] section 12.6.3.3 (flags2):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
type Flags2 uint8

const (
	// Flags2Reserved01 is reserved for implementations; must be ignored.
	Flags2Reserved01 Flags2 = 0x01
	// Flags2CancelPending indicates a cancel was pending at the time the call
	// completed.
	Flags2CancelPending Flags2 = 0x02
)

// Has reports whether all bits in flag are set.
func (f Flags2) Has(flag Flags2) bool { return f&flag == flag }

// String returns a "|"-joined list of set flag names, or "none".
func (f Flags2) String() string {
	if f == 0 {
		return "none"
	}
	var set []string
	if f.Has(Flags2Reserved01) {
		set = append(set, "reserved_01")
	}
	if f.Has(Flags2CancelPending) {
		set = append(set, "cancel_pending")
	}
	// Surface any reserved high bits so callers can spot unexpected values.
	if rest := f &^ (Flags2Reserved01 | Flags2CancelPending); rest != 0 {
		set = append(set, "reserved")
	}
	return strings.Join(set, "|")
}
