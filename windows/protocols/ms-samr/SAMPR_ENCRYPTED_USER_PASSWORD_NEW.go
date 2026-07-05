package mssamr

// SAMPR_ENCRYPTED_USER_PASSWORD_NEW carries an encrypted
// SAMPR_USER_PASSWORD_NEW ([MS-SAMR] 2.2.6.22). Buffer is a fixed
// (256*2)+4+16 = 532-byte array.
type SAMPR_ENCRYPTED_USER_PASSWORD_NEW struct {
	Buffer [532]byte
}
