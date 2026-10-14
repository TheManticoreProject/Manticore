// Package structures holds the receiver record layouts of QGY/QGYOLAUS:
// AUTU0100, AUTU0150, AUTU0200 and AUTU0250.
//
// These layouts belong to QGYOLAUS only. Other programs reuse the same
// format names for different layouts (QSYS/QSYLAUTU's AUTU0200 is a
// 228-byte entry), so they live under this program's import path and are
// never selected by format name alone.
//
// Each decoder requires the exact record length before reading any field.
// Indicators accept only EBCDIC '0' and '1'. Layouts: IBM i "Open List of
// Authorized Users (QGYOLAUS) API",
// https://www.ibm.com/docs/en/i/7.6.0?topic=q-open-list-authorized-users-qgyolaus
// (retrieved 2026-10-14).
package structures

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
)

// Record lengths.
const (
	AUTU0100Length = 12
	AUTU0150Length = 62
	AUTU0200Length = 176
	AUTU0250Length = 228

	// MaxGroupProfiles is the size of the AUTU0200/AUTU0250 group array.
	MaxGroupProfiles = 16
)

// ErrMalformed is matched by every record decoding failure.
var ErrMalformed = errors.New("structures: malformed QGYOLAUS record")

// AUTU0100 is the profile name with its user/group indicators.
type AUTU0100 struct {
	Name string
	// GroupProfile is the user-or-group indicator ('1' = group profile).
	GroupProfile bool
	// GroupHasMembers is the group-members indicator.
	GroupHasMembers bool
}

// AUTU0150 adds the text description.
type AUTU0150 struct {
	AUTU0100
	Description string
}

// AUTU0200 adds the group profiles the user belongs to.
type AUTU0200 struct {
	AUTU0100
	GroupProfiles []string
}

// AUTU0250 combines AUTU0150 and AUTU0200.
type AUTU0250 struct {
	AUTU0100
	Description   string
	GroupProfiles []string
}

func checkLength(name string, b []byte, want int) error {
	if len(b) != want {
		return fmt.Errorf("%w: %s record of %d bytes, want %d", ErrMalformed, name, len(b), want)
	}
	return nil
}

func indicator(name string, b byte, offset int) (bool, error) {
	switch b {
	case 0xF0: // EBCDIC '0'
		return false, nil
	case 0xF1: // EBCDIC '1'
		return true, nil
	}
	return false, fmt.Errorf("%w: %s indicator at offset %d is 0x%02X", ErrMalformed, name, offset, b)
}

// decodeHead decodes offsets 0..11, shared by every format.
func decodeHead(name string, b []byte, codec ccsid.Codec) (AUTU0100, error) {
	var r AUTU0100
	var err error
	if r.Name, err = ccsid.DecodeFixed(codec, b[0:10]); err != nil {
		return AUTU0100{}, err
	}
	if r.GroupProfile, err = indicator(name, b[10], 10); err != nil {
		return AUTU0100{}, err
	}
	if r.GroupHasMembers, err = indicator(name, b[11], 11); err != nil {
		return AUTU0100{}, err
	}
	return r, nil
}

// decodeGroups decodes a BINARY(4) count at countOff followed by the
// 16-entry CHAR(10) array; entries past the count are ignored.
func decodeGroups(name string, b []byte, countOff int, codec ccsid.Codec) ([]string, error) {
	n := int32(binary.BigEndian.Uint32(b[countOff:]))
	if n < 0 || n > MaxGroupProfiles {
		return nil, fmt.Errorf("%w: %s group count %d outside 0..%d", ErrMalformed, name, n, MaxGroupProfiles)
	}
	groups := make([]string, 0, n)
	for i := 0; i < int(n); i++ {
		off := countOff + 4 + 10*i
		g, err := ccsid.DecodeFixed(codec, b[off:off+10])
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

// DecodeAUTU0100 decodes one 12-byte record.
func DecodeAUTU0100(b []byte, codec ccsid.Codec) (AUTU0100, error) {
	if err := checkLength("AUTU0100", b, AUTU0100Length); err != nil {
		return AUTU0100{}, err
	}
	return decodeHead("AUTU0100", b, codec)
}

// DecodeAUTU0150 decodes one 62-byte record: head, then CHAR(50) text
// description at offset 12.
func DecodeAUTU0150(b []byte, codec ccsid.Codec) (AUTU0150, error) {
	if err := checkLength("AUTU0150", b, AUTU0150Length); err != nil {
		return AUTU0150{}, err
	}
	head, err := decodeHead("AUTU0150", b, codec)
	if err != nil {
		return AUTU0150{}, err
	}
	desc, err := ccsid.DecodeFixed(codec, b[12:62])
	if err != nil {
		return AUTU0150{}, err
	}
	return AUTU0150{AUTU0100: head, Description: desc}, nil
}

// DecodeAUTU0200 decodes one 176-byte record: head, BINARY(4) group count at
// offset 12, ARRAY(16) OF CHAR(10) group profiles at offset 16.
func DecodeAUTU0200(b []byte, codec ccsid.Codec) (AUTU0200, error) {
	if err := checkLength("AUTU0200", b, AUTU0200Length); err != nil {
		return AUTU0200{}, err
	}
	head, err := decodeHead("AUTU0200", b, codec)
	if err != nil {
		return AUTU0200{}, err
	}
	groups, err := decodeGroups("AUTU0200", b, 12, codec)
	if err != nil {
		return AUTU0200{}, err
	}
	return AUTU0200{AUTU0100: head, GroupProfiles: groups}, nil
}

// DecodeAUTU0250 decodes one 228-byte record: head, CHAR(50) description at
// 12, two reserved bytes at 62, BINARY(4) group count at 64, ARRAY(16) OF
// CHAR(10) group profiles at 68.
func DecodeAUTU0250(b []byte, codec ccsid.Codec) (AUTU0250, error) {
	if err := checkLength("AUTU0250", b, AUTU0250Length); err != nil {
		return AUTU0250{}, err
	}
	head, err := decodeHead("AUTU0250", b, codec)
	if err != nil {
		return AUTU0250{}, err
	}
	desc, err := ccsid.DecodeFixed(codec, b[12:62])
	if err != nil {
		return AUTU0250{}, err
	}
	groups, err := decodeGroups("AUTU0250", b, 64, codec)
	if err != nil {
		return AUTU0250{}, err
	}
	return AUTU0250{AUTU0100: head, Description: desc, GroupProfiles: groups}, nil
}
