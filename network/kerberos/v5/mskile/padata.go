// Package mskile implements the Microsoft-specific pre-authentication data
// (PA-DATA) payloads that MS-KILE layers onto RFC 4120's extension point:
// PA-PAC-REQUEST (128), PA-PAC-OPTIONS (167), and PA-SUPPORTED-ENCTYPES (165).
// These are additive extensions — the surrounding AS-REQ/TGS-REQ messages
// remain exactly RFC 4120's.
package mskile

import (
	"encoding/asn1"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// ── PA-PAC-REQUEST (padata-type 128, MS-KILE 2.2.3) ───────────────────────────
//
//	KERB-PA-PAC-REQUEST ::= SEQUENCE { include-pac [0] BOOLEAN }
//
// include-pac = TRUE asks the KDC to include a PAC; FALSE asks it to omit one
// (used for Kerberoasting to get shorter, PAC-free service tickets).

type pacRequest struct {
	IncludePAC bool `asn1:"explicit,tag:0"`
}

// PACRequest returns the DER-encoded PA-PAC-REQUEST padata-value.
func PACRequest(includePAC bool) ([]byte, error) {
	return asn1.Marshal(pacRequest{IncludePAC: includePAC})
}

// ParsePACRequest decodes a PA-PAC-REQUEST padata-value.
func ParsePACRequest(b []byte) (includePAC bool, err error) {
	var p pacRequest
	if _, err := asn1.Unmarshal(b, &p); err != nil {
		return false, fmt.Errorf("mskile: parse PA-PAC-REQUEST: %w", err)
	}
	return p.IncludePAC, nil
}

// PACRequestPAData builds the PA-DATA element for PA-PAC-REQUEST.
func PACRequestPAData(includePAC bool) (messages.PAData, error) {
	v, err := PACRequest(includePAC)
	if err != nil {
		return messages.PAData{}, err
	}
	return messages.PAData{PADataType: iana.PAPACRequest, PADataValue: v}, nil
}

// ── PA-PAC-OPTIONS (padata-type 167, MS-KILE 2.2.10 / MS-SFU 2.2.5) ────────────
//
//	PA-PAC-OPTIONS ::= SEQUENCE { flags [0] KerberosFlags }
//
// The flags field is a 32-bit KerberosFlags bit string (bit 0 = MSB).

// PA-PAC-OPTIONS flag bit positions.
const (
	PACOptionClaims                        = 0 // Claims
	PACOptionBranchAware                   = 1 // Branch Aware
	PACOptionForwardToFullDC               = 2 // Forward to Full DC
	PACOptionResourceBasedConstrainedDeleg = 3 // resource-based constrained delegation (MS-SFU)
)

type pacOptions struct {
	Flags asn1.BitString `asn1:"explicit,tag:0"`
}

// PACOptions returns the DER-encoded PA-PAC-OPTIONS padata-value with the given
// option bits set.
func PACOptions(bits ...int) ([]byte, error) {
	return asn1.Marshal(pacOptions{Flags: messages.NewKerberosFlags(bits...)})
}

// ParsePACOptions decodes a PA-PAC-OPTIONS padata-value and reports which option
// bit positions are set.
func ParsePACOptions(b []byte) ([]int, error) {
	var p pacOptions
	if _, err := asn1.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("mskile: parse PA-PAC-OPTIONS: %w", err)
	}
	var set []int
	for i := 0; i < p.Flags.BitLength; i++ {
		if p.Flags.At(i) == 1 {
			set = append(set, i)
		}
	}
	return set, nil
}

// PACOptionsPAData builds the PA-DATA element for PA-PAC-OPTIONS.
func PACOptionsPAData(bits ...int) (messages.PAData, error) {
	v, err := PACOptions(bits...)
	if err != nil {
		return messages.PAData{}, err
	}
	return messages.PAData{PADataType: iana.PAPACOptions, PADataValue: v}, nil
}

// ── PA-SUPPORTED-ENCTYPES (padata-type 165, MS-KILE 2.2.8) ─────────────────────
//
//	PA-SUPPORTED-ENCTYPES ::= Int32  -- a Supported Encryption Types bit field
//
// The padata-value is a DER-encoded INTEGER (NOT a raw little-endian word) whose
// bits are the msDS-SupportedEncryptionTypes flags of MS-KILE 2.2.7.

// Supported Encryption Types bit flags (MS-KILE 2.2.7).
const (
	EncTypeDESCBCCRC               = 0x00000001 // A
	EncTypeDESCBCMD5               = 0x00000002 // B
	EncTypeRC4HMAC                 = 0x00000004 // C
	EncTypeAES128CTSHMACSHA196     = 0x00000008 // D
	EncTypeAES256CTSHMACSHA196     = 0x00000010 // E
	EncTypeFASTSupported           = 0x00000020 // F
	EncTypeCompoundIdentity        = 0x00000040 // G
	EncTypeClaimsSupported         = 0x00000080 // H
	EncTypeResourceSIDCompDisabled = 0x00000100 // I
	EncTypeAES256SK                = 0x00000200 // J: enforce AES session keys
	EncTypeAES128CTSHMACSHA256     = 0x00000400 // K (RFC 8009)
	EncTypeAES256CTSHMACSHA384     = 0x00000800 // L (RFC 8009)
)

// SupportedEnctypes returns the DER-encoded PA-SUPPORTED-ENCTYPES padata-value.
func SupportedEnctypes(flags int32) ([]byte, error) {
	return asn1.Marshal(int(flags))
}

// ParseSupportedEnctypes decodes a PA-SUPPORTED-ENCTYPES padata-value.
func ParseSupportedEnctypes(b []byte) (int32, error) {
	var v int
	if _, err := asn1.Unmarshal(b, &v); err != nil {
		return 0, fmt.Errorf("mskile: parse PA-SUPPORTED-ENCTYPES: %w", err)
	}
	return int32(v), nil
}

// SupportedEnctypesPAData builds the PA-DATA element for PA-SUPPORTED-ENCTYPES.
func SupportedEnctypesPAData(flags int32) (messages.PAData, error) {
	v, err := SupportedEnctypes(flags)
	if err != nil {
		return messages.PAData{}, err
	}
	return messages.PAData{PADataType: iana.PASupportedEnctypes, PADataValue: v}, nil
}
