// Package kirbi reads and writes ".kirbi" credential files. A .kirbi file is a
// DER-encoded Kerberos KRB-CRED message (RFC 4120 Section 5.8.1) carrying one
// or more tickets — the interchange format used by Rubeus and (via -k) Impacket
// for pass-the-ticket. By convention the EncKrbCredPart is stored unencrypted
// with etype 0, so the session key and ticket metadata are in the clear.
package kirbi

import (
	"fmt"
	"os"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// Bytes returns the DER encoding of a KRB-CRED (the raw .kirbi contents).
func Bytes(cred *messages.KRBCred) ([]byte, error) {
	return cred.Marshal()
}

// Parse decodes .kirbi bytes into a KRB-CRED.
func Parse(data []byte) (*messages.KRBCred, error) {
	var c messages.KRBCred
	if _, err := c.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("kirbi: parse: %w", err)
	}
	return &c, nil
}

// Save writes cred to path in .kirbi (DER KRB-CRED) form, 0600.
func Save(path string, cred *messages.KRBCred) error {
	b, err := Bytes(cred)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return fmt.Errorf("kirbi: write %s: %w", path, err)
	}
	return nil
}

// Load reads and parses a .kirbi file.
func Load(path string) (*messages.KRBCred, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("kirbi: read %s: %w", path, err)
	}
	return Parse(b)
}

// New builds a KRB-CRED wrapping a single ticket, with the EncKrbCredPart stored
// unencrypted (etype 0) as .kirbi conventionally does. ticketRaw is the raw
// APPLICATION[1] ticket TLV as issued by the KDC; info carries the session key,
// principal names, flags, and times for that ticket.
func New(ticketRaw []byte, info messages.KrbCredInfo) (*messages.KRBCred, error) {
	enc := messages.EncKrbCredPart{TicketInfo: []messages.KrbCredInfo{info}}
	encBytes, err := enc.Marshal()
	if err != nil {
		return nil, fmt.Errorf("kirbi: marshal enc-part: %w", err)
	}
	return &messages.KRBCred{
		PVNO:       messages.KerberosV5,
		MsgType:    messages.MsgTypeKRBCred,
		TicketsRaw: [][]byte{ticketRaw},
		EncPart:    messages.EncryptedData{EType: 0, Cipher: encBytes},
	}, nil
}

// TicketInfo decodes and returns the EncKrbCredPart ticket-info entries from a
// KRB-CRED whose enc-part is stored unencrypted (etype 0), the .kirbi
// convention. It errors if the enc-part is encrypted (etype != 0).
func TicketInfo(cred *messages.KRBCred) ([]messages.KrbCredInfo, error) {
	if cred.EncPart.EType != 0 {
		return nil, fmt.Errorf("kirbi: enc-part is encrypted (etype %d); cannot read without a key", cred.EncPart.EType)
	}
	var enc messages.EncKrbCredPart
	if _, err := enc.Unmarshal(cred.EncPart.Cipher); err != nil {
		return nil, fmt.Errorf("kirbi: parse enc-part: %w", err)
	}
	return enc.TicketInfo, nil
}
