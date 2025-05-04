package ntlmv2

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/nt"
	"github.com/TheManticoreProject/Manticore/utils/encoding/utf16"
)

// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/aee311d6-21a7-4470-92a5-c4ecb022a87b

// NTLMv2 represents the components needed for NTLMv2 authentication
type NTLMv2 struct {
	Domain          string
	Username        string
	Password        string
	ServerChallenge [8]byte
	ClientChallenge [8]byte
	NTHash          [16]byte
	ResponseKeyNT   [16]byte
}

// NewNTLMv2 creates a new NTLMv2 instance with the provided credentials and challenges
func NewNTLMv2(domain, username, password string, serverChallenge, clientChallenge [8]byte) (*NTLMv2, error) {
	if len(serverChallenge) != 8 {
		return nil, errors.New("server challenge must be 8 bytes")
	}

	if len(clientChallenge) != 8 {
		return nil, errors.New("client challenge must be 8 bytes")
	}

	ntHash := nt.NTHash(password)

	ntlm := &NTLMv2{
		Domain:          domain,
		Username:        username,
		Password:        password,
		ServerChallenge: serverChallenge,
		ClientChallenge: clientChallenge,
		NTHash:          ntHash,
	}

	// Calculate the ResponseKeyNT (HMAC-MD5 of NT-Hash with username and domain)
	usernameUpper := strings.ToUpper(username)
	domainUpper := strings.ToUpper(domain)
	identity := utf16.EncodeUTF16LE(usernameUpper + domainUpper)

	h := hmac.New(md5.New, ntlm.NTHash[:])
	h.Write(identity)
	copy(ntlm.ResponseKeyNT[:], h.Sum(nil))

	return ntlm, nil
}

// NTLMv2HashHex computes the NTLMv2 response for a given domain, username, password,
// server challenge, and client challenge, and returns it as a hexadecimal string.
//
// Parameters:
//   - domain: The domain name
//   - username: The username
//   - password: The plaintext password
//   - serverChallenge: The 8-byte server challenge
//   - clientChallenge: The 8-byte client challenge
//
// Returns:
//   - The NTLMv2 response as a hexadecimal string
//   - An error if the computation fails
func (ntlm *NTLMv2) HashHex() (string, error) {
	response, err := ntlm.Hash()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(response), nil
}

// NTLMv2Hash computes the NTLMv2 response for a given domain, username, password,
// server challenge, and client challenge.
//
// Parameters:
//   - domain: The domain name
//   - username: The username
//   - password: The plaintext password
//   - serverChallenge: The 8-byte server challenge
//   - clientChallenge: The 8-byte client challenge
//
// Returns:
//   - The NTLMv2 response as a byte slice
//   - An error if the computation fails
func (ntlm *NTLMv2) Hash() ([]byte, error) {
	if len(ntlm.ServerChallenge) != 8 {
		return nil, errors.New("server challenge must be 8 bytes")
	}

	if len(ntlm.ClientChallenge) != 8 {
		return nil, errors.New("client challenge must be 8 bytes")
	}

	// Calculate the NT-Hash of the password
	ntHash := nt.NTHash(ntlm.Password)

	// Convert username and domain to uppercase and encode in UTF-16LE
	usernameUpper := strings.ToUpper(ntlm.Username)
	domainUpper := strings.ToUpper(ntlm.Domain)
	userDomain := utf16.EncodeUTF16LE(usernameUpper + domainUpper)

	// Calculate the NTLMv2 hash (HMAC-MD5 of NT-Hash with userDomain)
	v2Hash := hmac.New(md5.New, ntHash[:])
	v2Hash.Write(userDomain)
	v2HashBytes := v2Hash.Sum(nil)

	// Create the NTLMv2 blob with timestamp and domain name
	timestamp := make([]byte, 8)
	binary.LittleEndian.PutUint64(timestamp, uint64(time.Now().UnixNano()/100))
	blob := append([]byte{0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, timestamp...)
	blob = append(blob, ntlm.ClientChallenge[:]...)
	blob = append(blob, make([]byte, 4)...) // Reserved
	blob = append(blob, utf16.EncodeUTF16LE(ntlm.Domain)...)
	blob = append(blob, make([]byte, 4)...) // Reserved

	// Calculate the NTLMv2 response (HMAC-MD5 of v2-Hash with server challenge and blob)
	ntv2 := hmac.New(md5.New, v2HashBytes)
	ntv2.Write(ntlm.ServerChallenge[:])
	ntv2.Write(blob)
	ntv2Response := ntv2.Sum(nil)

	// Combine the NTLMv2 response and blob
	response := append(ntv2Response, blob...)

	return response, nil
}

// ToHashcatString converts the NTLMv2 response to a Hashcat string
//
// Returns:
//   - The Hashcat string
//   - An error if the conversion fails
func (ntlm *NTLMv2) ToHashcatString() (string, error) {
	response, err := ntlm.Hash()
	if err != nil {
		return "", err
	}

	hashcatString := fmt.Sprintf(
		"%s::%s:%s:%s:%s",
		ntlm.Username,
		ntlm.Domain,
		hex.EncodeToString(ntlm.ServerChallenge[:]),
		hex.EncodeToString(ntlm.ClientChallenge[:]),
		hex.EncodeToString(response),
	)

	return hashcatString, nil
}
