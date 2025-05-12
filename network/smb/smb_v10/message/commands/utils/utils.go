package utils

// GetNullTerminatedUnicodeString returns the null-terminated Unicode string from the data
// and the offset of the next byte after the null terminator
//
// Parameters:
// - data: The data to extract the null-terminated Unicode string from
//
// Returns:
// - The null-terminated Unicode string
// - The offset of the next byte after the null terminator
func GetNullTerminatedUnicodeString(data []byte) (string, int) {
	bytesString := []byte{}
	for i := 0; i < len(data); i += 2 {
		if data[i] == 0 && data[i+1] == 0 {
			break
		} else {
			bytesString = append(bytesString, data[i])
			bytesString = append(bytesString, data[i+1])
		}
	}
	return string(bytesString), len(bytesString) + 2
}

// GetNullTerminatedString returns the null-terminated string from the data
// and the offset of the next byte after the null terminator
//
// Parameters:
// - data: The data to extract the null-terminated string from
//
// Returns:
// - The null-terminated string
// - The offset of the next byte after the null terminator
func GetNullTerminatedString(data []byte) (string, int) {
	bytesString := []byte{}
	for i := 0; i < len(data); i++ {
		if data[i] == 0 {
			break
		} else {
			bytesString = append(bytesString, data[i])
		}
	}
	return string(bytesString), len(bytesString) + 1
}
