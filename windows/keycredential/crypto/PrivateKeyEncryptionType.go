package crypto

type PrivateKeyEncryptionType int

const (
	NONE PrivateKeyEncryptionType = iota
	PasswordRC4
	PasswordRC2CBC
)

func (pket PrivateKeyEncryptionType) String() string {
	return [...]string{"NONE", "PasswordRC4", "PasswordRC2CBC"}[pket]
}
