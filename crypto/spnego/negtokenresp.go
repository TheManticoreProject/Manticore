package spnego

import (
	"encoding/asn1"
	"fmt"
)

// NegState represents the negotiation state in SPNEGO
type NegState asn1.Enumerated

const (
	// NegStateAcceptCompleted indicates the negotiation is complete and successful
	NegStateAcceptCompleted NegState = 0
	// NegStateAcceptIncomplete indicates more negotiation messages are needed
	NegStateAcceptIncomplete NegState = 1
	// NegStateReject indicates the negotiation has failed
	NegStateReject NegState = 2
	// NegStateRequestMIC indicates a MIC token is requested
	NegStateRequestMIC NegState = 3
)

// String returns a string representation of the NegState
// Parameters:
//   - n: The NegState to convert to a string
//
// Returns:
//   - string: A string representation of the NegState
func (n NegState) String() string {
	switch n {
	case NegStateAcceptCompleted:
		return fmt.Sprintf("Accept Completed (%d)", n)
	case NegStateAcceptIncomplete:
		return fmt.Sprintf("Accept Incomplete (%d)", n)
	case NegStateReject:
		return fmt.Sprintf("Reject (%d)", n)
	case NegStateRequestMIC:
		return fmt.Sprintf("Request MIC (%d)", n)
	default:
		return fmt.Sprintf("Unknown (%d)", n)
	}
}

// NegTokenResp is the response token sent by the server
/*
	NegTokenResp ::= SEQUENCE {
	    negState[0]       ENUMERATED {
	        accept-completed(0),
	        accept-incomplete(1),
	        reject(2),
	        request-mic(3)
	    } OPTIONAL,
	    supportedMech[1]  MechType OPTIONAL,
	    responseToken[2]  OCTET STRING OPTIONAL,
	    mechListMIC[3]    OCTET STRING OPTIONAL
	}
*/
type NegTokenResp struct {
	NegState      NegState              `asn1:"optional,tag:0"`
	SupportedMech asn1.ObjectIdentifier `asn1:"optional,tag:1"`
	ResponseToken []byte                `asn1:"optional,tag:2,octet"`
	MechListMIC   []byte                `asn1:"optional,tag:3,octet"`

	// SuppressNegState omits the [0] negState field on Marshal. A client's
	// continuation token (the AUTHENTICATE) carries only the responseToken — no
	// negState and no supportedMech — matching the Windows client.
	SuppressNegState bool `asn1:"-"`
}

// NewNegTokenResp creates a new NegTokenResp
// Parameters:
//   - state: The negotiation state
//   - mech: The supported mechanism OID
//   - responseToken: The response token bytes
//
// Returns:
//   - *NegTokenResp: A new NegTokenResp instance
func NewNegTokenResp(state NegState, mech asn1.ObjectIdentifier, responseToken []byte) *NegTokenResp {
	return &NegTokenResp{
		NegState:      state,
		SupportedMech: mech,
		ResponseToken: responseToken,
	}
}

// SetMechToken sets the mech token in the NegTokenResp
// Parameters:
//   - responseToken: The response token bytes to set
func (n *NegTokenResp) SetMechToken(responseToken []byte) {
	n.ResponseToken = responseToken
}

// SetMechTokenNTLM sets the NTLM token in the NegTokenResp
// Parameters:
//   - responseToken: The NTLM response token bytes to set
func (n *NegTokenResp) SetMechTokenNTLM(responseToken []byte) {
	n.SupportedMech = NtlmOID
	n.ResponseToken = responseToken
}

// Marshal marshals the NegTokenResp into a byte slice
// Returns:
//   - []byte: The marshaled bytes
//   - error: An error if marshaling fails
func (n *NegTokenResp) Marshal() ([]byte, error) {
	var seq []asn1.RawValue

	// [0] EXPLICIT ENUMERATED
	if !n.SuppressNegState {
		enumBytes, err := asn1.Marshal(asn1.Enumerated(n.NegState))
		if err != nil {
			return nil, fmt.Errorf("marshal negState: %v", err)
		}
		seq = append(seq, asn1.RawValue{
			Class:      asn1.ClassContextSpecific,
			Tag:        0,
			IsCompound: true,
			Bytes:      enumBytes,
		})
	}

	// [1] EXPLICIT OBJECT IDENTIFIER
	if len(n.SupportedMech) > 0 {
		oidBytes, err := asn1.Marshal(n.SupportedMech)
		if err != nil {
			return nil, fmt.Errorf("marshal SupportedMech: %v", err)
		}
		seq = append(seq, asn1.RawValue{
			Class:      asn1.ClassContextSpecific,
			Tag:        1,
			IsCompound: true,
			Bytes:      oidBytes,
		})
	}

	// [2] EXPLICIT OCTET STRING
	if len(n.ResponseToken) > 0 {
		tokenBytes, err := asn1.Marshal(n.ResponseToken)
		if err != nil {
			return nil, fmt.Errorf("marshal ResponseToken: %v", err)
		}
		seq = append(seq, asn1.RawValue{
			Class:      asn1.ClassContextSpecific,
			Tag:        2,
			IsCompound: true,
			Bytes:      tokenBytes,
		})
	}

	// [3] EXPLICIT OCTET STRING (optional)
	if len(n.MechListMIC) > 0 {
		micBytes, err := asn1.Marshal(n.MechListMIC)
		if err != nil {
			return nil, fmt.Errorf("marshal MechListMIC: %v", err)
		}
		seq = append(seq, asn1.RawValue{
			Class:      asn1.ClassContextSpecific,
			Tag:        3,
			IsCompound: true,
			Bytes:      micBytes,
		})
	}

	// Marshal the whole SEQUENCE
	var seqBytes []byte
	for _, field := range seq {
		b, err := asn1.Marshal(field)
		if err != nil {
			return nil, fmt.Errorf("marshal field with tag [%d]: %v", field.Tag, err)
		}
		seqBytes = append(seqBytes, b...)
	}

	// Wrap as SEQUENCE
	final, err := asn1.Marshal(asn1.RawValue{
		Class:      asn1.ClassUniversal,
		Tag:        asn1.TagSequence,
		IsCompound: true,
		Bytes:      seqBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal final SEQUENCE: %v", err)
	}

	return final, nil
}

// Unmarshal unmarshals the NegTokenResp from a byte slice
// Parameters:
//   - data: The bytes to unmarshal
//
// Returns:
//   - int: The number of bytes read
//   - error: An error if unmarshaling fails
func (n *NegTokenResp) Unmarshal(data []byte) (int, error) {
	var outer asn1.RawValue
	rest, err := asn1.Unmarshal(data, &outer)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal outer SEQUENCE: %v", err)
	}
	if outer.Class != asn1.ClassUniversal || outer.Tag != asn1.TagSequence {
		return 0, fmt.Errorf("expected universal SEQUENCE, got class=%d tag=%d", outer.Class, outer.Tag)
	}

	inner := outer.Bytes
	for len(inner) > 0 {
		var field asn1.RawValue
		innerRest, err := asn1.Unmarshal(inner, &field)
		if err != nil {
			return 0, fmt.Errorf("failed to unmarshal inner field: %v", err)
		}

		if field.Class == asn1.ClassContextSpecific {
			switch field.Tag {
			case 0:
				// ENUMERATED inside EXPLICIT
				var enumRaw asn1.RawValue
				_, err := asn1.Unmarshal(field.Bytes, &enumRaw)
				if err != nil {
					return 0, fmt.Errorf("failed to decode negState: %v", err)
				}
				if enumRaw.Tag != asn1.TagEnum {
					return 0, fmt.Errorf("negState is not an ENUMERATED type")
				}
				if len(enumRaw.Bytes) == 0 {
					return 0, fmt.Errorf("negState ENUMERATED has no content")
				}
				n.NegState = NegState(enumRaw.Bytes[0])
			case 1:
				// OBJECT IDENTIFIER inside EXPLICIT
				var oidRaw asn1.RawValue
				_, err := asn1.Unmarshal(field.Bytes, &oidRaw)
				if err != nil {
					return 0, fmt.Errorf("failed to decode supportedMech: %v", err)
				}
				var oid asn1.ObjectIdentifier
				_, err = asn1.Unmarshal(field.Bytes, &oid)
				if err != nil {
					return 0, fmt.Errorf("failed to parse supportedMech OID: %v", err)
				}
				n.SupportedMech = oid
			case 2:
				// OCTET STRING inside EXPLICIT
				var token []byte
				_, err := asn1.Unmarshal(field.Bytes, &token)
				if err != nil {
					return 0, fmt.Errorf("failed to decode responseToken: %v", err)
				}
				n.ResponseToken = token
			case 3:
				// OCTET STRING inside EXPLICIT
				var mic []byte
				_, err := asn1.Unmarshal(field.Bytes, &mic)
				if err != nil {
					return 0, fmt.Errorf("failed to decode mechListMIC: %v", err)
				}
				n.MechListMIC = mic
			default:
				// unknown context-specific tag; skip
			}
		}

		inner = innerRest
	}

	bytesRead := len(data) - len(rest)
	return bytesRead, nil
}

// CreateNegTokenResp creates an ASN.1 encoded SPNEGO NegTokenResp
// Parameters:
//   - state: The negotiation state
//   - mech: The supported mechanism OID
//   - token: The response token bytes
//
// Returns:
//   - []byte: The encoded SPNEGO token
//   - error: An error if token creation fails
func CreateNegTokenResp(state NegState, mech asn1.ObjectIdentifier, token []byte) ([]byte, error) {
	resp := NegTokenResp{
		NegState:      state,
		SupportedMech: mech,
		ResponseToken: token,
	}

	respBytes, err := asn1.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal NegTokenResp: %v", err)
	}
	return wrapSPNEGO(respBytes)
}
