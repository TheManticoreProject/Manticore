package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
)

// preauthHashLength is the size of the SMB 3.1.1 pre-authentication integrity
// hash (SHA-512 digest).
const preauthHashLength = 64

// preauthUpdate folds a message into the running SMB 3.1.1 pre-authentication
// integrity hash: hash_new = SHA-512(hash_prev || message), where hash_prev is
// the 64-byte previous value (all zero at the start of the connection). See
// MS-SMB2 3.1.4.2 / the pre-auth integrity computation.
func preauthUpdate(prev, message []byte) []byte {
	h := sha512.New()
	h.Write(prev)
	h.Write(message)
	return h.Sum(nil)
}

// SMB 3.x SP800-108 key-derivation labels and contexts (MS-SMB2 3.1.4.2). Each
// string carries its trailing NUL byte, which also serves as the SP800-108
// separator between the Label and the Context.
var (
	// Dialects 3.0 and 3.0.2 use constant labels and contexts.
	kdfLabelSigning30 = []byte("SMB2AESCMAC\x00")
	kdfContextSign30  = []byte("SmbSign\x00")
	kdfLabelApp30     = []byte("SMB2APP\x00")
	kdfContextApp30   = []byte("SmbRpc\x00")
	kdfLabelCipher30  = []byte("SMB2AESCCM\x00")
	// From the node's perspective the "ServerIn " context yields the key the
	// client encrypts with (and the server decrypts with); "ServerOut" is the
	// reverse. The trailing space in "ServerIn " pads it to the length of
	// "ServerOut".
	kdfContextServerIn  = []byte("ServerIn \x00")
	kdfContextServerOut = []byte("ServerOut\x00")

	// Dialect 3.1.1 uses distinct labels; the context is the pre-authentication
	// integrity hash value instead of a constant.
	kdfLabelSigning311 = []byte("SMBSigningKey\x00")
	kdfLabelApp311     = []byte("SMBAppKey\x00")
	kdfLabelC2SCipher  = []byte("SMBC2SCipherKey\x00")
	kdfLabelS2CCipher  = []byte("SMBS2CCipherKey\x00")
)

// sp800108CounterKDF derives a 128-bit key using the NIST SP800-108 KDF in
// counter mode with HMAC-SHA256 as the PRF, as required by MS-SMB2 3.1.4.2 for
// the SMB 3.x key hierarchy. The counter width r is 32 bits and the output
// length L is 128 bits, so a single PRF invocation produces the whole key.
//
// The fixed input string is [i]_32 || Label || 0x00 || Context || [L]_32 with
// i = 1 and L = 128 (both 32-bit big-endian). The MS-SMB2 Label/Context byte
// strings already carry their own trailing NUL, and the KDF inserts the
// mandatory 0x00 separator between the label and the context.
func sp800108CounterKDF(ki, label, context []byte) []byte {
	mac := hmac.New(sha256.New, ki)

	var counter [4]byte
	binary.BigEndian.PutUint32(counter[:], 1)
	mac.Write(counter[:])

	mac.Write(label)
	mac.Write([]byte{0x00})
	mac.Write(context)

	var length [4]byte
	binary.BigEndian.PutUint32(length[:], 128)
	mac.Write(length[:])

	return mac.Sum(nil)[:16]
}

// deriveSMB3Keys computes the SMB 3.x signing, encryption, decryption, and
// application keys for a session from its 16-byte SessionKey, per MS-SMB2
// 3.1.4.2. For the 3.1.1 dialect the KDF context is the session's
// pre-authentication integrity hash value; for 3.0/3.0.2 it is a set of fixed
// constants. The derived SigningKey replaces the session key as the signing key
// (unlike the 2.x dialects, where the two are identical).
//
// EncryptionKey is the key this client uses to encrypt the messages it sends;
// DecryptionKey is the key it uses to decrypt the server's replies.
func deriveSMB3Keys(session *Session, dialect dialects.Dialect, preauthHash []byte) {
	key := session.SessionKey

	switch dialect {
	case dialects.SMB2_DIALECT_3_0_0, dialects.SMB2_DIALECT_3_0_2:
		session.SigningKey = sp800108CounterKDF(key, kdfLabelSigning30, kdfContextSign30)
		session.ApplicationKey = sp800108CounterKDF(key, kdfLabelApp30, kdfContextApp30)
		session.EncryptionKey = sp800108CounterKDF(key, kdfLabelCipher30, kdfContextServerIn)
		session.DecryptionKey = sp800108CounterKDF(key, kdfLabelCipher30, kdfContextServerOut)
	case dialects.SMB2_DIALECT_3_1_1:
		session.SigningKey = sp800108CounterKDF(key, kdfLabelSigning311, preauthHash)
		session.ApplicationKey = sp800108CounterKDF(key, kdfLabelApp311, preauthHash)
		session.EncryptionKey = sp800108CounterKDF(key, kdfLabelC2SCipher, preauthHash)
		session.DecryptionKey = sp800108CounterKDF(key, kdfLabelS2CCipher, preauthHash)
	}
}

// isSMB3Dialect reports whether a negotiated dialect belongs to the SMB 3.x
// family, which uses the SP800-108 key hierarchy and AES-based signing.
func isSMB3Dialect(d dialects.Dialect) bool {
	switch d {
	case dialects.SMB2_DIALECT_3_0_0, dialects.SMB2_DIALECT_3_0_2, dialects.SMB2_DIALECT_3_1_1:
		return true
	default:
		return false
	}
}
