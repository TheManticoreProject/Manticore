package mswkst

// JOINPR_ENCRYPTED_USER_PASSWORD is the RC4-encrypted join password ([MS-WKST] 2.2.5.18).
// Buffer is a FIXED array of JOIN_OBFUSCATOR_LENGTH + JOIN_MAX_PASSWORD_LENGTH*sizeof(wchar_t)
// + sizeof(unsigned long) = 8 + 256*2 + 4 = 524 bytes (no conformance, no referent id), so
// it is modeled as a Go fixed array rather than a slice.
type JOINPR_ENCRYPTED_USER_PASSWORD struct {
	Buffer [524]uint8
}
