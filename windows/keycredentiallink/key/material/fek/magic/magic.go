package magic

// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/735fd27a-3f22-4926-93f9-0298bb67a84b
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/701a55dc-d062-4032-a2da-dbdfc384c8cf
//
// The key material of KEY_USAGE_FEK (section 2.2.20.5.3) is a combination of an
// RSA 2048 public key (RFC8017) and an AES-256 KDF key. The version of the buffer
// is controlled by the FekKeyVersion field of the CUSTOM_KEY_INFORMATION structure
// and MUST be set to 1.
const (
	// FEK_KEY_VERSION_1 indicates the version 1 layout of the KEY_USAGE_FEK key
	// material buffer. This is the only version currently defined by MS-ADTS.
	FEK_KEY_VERSION_1 uint8 = 0x01
)
