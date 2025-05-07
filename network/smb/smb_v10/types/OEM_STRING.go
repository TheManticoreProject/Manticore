package types

// OEM_STRING
// Source: https://learn.microsoft.com/en-us/previous-versions/windows/hardware/kernel/ff558741(v=vs.85)
type OEM_STRING struct {
	SMB_STRING
}

// NewOEM_STRING creates a new OEM_STRING
//
// Returns:
// - A pointer to the new OEM_STRING
func NewOEM_STRING() *OEM_STRING {
	return &OEM_STRING{
		SMB_STRING: SMB_STRING{
			BufferFormat: SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING,
			Length:       0,
			Buffer:       []UCHAR{},
		},
	}
}

// NewOEM_STRINGFromString creates a new OEM_STRING from a string
//
// Parameters:
// - str: The string to create the OEM_STRING from
//
// Returns:
// - The new OEM_STRING
func NewOEM_STRINGFromString(str string) *OEM_STRING {
	return &OEM_STRING{
		SMB_STRING: SMB_STRING{
			BufferFormat: SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING,
			Length:       uint16(len(str)),
			Buffer:       []UCHAR(str),
		},
	}
}

// Marshal marshals the SMB_RESUME_KEY structure
//
// Returns:
// - A byte array representing the SMB_RESUME_KEY structure
// - An error if the marshaling fails
func (o *OEM_STRING) Marshal() ([]byte, error) {
	// This field MUST be 0x05, which indicates that a variable block is to follow.
	o.SetBufferFormat(SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	return o.SMB_STRING.Marshal()
}

// Unmarshal unmarshals the SMB_RESUME_KEY structure
//
// Parameters:
// - data: The byte array to unmarshal
//
// Returns:
// - The number of bytes unmarshalled
// - An error if the unmarshaling fails
func (o *OEM_STRING) Unmarshal(data []byte) (int, error) {
	bytesRead, err := o.SMB_STRING.Unmarshal(data)
	if err != nil {
		return 0, err
	}

	return bytesRead, nil
}

// GetString returns the string of the OEM_STRING
//
// Returns:
// - The string of the OEM_STRING
func (o *OEM_STRING) GetString() string {
	return string(o.Buffer)
}

// SetString sets the string of the OEM_STRING
//
// Parameters:
// - str: The string to set
func (o *OEM_STRING) SetString(str string) {
	o.Buffer = []UCHAR(str)
	o.Length = uint16(len(str))
}
