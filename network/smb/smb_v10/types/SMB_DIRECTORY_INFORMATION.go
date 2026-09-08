package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

// SMB_DIRECTORY_INFORMATION
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/b8674ab7-70a2-4b8b-bc30-3137b0ed4284
type SMB_DIRECTORY_INFORMATION struct {
	// ResumeKey (21 bytes): SMB_Resume_Key While each DirectoryInformationData entry has a ResumeKey field, the client
	// MUST use only the ResumeKey value from the last DirectoryInformationData entry when continuing the search with a
	// subsequent SMB_COM_SEARCH command.
	ResumeKey SMB_RESUME_KEY

	// FileAttributes (1 byte): These are the file system attributes of the file.
	FileAttributes UCHAR

	// LastWriteTime (2 bytes): The time at which the file was last modified.
	//
	// This is the 2-byte MS-DOS time, not the 8-byte FILETIME that the SMB_TIME
	// alias resolves to — the two share a name across [MS-CIFS] and [MS-DTYP] and
	// only one of them fits this field.
	LastWriteTime SMB_TIME_DOS

	// LastWriteDate (2 bytes): The date when the file was last modified.
	LastWriteDate SMB_DATE

	// FileSize (4 bytes): The size of the file, in bytes. If the file is larger than (2 ** 32 - 1) bytes in size,
	//the server SHOULD return the least-significant 32 bits of the file size.
	FileSize ULONG

	// FileName (13 bytes): The null-terminated 8.3 name format file name. The file name and extension, including the '.'
	// delimiter MUST be left-justified in the field. The character string MUST be padded with " " (space) characters, as
	// necessary, to reach 12 bytes in length. The final byte of the field MUST contain the terminating null character.
	FileName OEM_STRING
}

// NewSMB_DIRECTORY_INFORMATION creates a new SMB_DIRECTORY_INFORMATION structure
//
// Returns:
// - A pointer to the new SMB_DIRECTORY_INFORMATION structure
func NewSMB_DIRECTORY_INFORMATION() *SMB_DIRECTORY_INFORMATION {
	return &SMB_DIRECTORY_INFORMATION{
		ResumeKey:      SMB_RESUME_KEY{},
		FileAttributes: UCHAR(0),
		LastWriteTime:  SMB_TIME_DOS{},
		LastWriteDate:  SMB_DATE{},
		FileSize:       ULONG(0),
		FileName:       OEM_STRING{},
	}
}

// Marshal marshals the SMB_DIRECTORY_INFORMATION structure
//
// Returns:
// - A byte array representing the SMB_DIRECTORY_INFORMATION structure
// - An error if the marshaling fails
func (d *SMB_DIRECTORY_INFORMATION) Marshal() ([]byte, error) {
	marshalled := []byte{}

	// Marshal the SMB_Resume_Key
	resumeKeyBytes, err := d.ResumeKey.Marshal()
	if err != nil {
		return nil, err
	}
	marshalled = append(marshalled, resumeKeyBytes...)

	// Marshal the FileAttributes
	marshalled = append(marshalled, byte(d.FileAttributes))

	// Marshal the LastWriteTime
	timeBytes, err := d.LastWriteTime.Marshal()
	if err != nil {
		return nil, err
	}
	marshalled = append(marshalled, timeBytes...)

	// Marshal the LastWriteDate
	dateBytes, err := d.LastWriteDate.Marshal()
	if err != nil {
		return nil, err
	}
	marshalled = append(marshalled, dateBytes...)

	// Marshal the FileSize
	fileSizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(fileSizeBytes, uint32(d.FileSize))
	marshalled = append(marshalled, fileSizeBytes...)

	// Marshal the FileName as the fixed 13-byte field.
	//
	// [MS-CIFS] section 2.2.8.1.4: the 8.3 name is left-justified, padded with
	// spaces to 12 bytes, and the thirteenth byte is the terminating null. The
	// bytes are written directly rather than through OEM_STRING.Marshal, which
	// prefixes a buffer-format byte — correct where an SMB_STRING is expected and
	// one byte too many here, which is enough to desynchronise the whole array.
	fileName := d.FileName.GetString()
	if len(fileName) > smbDirectoryInformationNameLength {
		return nil, fmt.Errorf("file name %q is longer than the %d bytes an 8.3 name field holds",
			fileName, smbDirectoryInformationNameLength)
	}

	nameField := make([]byte, smbDirectoryInformationNameLength+1)
	copy(nameField, fileName)
	for index := len(fileName); index < smbDirectoryInformationNameLength; index++ {
		nameField[index] = ' '
	}
	// The final byte stays zero: the terminating null.
	marshalled = append(marshalled, nameField...)

	return marshalled, nil
}

// Unmarshal unmarshals the SMB_DIRECTORY_INFORMATION structure
//
// Returns:
// - An error if the unmarshaling fails
func (d *SMB_DIRECTORY_INFORMATION) Unmarshal(data []byte) (int, error) {
	offset := 0

	// Unmarshal the ResumeKey
	bytesRead, err := d.ResumeKey.Unmarshal(data[offset:])
	if err != nil {
		return 0, err
	}
	offset += bytesRead

	// Unmarshal the FileAttributes
	if offset >= len(data) {
		return offset, fmt.Errorf("data too short for FileAttributes")
	}
	d.FileAttributes = UCHAR(data[offset])
	offset++

	// Unmarshal the LastWriteTime
	if offset+2 > len(data) {
		return offset, fmt.Errorf("data too short for LastWriteTime")
	}
	bytesRead, err = d.LastWriteTime.Unmarshal(data[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshal the LastWriteDate
	if offset+2 > len(data) {
		return offset, fmt.Errorf("data too short for LastWriteDate")
	}
	bytesRead, err = d.LastWriteDate.Unmarshal(data[offset : offset+2])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshal the FileSize
	if offset+4 > len(data) {
		return offset, fmt.Errorf("data too short for FileSize")
	}
	d.FileSize = ULONG(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	// Unmarshal the FileName from the fixed 13-byte field, trimming the padding
	// and the terminator that the format requires.
	if offset+smbDirectoryInformationNameLength+1 > len(data) {
		return offset, fmt.Errorf("data too short for FileName")
	}
	nameField := data[offset : offset+smbDirectoryInformationNameLength+1]
	if terminator := bytes.IndexByte(nameField, 0x00); terminator >= 0 {
		nameField = nameField[:terminator]
	}
	d.FileName.SetString(strings.TrimRight(string(nameField), " "))
	offset += smbDirectoryInformationNameLength + 1

	return offset, nil
}

// smbDirectoryInformationNameLength is the number of name bytes in the entry's
// FileName field, which is followed by a single terminating null to make 13.
const smbDirectoryInformationNameLength = 12

// SMB_DIRECTORY_INFORMATION_SIZE is the size of the structure on the wire:
// ResumeKey(21) FileAttributes(1) LastWriteTime(2) LastWriteDate(2) FileSize(4)
// FileName(13), per [MS-CIFS] section 2.2.8.1.4.
//
// The entries are fixed-size and packed with no offset linking them, so a client
// walks the array by this stride. An entry of any other length desynchronises
// everything after it.
const SMB_DIRECTORY_INFORMATION_SIZE = 43
