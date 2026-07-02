package mseerr

// BinaryEEInfo models the BinaryEEInfo structure ([MS-EERR] 2.2.1.3): an opaque binary
// blob. PBlob is a [unique] pointer to a conformant byte array bounded by NSize; nil
// when NSize is 0.
type BinaryEEInfo struct {
	NSize int16
	PBlob []uint8 `ndr:"unique,size_is=NSize"`
}
