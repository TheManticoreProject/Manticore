package structures

// NL_TRUST_PASSWORD ([MS-NRPC] 2.2.1.3.7) holds a new machine/trust password as a fixed
// 256-WCHAR (512-octet) buffer followed by its byte length, for a 516-octet structure. An
// all-zero value (empty buffer, Length 0) sets the target account password to the empty
// string.
type NL_TRUST_PASSWORD struct {
	Buffer [512]byte
	Length uint32
}
