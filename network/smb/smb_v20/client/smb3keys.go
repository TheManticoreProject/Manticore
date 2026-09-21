package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
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

// sp800108CounterKDF derives a key of the requested bit-length using the NIST
// SP800-108 KDF in counter mode with HMAC-SHA256 as the PRF, as required by
// MS-SMB2 3.1.4.2 for the SMB 3.x key hierarchy. The counter width r is 32
// bits and bits is the desired output length L (128 or 256). Since HMAC-SHA256
// produces 256 bits per iteration, a single PRF invocation suffices for both.
//
// The fixed input string is [i]_32 || Label || 0x00 || Context || [L]_32 with
// i = 1 (32-bit big-endian). The MS-SMB2 Label/Context byte strings already
// carry their own trailing NUL, and the KDF inserts the mandatory 0x00
// separator between the label and the context.
func sp800108CounterKDF(ki, label, context []byte, bits int) []byte {
	mac := hmac.New(sha256.New, ki)

	var counter [4]byte
	binary.BigEndian.PutUint32(counter[:], 1)
	mac.Write(counter[:])

	mac.Write(label)
	mac.Write([]byte{0x00})
	mac.Write(context)

	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(bits))
	mac.Write(length[:])

	return mac.Sum(nil)[:bits/8]
}

// deriveSMB3Keys computes the SMB 3.x signing, encryption, decryption, and
// application keys for a session from its 16-byte SessionKey, per MS-SMB2
// 3.1.4.2. For the 3.1.1 dialect the KDF context is the session's
// pre-authentication integrity hash value; for 3.0/3.0.2 it is a set of fixed
// constants. The derived SigningKey replaces the session key as the signing key
// (unlike the 2.x dialects, where the two are identical).
//
// EncryptionKey is the key this client uses to encrypt the messages it sends;
// DecryptionKey is the key it uses to decrypt the server's replies. When the
// negotiated cipher is AES-256-CCM or AES-256-GCM the encryption and decryption
// keys are 256 bits; all other keys remain 128 bits.
func deriveSMB3Keys(session *Session, dialect dialects.Dialect, preauthHash []byte, cipher uint16, signingAlg int) {
	key := session.SessionKey

	switch dialect {
	case dialects.SMB2_DIALECT_3_0_0, dialects.SMB2_DIALECT_3_0_2:
		session.SigningKey = sp800108CounterKDF(key, kdfLabelSigning30, kdfContextSign30, 128)
		session.ApplicationKey = sp800108CounterKDF(key, kdfLabelApp30, kdfContextApp30, 128)
		session.EncryptionKey = sp800108CounterKDF(key, kdfLabelCipher30, kdfContextServerIn, 128)
		session.DecryptionKey = sp800108CounterKDF(key, kdfLabelCipher30, kdfContextServerOut, 128)
	case dialects.SMB2_DIALECT_3_1_1:
		cipherBits := 128
		if cipher == commands.SMB2_ENCRYPTION_AES256_CCM || cipher == commands.SMB2_ENCRYPTION_AES256_GCM {
			cipherBits = 256
		}
		signingBits := 128
		if signingAlg == commands.SMB2_SIGNING_ALG_AES_GMAC {
			signingBits = 256
		}
		session.SigningKey = sp800108CounterKDF(key, kdfLabelSigning311, preauthHash, signingBits)
		session.ApplicationKey = sp800108CounterKDF(key, kdfLabelApp311, preauthHash, 128)
		session.EncryptionKey = sp800108CounterKDF(key, kdfLabelC2SCipher, preauthHash, cipherBits)
		session.DecryptionKey = sp800108CounterKDF(key, kdfLabelS2CCipher, preauthHash, cipherBits)
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
