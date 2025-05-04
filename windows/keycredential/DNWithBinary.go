package keycredentiallink

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

func (d *DNWithBinary) Parse(rawBytes []byte) error {
	d.RawBytes = rawBytes
	d.RawBytesSize = uint32(len(rawBytes))

	parts := bytes.Split(rawBytes, []byte(StringFormatSeparator))
	if len(parts) != 4 {
		return errors.New("rawBytes should have exactly four parts separated by colons (:)")
	}

	size, err := strconv.Atoi(string(parts[1]))
	if err != nil {
		return errors.New("invalid size in rawBytes")
	}

	binaryPart, err := hex.DecodeString(string(parts[2]))
	if err != nil {
		return errors.New("invalid hexadecimal string in rawBytes")
	}

	if (len(binaryPart) * 2) != size {
		return fmt.Errorf("invalid BinaryData length. The length specified in the header (%d) does not match the actual data length (%d)", size, len(binaryPart))
	}

	d.DistinguishedName = string(parts[3])
	d.BinaryData = binaryPart

	return nil
}

func (d *DNWithBinary) ToString() string {
	hexData := hex.EncodeToString(d.BinaryData)
	return fmt.Sprintf(StringFormatPrefix+"%d"+StringFormatSeparator+"%s"+StringFormatSeparator+"%s", len(d.BinaryData)*2, hexData, d.DistinguishedName)
}

func (d *DNWithBinary) String() string {
	return d.ToString()
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
