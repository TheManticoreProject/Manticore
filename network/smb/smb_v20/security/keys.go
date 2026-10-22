package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

// PreauthHashSize is the size of the SMB 3.1.1 pre-authentication integrity
// hash (SHA-512 digest).
const PreauthHashSize = 64

// PreauthUpdate folds a message into the running SMB 3.1.1 pre-authentication
// integrity hash: hash_new = SHA-512(hash_prev || message), where hash_prev is
// the 64-byte previous value (all zero at the start of the connection). See
// MS-SMB2 3.1.4.2 / the pre-auth integrity computation.
func PreauthUpdate(prev, message []byte) []byte {
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

// Keys is the SMB 3.x key hierarchy of one session (MS-SMB2 3.1.4.2).
// ClientToServer encrypts what the client sends; ServerToClient what the server
// sends. With an AES-256 cipher both are 256 bits, every other key is 128 bits.
type Keys struct {
	Signing        []byte
	Application    []byte
	ClientToServer []byte
	ServerToClient []byte
}

// DeriveKeys derives the SMB 3.x keys from a 16-byte session key. For 3.1.1 the
// KDF context is the session's pre-authentication integrity hash; 3.0 and 3.0.2
// use constant contexts. Other dialects yield empty keys.
func DeriveKeys(sessionKey []byte, dialect dialects.Dialect, preauthHash []byte, cipher uint16) Keys {
	switch dialect {
	case dialects.SMB2_DIALECT_3_0_0, dialects.SMB2_DIALECT_3_0_2:
		return Keys{
			Signing:        sp800108CounterKDF(sessionKey, kdfLabelSigning30, kdfContextSign30, 128),
			Application:    sp800108CounterKDF(sessionKey, kdfLabelApp30, kdfContextApp30, 128),
			ClientToServer: sp800108CounterKDF(sessionKey, kdfLabelCipher30, kdfContextServerIn, 128),
			ServerToClient: sp800108CounterKDF(sessionKey, kdfLabelCipher30, kdfContextServerOut, 128),
		}
	case dialects.SMB2_DIALECT_3_1_1:
		cipherBits := 128
		if cipher == commands.SMB2_ENCRYPTION_AES256_CCM || cipher == commands.SMB2_ENCRYPTION_AES256_GCM {
			cipherBits = 256
		}
		return Keys{
			Signing:        sp800108CounterKDF(sessionKey, kdfLabelSigning311, preauthHash, 128),
			Application:    sp800108CounterKDF(sessionKey, kdfLabelApp311, preauthHash, 128),
			ClientToServer: sp800108CounterKDF(sessionKey, kdfLabelC2SCipher, preauthHash, cipherBits),
			ServerToClient: sp800108CounterKDF(sessionKey, kdfLabelS2CCipher, preauthHash, cipherBits),
		}
	}
	return Keys{}
}

// IsSMB3 reports whether a negotiated dialect belongs to the SMB 3.x
// family, which uses the SP800-108 key hierarchy and AES-based signing.
func IsSMB3(d dialects.Dialect) bool {
	switch d {
	case dialects.SMB2_DIALECT_3_0_0, dialects.SMB2_DIALECT_3_0_2, dialects.SMB2_DIALECT_3_1_1:
		return true
	default:
		return false
	}
}
