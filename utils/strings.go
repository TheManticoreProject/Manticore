package utils

import "fmt"

// PadStringRight pads the input string with the specified padChar on the right side until the string reaches the specified length.
// If the input string is already longer than or equal to the specified length, it returns the input string as is.
//
// Parameters:
// - input: The original string to be padded.
// - padChar: The character to pad the input string with.
// - length: The desired total length of the output string.
//
// Returns:
// - A new string that is padded with padChar on the right side to the specified length.
//
// Example:
// PadStringRight("hello", "*", 8) returns "hello***"
//
// Note:
// The function adds a space to the input string before padding to ensure proper alignment.
func PadStringRight(input string, padChar string, length int) string {
	if len(input) < length {
		for range length - len(input) {
			input += padChar
		}
	}

	return input
}

// PadStringLeft pads the input string with the specified padChar on the left side until the string reaches the specified length.
// If the input string is already longer than or equal to the specified length, it returns the input string as is.
//
// Parameters:
// - input: The original string to be padded.
// - padChar: The character to pad the input string with.
// - length: The desired total length of the output string.
//
// Returns:
// - A new string that is padded with padChar on the left side to the specified length.

// Example:
// PadStringLeft("hello", "*", 8) returns "***hello"
//
// Note:
// The function adds a space to the input string before padding to ensure proper alignment.
func PadStringLeft(input string, padChar string, length int) string {
	if len(input) < length {
		for range length - len(input) {
			input = padChar + input
		}
	}

	return input
}

// SizeInBytes converts a size in bytes to a human-readable string representation using binary prefixes (KiB, MiB, GiB).
//
// Parameters:
// - size: The size in bytes to be converted.
//
// Returns:
// - A string representing the size in a human-readable format with binary prefixes.
//
// Example:
// SizeInBytes(1048576) returns "1.00 MiB"
//
// Note:
// The function uses binary prefixes where 1 KiB = 1024 bytes, 1 MiB = 1024 KiB, and 1 GiB = 1024 MiB.
func SizeInBytes(size uint64) string {
	KiB := uint64(1024)
	MiB := KiB * 1024
	GiB := MiB * 1024
	TiB := GiB * 1024
	PiB := TiB * 1024

	switch {
	case size >= PiB:
		return fmt.Sprintf("%.2f PiB", float64(size)/float64(PiB))
	case size >= TiB:
		return fmt.Sprintf("%.2f TiB", float64(size)/float64(TiB))
	case size >= GiB:
		return fmt.Sprintf("%.2f GiB", float64(size)/float64(GiB))
	case size >= MiB:
		return fmt.Sprintf("%.2f MiB", float64(size)/float64(MiB))
	case size >= KiB:
		return fmt.Sprintf("%.2f KiB", float64(size)/float64(KiB))
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}
