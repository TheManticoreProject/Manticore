package structures

// ENCRYPTED_LM_OWF_PASSWORD holds an encrypted LM (or NT) one-way function of a
// cleartext password ([MS-SAMR] 2.2.7.3). The IDL declares both the LM and NT
// names for this same 16-byte structure.
type ENCRYPTED_LM_OWF_PASSWORD struct {
	Data [16]byte
}

// ENCRYPTED_NT_OWF_PASSWORD is an alias for ENCRYPTED_LM_OWF_PASSWORD; the IDL
// gives the same 16-byte structure both names ([MS-SAMR] 2.2.7.3).
type ENCRYPTED_NT_OWF_PASSWORD = ENCRYPTED_LM_OWF_PASSWORD
