package spnego

import (
	"encoding/asn1"
	"fmt"
)

// NegTokenInit is the initial SPNEGO token sent by the client
// wrapped in a [0] EXPLICIT context tag.
type NegTokenInit struct {
	MechTypes    []asn1.ObjectIdentifier `asn1:"explicit,tag:0"`
	ReqFlags     asn1.BitString          `asn1:"explicit,optional,tag:1"`
	MechToken    []byte                  `asn1:"explicit,optional,tag:2"`
	MechTokenMIC []byte                  `asn1:"explicit,optional,tag:3"`
}

// MarshalMechTypeList returns the DER encoding of a MechTypeList: the bare
// SEQUENCE OF MechType, without the [0] context tag that wraps it inside a
// NegTokenInit.
//
// This is the exact input a mechListMIC is computed over. RFC 4178 section 5 is
// explicit that the message is "the DER encoding of the value of type
// MechTypeList, which is contained in the mechTypes field of the NegTokenInit",
// and NOT the DER encoding of the type "[0] MechTypeList" — so the context tag
// must be absent or the MIC will not verify against a peer's.
func MarshalMechTypeList(mechTypes []asn1.ObjectIdentifier) ([]byte, error) {
	der, err := asn1.Marshal(mechTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MechTypeList: %v", err)
	}
	return der, nil
}

// CreateNegTokenInit creates a SPNEGO NegTokenInit with the given NTLM token and marshals it.
// Parameters:
//   - ntlmToken: The NTLM token bytes to include in the SPNEGO token
//
// Returns:
//   - []byte: The marshaled SPNEGO token containing the NTLM token
//   - error: An error if token creation fails
func CreateNegTokenInit(ntlmToken []byte) ([]byte, error) {
	init := NegTokenInit{
		MechTypes: []asn1.ObjectIdentifier{NtlmOID},
		MechToken: ntlmToken,
	}
	return init.Marshal()
}

// CreateNegTokenInitKerberos creates a SPNEGO NegTokenInit advertising the
// Kerberos mechanism and carrying the given Kerberos GSS-API token (a
// KRB_AP_REQ) as the mechToken.
func CreateNegTokenInitKerberos(kerberosToken []byte) ([]byte, error) {
	init := NegTokenInit{
		MechTypes: []asn1.ObjectIdentifier{KerberosOID},
		MechToken: kerberosToken,
	}
	return init.Marshal()
}

// NewNegTokenInit creates a new NegTokenInit with the specified parameters
// Parameters:
//   - mechTypes: The mechanism type identifiers to include
//   - reqFlags: The requested flags for the token
//   - mechToken: The mechanism token bytes
//   - mechTokenMIC: The mechanism token MIC bytes
//
// Returns:
//   - *NegTokenInit: A new NegTokenInit initialized with the provided parameters
func NewNegTokenInit(mechTypes []asn1.ObjectIdentifier, reqFlags asn1.BitString, mechToken []byte, mechTokenMIC []byte) *NegTokenInit {
	return &NegTokenInit{
		MechTypes:    mechTypes,
		ReqFlags:     reqFlags,
		MechToken:    mechToken,
		MechTokenMIC: mechTokenMIC,
	}
}

// SetMechTokenNTLM sets the NTLM token in the NegTokenInit
// Parameters:
//   - mechToken: The NTLM mechanism token bytes to set
func (n *NegTokenInit) SetMechTokenNTLM(mechToken []byte) {
	n.MechTypes = []asn1.ObjectIdentifier{NtlmOID}
	n.MechToken = mechToken
}

// SetMechToken sets the mech token in the NegTokenInit
// Parameters:
//   - mechToken: The mechanism token bytes to set
func (n *NegTokenInit) SetMechToken(mechToken []byte) {
	n.MechToken = mechToken
}

// Marshal marshals the NegTokenInit into a byte slice
// Returns:
//   - []byte: The marshaled SPNEGO token bytes
//   - error: An error if marshaling fails
func (n *NegTokenInit) Marshal() ([]byte, error) {
	marshalled, err := asn1.Marshal(*n)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal NegTokenInit: %v", err)
	}
	// Wrap and return
	return wrapSPNEGO(marshalled)
}

// Unmarshal unmarshals the NegTokenInit from a byte slice
// Parameters:
//   - data: The bytes to unmarshal from
//
// Returns:
//   - error: An error if unmarshaling fails
func (n *NegTokenInit) Unmarshal(data []byte) error {
	_, err := asn1.Unmarshal(data, n)
	return err
}
