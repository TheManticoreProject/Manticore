package commands

// Helpers for the string fields that a command carries raw in its data block.
//
// [MS-CIFS] describes several such fields — SMB_COM_TRANSACTION's Name, and
// SMB_COM_NT_CREATE_ANDX's FileName among them — as a null-terminated array of
// OEM characters, or of 16-bit Unicode characters when SMB_FLAGS2_UNICODE is set
// in the request's Flags2. None of them carries an SMB_STRING buffer-format byte:
// the commands that do carry one declare it as a separate field. Reading such a
// field through SMB_STRING treats the name's first character as a format code,
// which fails outright for a name beginning with a backslash.
//
// The width of the terminator follows the encoding, so both helpers take it.

// nameTerminator returns the terminator that ends a raw name field.
func nameTerminator(unicode bool) []byte {
	if unicode {
		return []byte{0x00, 0x00}
	}
	return []byte{0x00}
}

// readTerminatedName reads a raw name from the start of a data block and returns
// the name's bytes together with the number of bytes it occupied, terminator
// included.
//
// A name that runs to the end of the block without a terminator is returned
// whole: the block is bounded by ByteCount, so there is nothing to overrun, and
// refusing the message outright would be harsher than the field warrants.
func readTerminatedName(data []byte, unicode bool) (name []byte, size int) {
	if !unicode {
		for index := 0; index < len(data); index++ {
			if data[index] == 0x00 {
				return data[:index], index + 1
			}
		}
		return data, len(data)
	}

	// A UTF-16 name ends at the first pair of zero bytes on an even offset. An odd
	// offset cannot end it: the low half of one character and the high half of the
	// next can both be zero without the string having ended.
	for index := 0; index+1 < len(data); index += 2 {
		if data[index] == 0x00 && data[index+1] == 0x00 {
			return data[:index], index + 2
		}
	}
	return data, len(data)
}

// treeConnectAndxDataOffset is where an SMB_COM_TREE_CONNECT_ANDX request's data
// block begins, measured from the start of the SMB header: SMB_HEADER_SIZE(32) +
// WordCount(1) + four parameter words(8) + ByteCount(2). The alignment of a
// Unicode field inside the block is relative to the header, so the block's own
// offset is part of the arithmetic.
const treeConnectAndxDataOffset = 43

// readTerminatedField measures a null-terminated field at the start of a buffer,
// returning the number of bytes it occupies including its terminator, and whether
// a terminator was found at all.
//
// It is readTerminatedName measured rather than sliced, for the fields a decoder
// keeps with their terminator. Reporting whether the field was terminated is what
// lets a caller tell an unterminated field running to the end of the block — after
// which there is nothing — from a terminated one with more fields behind it.
func readTerminatedField(data []byte, unicode bool) (size int, terminated bool) {
	value, size := readTerminatedName(data, unicode)
	return size, size > len(value)
}

// ntCreateAndxDataOffset is where an SMB_COM_NT_CREATE_ANDX request's data block
// begins, measured from the start of the SMB header: SMB_HEADER_SIZE(32) +
// WordCount(1) + twenty-four parameter words(48) + ByteCount(2).
const ntCreateAndxDataOffset = 83

// Where the data block of the two-name commands begins, measured from the start of
// the SMB header: SMB_HEADER_SIZE(32) + WordCount(1) + the parameter words +
// ByteCount(2). A Unicode string inside the block is aligned relative to the
// header, so the block's own offset is part of that arithmetic.
const (
	// SMB_COM_RENAME carries one parameter word, SearchAttributes.
	renameDataOffset = 37
	// SMB_COM_NT_RENAME carries four: SearchAttributes, InformationLevel and the
	// two halves of ClusterCount.
	ntRenameDataOffset = 43
)
