package commands

import "fmt"

// Locating the parameter and data blocks of a transaction message.
//
// All three transaction families — SMB_COM_TRANSACTION, SMB_COM_TRANSACTION2 and
// SMB_COM_NT_TRANSACT, primary and secondary alike — carry their payload in two
// blocks inside SMB_Data.Bytes, each preceded by however much padding the sender
// chose. The padding is not described by a length: it is implied by
// ParameterOffset and DataOffset, which [MS-CIFS] measures "from the start of the
// SMB Header" and which it requires the receiver to use — of ParameterOffset,
// "Server implementations MUST use this value to locate the transaction parameter
// block within the request".
//
// Deriving the padding by subtracting the field lengths from the block's total
// size instead looks equivalent and is not. It assumes the block ends exactly
// where the last field ends, and a real client pads the END of the block as well
// as the middle — to a four-byte boundary — so the surplus is charged to the
// leading pad and every field in both blocks is read at the wrong offset. That is
// invisible against a sender whose blocks happen to end flush, which is why it can
// survive a round-trip test.

// smbHeaderSize is the size of the fixed SMB header. It is defined here rather
// than imported to keep this package free of a dependency on the header package.
const smbHeaderSize = 32

// transactionBlockStart returns the offset, from the start of the SMB header, of
// the first byte of SMB_Data.Bytes — the byte the declared offsets are relative
// to.
//
// It is derived from the parameter block actually decoded rather than from a
// per-command constant, so it stays correct for the variable-length Setup array
// that every one of these commands carries.
func transactionBlockStart(parameterWords []byte) int {
	// WordCount(1) + the parameter words + ByteCount(2).
	return smbHeaderSize + 1 + len(parameterWords) + 2
}

// locateTransactionBlock resolves one declared block within SMB_Data.Bytes.
//
// A count of zero yields an empty block and no offset check, since [MS-CIFS]
// allows the offset to be left at zero when the count is — a sender with nothing
// to place has no offset to report.
//
// Parameters:
//   - block: the bytes of SMB_Data.Bytes
//   - blockStart: the offset of block[0] from the start of the SMB header
//   - declaredOffset: the header-relative offset the message declares
//   - count: the number of bytes the message declares at that offset
//   - name: the field's name, for the error
//
// Returns:
//   - The bytes of the block
//   - The offset of its first byte within block, or len(block) when it is empty
//   - An error if the declared offset and count do not lie within the block
func locateTransactionBlock(
	block []byte,
	blockStart int,
	declaredOffset int,
	count int,
	name string,
) ([]byte, int, error) {
	if count == 0 {
		return []byte{}, len(block), nil
	}
	if count < 0 {
		return nil, 0, fmt.Errorf("%s count is negative (%d)", name, count)
	}

	at := declaredOffset - blockStart
	if at < 0 {
		return nil, 0, fmt.Errorf("%s offset %d precedes the data block, which starts at %d",
			name, declaredOffset, blockStart)
	}
	if at+count > len(block) {
		return nil, 0, fmt.Errorf("%s of %d bytes at offset %d overruns the %d-byte data block",
			name, count, declaredOffset, len(block))
	}

	return block[at : at+count], at, nil
}

// locateTransactionBlocks resolves both declared blocks of a transaction message
// and the padding that precedes each.
//
// The pads are returned as they were received rather than as lengths: a sender is
// free to put anything in them — [MS-CIFS] says only that they SHOULD be zero and
// MUST be ignored — and keeping the bytes lets a re-marshalling reproduce the
// message it came from.
//
// Parameters:
//   - block: the bytes of SMB_Data.Bytes
//   - parameterWords: the decoded parameter words, used to locate block[0]
//   - nameSize: the size of the name field that precedes the first pad
//   - parameterOffset, parameterCount: the declared parameter block
//   - dataOffset, dataCount: the declared data block
//
// Returns:
//   - pad1, parameters, pad2, data
//   - The total number of bytes of block accounted for
//   - An error if either declared block does not lie within the block
func locateTransactionBlocks(
	block []byte,
	parameterWords []byte,
	nameSize int,
	parameterOffset, parameterCount int,
	dataOffset, dataCount int,
) (pad1, parameters, pad2, data []byte, consumed int, err error) {
	blockStart := transactionBlockStart(parameterWords)

	parameters, parametersAt, err := locateTransactionBlock(
		block, blockStart, parameterOffset, parameterCount, "the parameter block")
	if err != nil {
		return nil, nil, nil, nil, 0, err
	}
	data, dataAt, err := locateTransactionBlock(
		block, blockStart, dataOffset, dataCount, "the data block")
	if err != nil {
		return nil, nil, nil, nil, 0, err
	}

	// The pads are whatever lies between the name and the parameters, and between
	// the parameters and the data. An empty block reports its position as the end
	// of the buffer, which leaves the pad before it empty rather than negative.
	if parametersAt < nameSize {
		return nil, nil, nil, nil, 0, fmt.Errorf(
			"the parameter block at offset %d overlaps the %d-byte name that precedes it",
			parameterOffset, nameSize)
	}
	pad1 = block[nameSize:parametersAt]

	afterParameters := parametersAt + parameterCount
	if parameterCount == 0 {
		afterParameters = nameSize
	}
	switch {
	case dataCount == 0:
		pad2 = []byte{}
	case dataAt < afterParameters:
		return nil, nil, nil, nil, 0, fmt.Errorf(
			"the data block at offset %d precedes the end of the parameter block", dataOffset)
	default:
		pad2 = block[afterParameters:dataAt]
	}

	consumed = afterParameters
	if dataCount > 0 {
		consumed = dataAt + dataCount
	}
	return pad1, parameters, pad2, data, consumed, nil
}

// Word counts of the transaction-family messages, which fix where their data
// block begins. Each was confirmed against the marshalled output rather than read
// off the specification, since it is the layout this package actually produces
// that the declared offsets have to describe.
const (
	// transactionWordCount and transaction2WordCount are the fixed part; both
	// families add one word per setup word.
	transactionWordCount  = 14
	transaction2WordCount = 14
	// ntTransactWordCount likewise, for the wider family.
	ntTransactWordCount = 19

	// The secondary messages carry no setup words.
	transactionSecondaryWordCount  = 8
	transaction2SecondaryWordCount = 9
	ntTransactSecondaryWordCount   = 18
)

// deriveTransactionOffsets computes the header-relative offsets of the parameter
// and data blocks from the layout a marshaller is about to write.
//
// A marshaller has to declare where it put the blocks, and the only way for the
// declaration to be right is to derive it from the placement. Leaving the fields
// for a caller to fill in — which is what they were before — means they are
// either unset, and the receiver cannot find the blocks, or set to something that
// disagrees with the bytes, which is worse: a receiver that trusts them, as
// [MS-CIFS] tells it to, reads the payload at the wrong place and gets plausible
// nonsense rather than an error.
//
// An empty block is declared at offset zero, which [MS-CIFS] permits when its
// count is zero: there is no position to report for bytes that are not there.
//
// Parameters:
//   - wordCount: the message's WordCount, setup words included
//   - nameSize: the size of the name field at the start of the data block
//   - pad1Size, parameterCount, pad2Size, dataCount: the layout being written
//
// Returns:
//   - The offsets to declare for the parameter and data blocks
func deriveTransactionOffsets(
	wordCount int,
	nameSize, pad1Size, parameterCount, pad2Size, dataCount int,
) (parameterOffset, dataOffset int) {
	blockStart := smbHeaderSize + 1 + 2*wordCount + 2

	parameterOffset = blockStart + nameSize + pad1Size
	dataOffset = parameterOffset + parameterCount + pad2Size

	if parameterCount == 0 {
		parameterOffset = 0
	}
	if dataCount == 0 {
		dataOffset = 0
	}
	return parameterOffset, dataOffset
}
