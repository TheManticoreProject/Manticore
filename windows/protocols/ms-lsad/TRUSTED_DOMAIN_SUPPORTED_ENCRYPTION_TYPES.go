package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TRUSTED_DOMAIN_SUPPORTED_ENCRYPTION_TYPES communicates the supported encryption types
// of a trusted domain ([MS-LSAD] 2.2.7.18).
type TRUSTED_DOMAIN_SUPPORTED_ENCRYPTION_TYPES struct {
	SupportedEncryptionTypes ndr.DWORD
}
