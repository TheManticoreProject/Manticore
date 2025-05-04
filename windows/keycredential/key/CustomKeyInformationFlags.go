package key

// CustomKeyInformationFlags represents custom key flags.
//
// See: https://msdn.microsoft.com/en-us/library/mt220496.aspx
type CustomKeyInformationFlags struct {
	Value uint8
	Name  []string

	// Internal
	RawBytes     []byte
	RawBytesSize uint32
}

const (
	// No flags specified.
	CustomKeyInformationFlags_None uint8 = 0

	// Reserved for future use. (CUSTOMKEYINFO_FLAGS_ATTESTATION)
	CustomKeyInformationFlags_Attestation uint8 = 0x01

	// During creation of this key, the requesting client authenticated using
	// only a single credential. (CUSTOMKEYINFO_FLAGS_MFA_NOT_USED)
	CustomKeyInformationFlags_MFANotUsed uint8 = 0x02
)

// Parse parses the provided byte slice into the CustomKeyInformationFlags structure.
//
// Parameters:
// - value: A byte slice containing the raw key flags to be parsed.
//
// Note:
// The function expects the byte slice to contain a single byte representing the key flags.
// It extracts the flags value from the byte slice and assigns it to the CustomKeyInformationFlags structure.
func (kf *CustomKeyInformationFlags) FromBytes(value byte) {
	kf.Value = value

	kf.Name = []string{}
	if kf.Value&CustomKeyInformationFlags_Attestation == CustomKeyInformationFlags_Attestation {
		kf.Name = append(kf.Name, "Attestation")
	}
	if kf.Value&CustomKeyInformationFlags_MFANotUsed == CustomKeyInformationFlags_MFANotUsed {
		kf.Name = append(kf.Name, "MFA not used")
	}

	if len(kf.Name) == 0 {
		kf.Name = append(kf.Name, "None")
	}
}
