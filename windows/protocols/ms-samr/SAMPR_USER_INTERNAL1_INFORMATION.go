package mssamr

// SAMPR_USER_INTERNAL1_INFORMATION carries encrypted NT/LM OWF password hashes
// and presence flags ([MS-SAMR] 2.2.6.24).
type SAMPR_USER_INTERNAL1_INFORMATION struct {
	EncryptedNtOwfPassword ENCRYPTED_NT_OWF_PASSWORD
	EncryptedLmOwfPassword ENCRYPTED_LM_OWF_PASSWORD
	NtPasswordPresent      uint8
	LmPasswordPresent      uint8
	PasswordExpired        uint8
}
