package signon

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/ibmi"
	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
)

// Info is the non-secret metadata learned during Sign-on. It never contains
// the profile, password, seeds or substitute.
type Info struct {
	// ServerVersion is the raw server VRM integer from CP 0x1101.
	ServerVersion uint32
	// VRM decodes ServerVersion as version<<16 | release<<8 | modification.
	// It is zero, with a warning, if the value does not decode.
	VRM ibmi.VRM
	// ServerLevel is the Sign-on data-stream level (not the Remote Command
	// level, which is negotiated separately).
	ServerLevel uint16
	// PasswordLevel is the raw CP 0x1119 byte; PasswordLevelReported is
	// false when the server omitted it and the DES family was assumed.
	PasswordLevel         byte
	PasswordLevelReported bool
	PasswordFamily        ibminet.PasswordFamily
	// SystemCCSID is CP 0x1114 from 0xF004. The Remote Command job CCSID is
	// a separate value.
	SystemCCSID uint32
	Warnings    []string
}

// decodeVRM interprets the raw server version.
func decodeVRM(raw uint32) (ibmi.VRM, error) {
	version := raw >> 16
	if version > 0xFF {
		return ibmi.VRM{}, fmt.Errorf("%w: version field %d", ibmi.ErrInvalidVRM, version)
	}
	return ibmi.NewVRM(uint8(version), uint8(raw>>8), uint8(raw))
}
