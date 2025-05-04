package pkcs7

// Adapted from: https://github.com/zenazn/pkcs7pad/

import (
	"crypto/subtle"
	"errors"
)

var (
	ErrInvalidPadding          = errors.New("invalid padding")
	ErrUnPaddingEmptyBuffer    = errors.New("cannot unpad an empty buffer")
	ErrBlockSizeLessThanOne    = errors.New("block size must be greater than 0")
	ErrBlockSizeGreaterThan255 = errors.New("block size must be less than 255")
)

// Pad appends padding to the input buffer to make its length a multiple of the block size.
//
// Parameters:
// - buffer: A byte slice representing the input data to be padded.
// - blockSize: An unsigned 8-bit integer specifying the block size for padding.
//
// Returns:
// - A byte slice containing the padded data.
// - An error if the block size is less than 1 or greater than 255.
//
// The function performs the following steps:
// 1. Checks if the block size is valid (between 1 and 255). If not, it returns an error.
// 2. Calculates the padding length required to make the buffer length a multiple of the block size.
// 3. Appends the padding bytes to the buffer. Each padding byte has a value equal to the padding length.
//
// Example usage:
// paddedData, err := Pad([]byte("example"), 8)
//
//	if err != nil {
//	    fmt.Printf("Error padding data: %s\n", err)
//	}
//
// fmt.Printf("Padded data: %x\n", paddedData)
func Pad(buffer []byte, blockSize uint8) ([]byte, error) {
	if blockSize < 1 {
		return nil, ErrBlockSizeLessThanOne
	}

	padLen := int(blockSize) - (len(buffer) % int(blockSize))

	for i := 0; i < padLen; i++ {
		buffer = append(buffer, byte(padLen))
	}

	return buffer, nil
}

func Unpad(buffer []byte) ([]byte, error) {
	if len(buffer) == 0 {
		return nil, ErrUnPaddingEmptyBuffer
	}

	// Detect the padding length
	padLen := buffer[len(buffer)-1]

	// Check if the padding length is valid
	blockSize := 255
	good := 1
	if blockSize > len(buffer) {
		blockSize = len(buffer)
	}
	for i := 0; i < blockSize; i++ {
		b := buffer[len(buffer)-1-i]

		outOfRange := subtle.ConstantTimeLessOrEq(int(padLen), i)
		equal := subtle.ConstantTimeByteEq(padLen, b)
		good &= subtle.ConstantTimeSelect(outOfRange, 1, equal)
	}

	good &= subtle.ConstantTimeLessOrEq(1, int(padLen))
	good &= subtle.ConstantTimeLessOrEq(int(padLen), len(buffer))

	if good != 1 {
		return nil, ErrInvalidPadding
	}

	return buffer[:len(buffer)-int(padLen)], nil
}
