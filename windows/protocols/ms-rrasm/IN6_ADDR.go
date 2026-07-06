package msrrasm

// IN6_ADDR is an on-wire IPv6 address ([MS-RRASM] 2.2.1.2.221, RFC 2553). In the
// IDL it is a C union of `UCHAR Byte[16]` and `USHORT Word[8]`; both arms occupy
// 16 bytes, so it is modeled as a fixed 16-byte field.
type IN6_ADDR struct {
	Byte [16]byte
}
