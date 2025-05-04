package credentials

import (
	"errors"
	"regexp"
	"strings"
)

type Credentials struct {
	Domain   string
	Username string
	Password string

	LMHash string
	NTHash string
}

// NewCredentials creates a new Credentials object.
// authDomain is the domain to authenticate to.
// authUsername is the username to authenticate as.
// authPassword is the password to authenticate with.
// authHashes is the NT/LM hashes to use for authentication.
func NewCredentials(authDomain, authUsername, authPassword, authHashes string) (*Credentials, error) {
	lmHash, ntHash, err := ParseLMNTHashes(authHashes)
	if err != nil {
		return nil, err
	}

	return &Credentials{
		Domain:   authDomain,
		Username: authUsername,
		Password: authPassword,
		LMHash:   lmHash,
		NTHash:   ntHash,
	}, nil
}

// ParseLMNTHashes parses the NT/LM hashes and returns the LM hash and NT hash.
func ParseLMNTHashes(authHashes string) (string, string, error) {
	// Check if the authHashes string matches the pattern
	matched, err := regexp.MatchString(`(?i)^([0-9a-f]{32})?(:[0-9a-f]{32})?$`, strings.TrimSpace(authHashes))
	if err != nil {
		return "", "", err
	}
	if !matched {
		return "", "", errors.New("invalid hash format, it must be 32 characters of 0-9a-f followed by a colon and another 32 characters of 0-9a-f")
	}

	if !strings.Contains(authHashes, ":") {
		authHashes = ":" + authHashes
	}

	parts := strings.Split(authHashes, ":")

	lmHash, ntHash := parts[0], parts[1]
	if len(lmHash) != 32 {
		lmHash = ""
	}

	if len(ntHash) != 32 {
		ntHash = ""
	}

	return lmHash, ntHash, nil
}

// IsDomain returns true if the credentials are for a domain.
func (c *Credentials) IsDomainIdentity() bool {
	return c.Domain != ""
}

// IsLocal returns true if the credentials are for a local account.
func (c *Credentials) IsLocalIdentity() bool {
	return c.Domain == ""
}

// CanPassTheHash returns true if the credentials can be used to pass the hash attack.
func (c *Credentials) CanPassTheHash() bool {
	return c.NTHash != "" && c.Username != ""
}

// Getters

func (c *Credentials) GetLMHash() string {
	return c.LMHash
}

func (c *Credentials) GetNTHash() string {
	return c.NTHash
}

func (c *Credentials) GetDomain() string {
	return c.Domain
}

func (c *Credentials) GetUsername() string {
	return c.Username
}

func (c *Credentials) GetPassword() string {
	return c.Password
}
