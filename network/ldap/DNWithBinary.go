package ldap

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type DNWithBinary struct {
	DistinguishedName string
	BinaryData        []byte

	// Internal
	RawBytes     []byte
	RawBytesSize uint32
}

const (
	StringFormatPrefix    = "B:"
	StringFormatSeparator = ":"
)

func (d *DNWithBinary) Unmarshal(rawBytes []byte) (int, error) {
	d.RawBytes = rawBytes
	d.RawBytesSize = uint32(len(rawBytes))

	parts := bytes.Split(rawBytes, []byte(StringFormatSeparator))
	if len(parts) != 4 {
		return 0, errors.New("rawBytes should have exactly four parts separated by colons (:)")
	}

	size, err := strconv.Atoi(string(parts[1]))
	if err != nil {
		return 0, errors.New("invalid size in rawBytes")
	}

	binaryPart, err := hex.DecodeString(string(parts[2]))
	if err != nil {
		return 0, errors.New("invalid hexadecimal string in rawBytes")
	}

	if (len(binaryPart) * 2) != size {
		return 0, fmt.Errorf("invalid BinaryData length. The length specified in the header (%d) does not match the actual data length (%d)", size, len(binaryPart))
	}

	d.DistinguishedName = string(parts[3])
	d.BinaryData = binaryPart

	return len(rawBytes), nil
}

func (d *DNWithBinary) Marshal() ([]byte, error) {
	hexData := hex.EncodeToString(d.BinaryData)

	marshalledData := []byte{}
	marshalledData = append(marshalledData, []byte(StringFormatPrefix)...)
	marshalledData = append(marshalledData, []byte(fmt.Sprintf("%d", len(d.BinaryData)*2))...)
	marshalledData = append(marshalledData, []byte(StringFormatSeparator)...)
	marshalledData = append(marshalledData, []byte(hexData)...)
	marshalledData = append(marshalledData, []byte(StringFormatSeparator)...)
	marshalledData = append(marshalledData, []byte(d.DistinguishedName)...)

	return marshalledData, nil
}

func (d *DNWithBinary) String() string {
	marshalledData, err := d.Marshal()
	if err != nil {
		return ""
	}
	return string(marshalledData)
}

// Describe prints a detailed description of the DNWithBinary structure.
//
// Parameters:
// - indent: An integer value specifying the indentation level for the output.
func (d *DNWithBinary) Describe(indent int) {
	indentPrompt := strings.Repeat(" │ ", indent)
	fmt.Printf("%s<DNWithBinary structure>\n", indentPrompt)
	fmt.Printf("%s │ \x1b[93mDistinguishedName\x1b[0m: %s\n", indentPrompt, d.DistinguishedName)
	fmt.Printf("%s │ \x1b[93mBinaryData\x1b[0m: %s\n", indentPrompt, hex.EncodeToString(d.BinaryData))
	fmt.Printf("%s └───\n", indentPrompt)
}
