package uuid

// UUIDInterface is an interface that defines the methods for a UUID
type UUIDInterface interface {
	// Marshal converts a UUID structure into a 16-byte array
	Marshal() ([]byte, error)
	// Unmarshal converts a 16-byte array into a UUID structure
	Unmarshal(data []byte) (int, error)
	// FromString parses a UUID string in the format "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
	FromString(s string) error
	// String returns the string representation of the UUID
	String() string
}
