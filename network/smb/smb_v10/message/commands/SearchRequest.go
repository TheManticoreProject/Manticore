package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// SearchRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/239b0def-8370-4dc7-8391-ee60952901b1
type SearchRequest struct {
	command_interface.Command

	// Parameters

	// MaxCount (2 bytes): The maximum number of directory entries to return. This
	// value represents the maximum number of entries across the entirety of the
	// search, not just the initial response.
	MaxCount types.USHORT

	// SearchAttributes (2 bytes): An attribute mask used to specify the standard
	// attributes a file MUST have in order to match the search. If the value of this
	// field is 0x0000, then only normal files are returned. If the Volume Label
	// attribute is set, the server MUST return only the volume label (the Volume Label
	// attribute is exclusive). If the Directory, System, or Hidden attributes are
	// specified, then those entries are returned in addition to the normal files.
	// Exclusive search attributes (see section 2.2.1.2.4) can also be set.
	SearchAttributes types.SMB_FILE_ATTRIBUTES

	// Data

	// BufferFormat1 (1 byte): This field MUST be 0x04, which indicates that a
	// null-terminated SMB_STRING is to follow.
	// FileName (variable): A null-terminated SMB_STRING. This is the full directory
	// path (relative to the TID) of the file(s) being sought. Only the final component
	// of the path MAY contain wildcards. This string MAY be the empty string.
	FileName types.SMB_STRING

	// BufferFormat2 (1 byte): This field MUST be 0x05, which indicates a variable
	// block is to follow.
	// ResumeKeyLength (2 bytes): This field MUST be either 0x0000 or 21 (0x0015). If
	// the value of this field is 0x0000, this is an initial search request. The server
	// MUST allocate resources to maintain search state so that subsequent requests MAY
	// be processed. If the value of this field is 21 (0x0015), this request MUST be
	// the continuation of a previous search, and the next field MUST contain a
	// ResumeKey previously returned by the server.
	// ResumeKey (variable): SMB_Resume_Key If the value of ResumeKeyLength is 21
	// (0x0015), this field MUST contain a ResumeKey returned by the server in response
	// to a previous SMB_COM_SEARCH request. The ResumeKey contains data used by both
	// the client and the server to maintain the state of the search. The structure of the
	// ResumeKey follows:
	ResumeKey types.SMB_RESUME_KEY
}

// NewSearchRequest creates a new SearchRequest structure
//
// Returns:
// - A pointer to the new SearchRequest structure
func NewSearchRequest() *SearchRequest {
	c := &SearchRequest{
		// Parameters
		MaxCount:         types.USHORT(0),
		SearchAttributes: types.SMB_FILE_ATTRIBUTES{},

		// Data
		FileName:  types.SMB_STRING{},
		ResumeKey: types.SMB_RESUME_KEY{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_SEARCH)

	return c
}

// Marshal marshals the SearchRequest structure into a byte array
//
// Returns:
// - A byte array representing the SearchRequest structure
// - An error if the marshaling fails
func (c *SearchRequest) Marshal() ([]byte, error) {
	marshalledCommand := []byte{}

	// Create the Parameters structure if it is nil
	if c.GetParameters() == nil {
		c.SetParameters(parameters.NewParameters())
	}
	// Create the Data structure if it is nil
	if c.GetData() == nil {
		c.SetData(data.NewData())
	}

	// In case of AndX, we need to add the parameters to the Parameters structure first
	if c.IsAndX() {
		if c.GetAndX() == nil {
			c.SetAndX(andx.NewAndX())
			c.GetAndX().AndXCommand = c.GetCommandCode()
		}

		for _, parameter := range c.GetAndX().GetParameters() {
			c.GetParameters().AddWord(parameter)
		}
	}

	// First marshal the data and then the parameters
	// This is because some parameters are dependent on the data, for example the size of some fields within
	// the data will be stored in the parameters
	rawDataContent := []byte{}

	// Marshalling data FileName
	c.FileName.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	bytesStream, err := c.FileName.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Marshalling data ResumeKey
	bytesStream, err = c.ResumeKey.Marshal()
	if err != nil {
		return nil, err
	}
	rawDataContent = append(rawDataContent, bytesStream...)

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter MaxCount
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.MaxCount))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter SearchAttributes
	bytesStream, err = c.SearchAttributes.Marshal()
	if err != nil {
		return nil, err
	}
	rawParametersContent = append(rawParametersContent, bytesStream...)

	// Marshalling parameters
	c.GetParameters().AddWordsFromBytesStream(rawParametersContent)
	marshalledParameters, err := c.GetParameters().Marshal()
	if err != nil {
		return nil, err
	}
	marshalledCommand = append(marshalledCommand, marshalledParameters...)

	// Marshalling data
	c.GetData().Add(rawDataContent)
	marshalledData, err := c.GetData().Marshal()
	if err != nil {
		return nil, err
	}
	marshalledCommand = append(marshalledCommand, marshalledData...)

	return marshalledCommand, nil
}

// Unmarshal unmarshals a byte array into the command structure
//
// Parameters:
// - data: The byte array to unmarshal
//
// Returns:
// - The number of bytes unmarshalled
func (c *SearchRequest) Unmarshal(data []byte) (int, error) {
	offset := 0

	// First unmarshal the two structures
	bytesRead, err := c.GetParameters().Unmarshal(data)
	if err != nil {
		return 0, err
	}
	rawParametersContent := c.GetParameters().GetBytes()
	_, err = c.GetData().Unmarshal(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	rawDataContent := c.GetData().GetBytes()

	// If the parameters and data are empty, this is a response containing an error code in
	// the SMB Header Status field
	if len(rawParametersContent) == 0 && len(rawDataContent) == 0 {
		return 0, nil
	}

	// First unmarshal the parameters
	offset = 0

	// Unmarshalling parameter MaxCount
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for MaxCount")
	}
	c.MaxCount = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter SearchAttributes
	bytesRead, err = c.SearchAttributes.Unmarshal(rawParametersContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Then unmarshal the data
	offset = 0

	// Unmarshalling data FileName
	bytesRead, err = c.FileName.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	// Unmarshalling data ResumeKey
	bytesRead, err = c.ResumeKey.Unmarshal(rawDataContent[offset:])
	if err != nil {
		return offset, err
	}
	offset += bytesRead

	return offset, nil
}
