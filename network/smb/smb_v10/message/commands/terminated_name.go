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
