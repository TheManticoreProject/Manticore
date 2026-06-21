package structures

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// DRS_EXTENSIONS_INT is the capability structure carried by DRS_EXTENSIONS at
// IDL_DRSBind ([MS-DRSR] 5.39). It is NOT an NDR type: it is a fixed little-endian byte
// layout, (un)marshalled by hand. Both peers send it to negotiate which request/reply
// message versions and capabilities they support.
//
// Wire relationship (verified against a live DC): DRS_EXTENSIONS is { DWORD cb; BYTE
// rgb[cb]; }, and the DRS_EXTENSIONS_INT.cb field IS that same DRS_EXTENSIONS.cb — it is
// NOT repeated inside rgb. So rgb holds only dwFlags..dwExtCaps (52 bytes) and Cb counts
// exactly those bytes. (An earlier revision wrongly prepended cb to rgb, which shifted
// dwFlags and made the server miss the client's STRONG_ENCRYPTION flag.) Cb here is
// informational; Marshal/ToExtensions derive it from the field block length.
//
// GUID fields are raw 16-octet RPC-wire values (guid.GUID.ToBytes order); a non-DC
// client leaves them zero.
type DRS_EXTENSIONS_INT struct {
	Cb            uint32
	DwFlags       uint32
	SiteObjGuid   [16]byte
	Pid           int32
	DwReplEpoch   uint32
	DwFlagsExt    uint32
	ConfigObjGUID [16]byte
	DwExtCaps     uint32
}

// extIntFieldsSize is the byte count of dwFlags..dwExtCaps, i.e. the length of rgb and
// the value of cb: 4 + 16 + 4 + 4 + 4 + 16 + 4.
const extIntFieldsSize = 52

// dwFlags capability bits ([MS-DRSR] 5.39).
const (
	DRS_EXT_BASE                     uint32 = 0x00000001
	DRS_EXT_ASYNCREPL                uint32 = 0x00000002
	DRS_EXT_REMOVEAPI                uint32 = 0x00000004
	DRS_EXT_MOVEREQ_V2               uint32 = 0x00000008
	DRS_EXT_GETCHG_DEFLATE           uint32 = 0x00000010
	DRS_EXT_DCINFO_V1                uint32 = 0x00000020
	DRS_EXT_RESTORE_USN_OPTIMIZATION uint32 = 0x00000040
	DRS_EXT_ADDENTRY                 uint32 = 0x00000080
	DRS_EXT_KCC_EXECUTE              uint32 = 0x00000100
	DRS_EXT_ADDENTRY_V2              uint32 = 0x00000200
	DRS_EXT_LINKED_VALUE_REPLICATION uint32 = 0x00000400
	DRS_EXT_DCINFO_V2                uint32 = 0x00000800
	DRS_EXT_INSTANCE_TYPE_NOT_REQ    uint32 = 0x00001000
	DRS_EXT_CRYPTO_BIND              uint32 = 0x00002000
	DRS_EXT_GET_REPL_INFO            uint32 = 0x00004000
	DRS_EXT_STRONG_ENCRYPTION        uint32 = 0x00008000
	DRS_EXT_DCINFO_VFFFFFFFF         uint32 = 0x00010000
	DRS_EXT_TRANSITIVE_MEMBERSHIP    uint32 = 0x00020000
	DRS_EXT_ADD_SID_HISTORY          uint32 = 0x00040000
	DRS_EXT_POST_BETA3               uint32 = 0x00080000
	DRS_EXT_GETCHGREQ_V5             uint32 = 0x00100000
	DRS_EXT_GETMEMBERSHIPS2          uint32 = 0x00200000
	DRS_EXT_GETCHGREQ_V6             uint32 = 0x00400000
	DRS_EXT_NONDOMAIN_NCS            uint32 = 0x00800000
	DRS_EXT_GETCHGREQ_V8             uint32 = 0x01000000
	DRS_EXT_GETCHGREPLY_V5           uint32 = 0x02000000
	DRS_EXT_GETCHGREPLY_V6           uint32 = 0x04000000
	DRS_EXT_WHISTLER_BETA3           uint32 = 0x08000000
	DRS_EXT_W2K3_DEFLATE             uint32 = 0x10000000
	DRS_EXT_GETCHGREQ_V10            uint32 = 0x20000000
)

// dwFlagsExt capability bits ([MS-DRSR] 5.39).
const (
	DRS_EXT_ADAM                uint32 = 0x00000001
	DRS_EXT_LH_BETA2            uint32 = 0x00000002
	DRS_EXT_RECYCLE_BIN         uint32 = 0x00000004
	DRS_EXT_GETCHGREPLY_V9      uint32 = 0x00000100
	DRS_EXT_RPC_CORRELATIONID_1 uint32 = 0x00000400
)

// NTDSAPIClientGUID is the well-known client DSA GUID a non-DC tool passes as
// IDL_DRSBind's puuidClientDsa (the value the NTDSAPI DsBind* client uses):
// e24d201a-4fd6-11d1-a3da-0000f875ae0d. The server rejects only the NULL GUID; this
// value is also the one required if the handle is later used for IDL_DRSWriteSPN
// ([MS-DRSR] 4.1.3.1).
func NTDSAPIClientGUID() UUID {
	g, err := guid.FromFormatD("e24d201a-4fd6-11d1-a3da-0000f875ae0d")
	if err != nil {
		panic(fmt.Sprintf("drsuapi: bad NTDSAPI client GUID literal: %v", err))
	}
	return UUIDFromGUID(*g)
}

// DefaultClientExtensions returns the DRS_EXTENSIONS_INT a replication client (DCSync)
// sends at bind. The flags ask the server for the message versions IDL_DRSGetNCChanges
// needs and assert support for on-wire secret encryption:
//
//   - GETCHGREQ_V8 + RESTORE_USN_OPTIMIZATION — issue V8 change requests ([MS-DRSR] 4.1.3.2)
//   - GETCHGREPLY_V6 — receive the V6 reply DCSync parses ([MS-DRSR] 4.1.3.1)
//   - STRONG_ENCRYPTION — required for the DC to send secrets over the wire
//   - GETCHGREQ_V6, NONDOMAIN_NCS — match the widely-deployed impacket client set
//
// dwExtCaps is set to all-ones (impacket's choice) to assert every extended capability.
// SiteObjGuid/ConfigObjGUID/Pid/dwReplEpoch stay zero, as a non-DC client has none.
func DefaultClientExtensions() *DRS_EXTENSIONS_INT {
	return &DRS_EXTENSIONS_INT{
		Cb: extIntFieldsSize,
		DwFlags: DRS_EXT_GETCHGREQ_V6 | DRS_EXT_GETCHGREPLY_V6 | DRS_EXT_GETCHGREQ_V8 |
			DRS_EXT_RESTORE_USN_OPTIMIZATION | DRS_EXT_STRONG_ENCRYPTION | DRS_EXT_NONDOMAIN_NCS,
		DwExtCaps: 0xFFFFFFFF,
	}
}

// Marshal serializes the field block dwFlags..dwExtCaps to its [MS-DRSR] 5.39
// little-endian wire bytes (52 octets). This is the rgb content of DRS_EXTENSIONS; the
// cb that precedes rgb is the DRS_EXTENSIONS.Cb field, not part of this block.
func (e *DRS_EXTENSIONS_INT) Marshal() []byte {
	b := make([]byte, extIntFieldsSize)
	binary.LittleEndian.PutUint32(b[0:4], e.DwFlags)
	copy(b[4:20], e.SiteObjGuid[:])
	binary.LittleEndian.PutUint32(b[20:24], uint32(e.Pid))
	binary.LittleEndian.PutUint32(b[24:28], e.DwReplEpoch)
	binary.LittleEndian.PutUint32(b[28:32], e.DwFlagsExt)
	copy(b[32:48], e.ConfigObjGUID[:])
	binary.LittleEndian.PutUint32(b[48:52], e.DwExtCaps)
	return b
}

// ToExtensions wraps the marshalled field block in a DRS_EXTENSIONS, setting Cb to the
// block length (the value the [size_is(cb)] rgb array is sized by).
func (e *DRS_EXTENSIONS_INT) ToExtensions() DRS_EXTENSIONS {
	rgb := e.Marshal()
	return DRS_EXTENSIONS{Cb: uint32(len(rgb)), Rgb: rgb}
}

// ParseExtensionsInt decodes a DRS_EXTENSIONS_INT from a DRS_EXTENSIONS rgb block (which
// begins at dwFlags, not at cb). Trailing fields a shorter peer omitted are left zero
// (the structure is forward-extensible). Cb is set to the block length.
func ParseExtensionsInt(b []byte) (*DRS_EXTENSIONS_INT, error) {
	if len(b) < 4 {
		return nil, fmt.Errorf("drsuapi: DRS_EXTENSIONS_INT too short: %d bytes", len(b))
	}
	var e DRS_EXTENSIONS_INT
	e.Cb = uint32(len(b))
	get := func(off int) (uint32, bool) {
		if off+4 > len(b) {
			return 0, false
		}
		return binary.LittleEndian.Uint32(b[off : off+4]), true
	}
	e.DwFlags, _ = get(0)
	if len(b) >= 20 {
		copy(e.SiteObjGuid[:], b[4:20])
	}
	if v, ok := get(20); ok {
		e.Pid = int32(v)
	}
	if v, ok := get(24); ok {
		e.DwReplEpoch = v
	}
	if v, ok := get(28); ok {
		e.DwFlagsExt = v
	}
	if len(b) >= 48 {
		copy(e.ConfigObjGUID[:], b[32:48])
	}
	if v, ok := get(48); ok {
		e.DwExtCaps = v
	}
	return &e, nil
}

// ParseInt decodes the DRS_EXTENSIONS_INT carried in a DRS_EXTENSIONS blob (e.g. the
// server's negotiated extensions returned by IDL_DRSBind).
func (x DRS_EXTENSIONS) ParseInt() (*DRS_EXTENSIONS_INT, error) {
	return ParseExtensionsInt(x.Rgb)
}
