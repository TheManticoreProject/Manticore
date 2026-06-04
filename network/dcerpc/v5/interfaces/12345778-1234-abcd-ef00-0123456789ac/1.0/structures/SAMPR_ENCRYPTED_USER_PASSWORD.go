package structures

// SAMPR_ENCRYPTED_USER_PASSWORD carries an encrypted SAMPR_USER_PASSWORD
// ([MS-SAMR] 2.2.6.21). Buffer is a fixed (256*2)+4 = 516-byte array.
type SAMPR_ENCRYPTED_USER_PASSWORD struct {
	Buffer [516]byte
}
