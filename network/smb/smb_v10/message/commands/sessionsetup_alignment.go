package commands

import (
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// sessionSetupAndxDataOffset is where the SMB_Data bytes of an extended-security
// SMB_COM_SESSION_SETUP_ANDX begin, measured from the start of the SMB header:
// SMB_HEADER_SIZE(32) + WordCount(1) + Words(24) + ByteCount(2).
//
// It is odd, which is the whole reason the Unicode strings that follow the
// security blob need an alignment byte ([MS-SMB] 2.2.4.6.1).
const sessionSetupAndxDataOffset = 59

// decodeNativeString renders the bytes of a NativeOS or NativeLanMan field as a
// string, in the encoding the enclosing message declared.
func decodeNativeString(raw []byte, unicode bool) string {
	if unicode {
		return utf16.DecodeUTF16LE(raw)
	}
	return string(raw)
}
