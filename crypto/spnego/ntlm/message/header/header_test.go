package header_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/header"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/types"
)

func TestHeaderMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name      string
		signature [8]byte
		msgType   types.MessageType
		wantError bool
		// wantUnmarshalError says the bytes must be refused when read back. A
		// header whose signature is not the NTLMSSP string is not an NTLM message,
		// so it marshals -- the field is whatever the caller put there -- and does
		// not parse.
		wantUnmarshalError bool
	}{
		{
			name:      "Standard NTLM Header",
			signature: header.NTLM_SIGNATURE,
			msgType:   types.MESSAGE_TYPE_NEGOTIATE,
			wantError: false,
		},
		{
			name:      "Standard NTLM Header, challenge",
			signature: header.NTLM_SIGNATURE,
			msgType:   types.MESSAGE_TYPE_CHALLENGE,
			wantError: false,
		},
		{
			name:               "Custom Signature",
			signature:          [8]byte{'T', 'E', 'S', 'T', 'S', 'I', 'G', 0},
			msgType:            types.MESSAGE_TYPE_CHALLENGE,
			wantError:          false,
			wantUnmarshalError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create original header
			original := &header.Header{
				Signature:   tt.signature,
				MessageType: tt.msgType,
			}

			// Marshal
			data, err := original.Marshal()
			if (err != nil) != tt.wantError {
				t.Errorf("Marshal() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if tt.wantError {
				return
			}

			// Unmarshal into new header
			unmarshalled := &header.Header{}
			_, err = unmarshalled.Unmarshal(data)
			if tt.wantUnmarshalError {
				if err == nil {
					t.Error("Unmarshal() accepted a header whose signature is not NTLMSSP")
				}
				return
			}
			if err != nil {
				t.Errorf("Unmarshal() error = %v", err)
				return
			}

			// Compare values
			if original.GetSignature() != unmarshalled.GetSignature() {
				t.Errorf("Signature mismatch after marshal/unmarshal: got %v, want %v",
					unmarshalled.GetSignature(), original.GetSignature())
			}

			if original.GetType() != unmarshalled.GetType() {
				t.Errorf("MessageType mismatch after marshal/unmarshal: got %v, want %v",
					unmarshalled.GetType(), original.GetType())
			}
		})
	}
}

// TestUnmarshalRejectsForeignBytes asserts a buffer that is not an NTLM message
// is refused rather than parsed into whatever the bytes happen to say.
func TestUnmarshalRejectsForeignBytes(t *testing.T) {
	foreign := map[string][]byte{
		"ASCII digits":      []byte("000000000000"),
		"zeroes":            make([]byte, 12),
		"truncated NTLMSSP": append([]byte("NTLMSS"), 0x00, 0x00, 0x01, 0x00, 0x00, 0x00),
	}

	for name, data := range foreign {
		t.Run(name, func(t *testing.T) {
			parsed := &header.Header{}
			if _, err := parsed.Unmarshal(data); err == nil {
				t.Error("Unmarshal() accepted a buffer that is not an NTLM message")
			}
		})
	}
}

// TestExpectRejectsTheWrongMessageType asserts the type check distinguishes the
// three messages, which share no other discriminator.
func TestExpectRejectsTheWrongMessageType(t *testing.T) {
	parsed := &header.Header{Signature: header.NTLM_SIGNATURE, MessageType: types.MESSAGE_TYPE_CHALLENGE}

	if err := parsed.Expect(types.MESSAGE_TYPE_CHALLENGE); err != nil {
		t.Errorf("Expect() rejected the matching type: %v", err)
	}
	for _, other := range []types.MessageType{types.MESSAGE_TYPE_NEGOTIATE, types.MESSAGE_TYPE_AUTHENTICATE} {
		if err := parsed.Expect(other); err == nil {
			t.Errorf("Expect(%s) accepted a %s", other, parsed.MessageType)
		}
	}
}
