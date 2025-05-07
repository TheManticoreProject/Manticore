package commands

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
)

// Transaction2Response
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/216e606a-eee1-4c3f-b88e-0eb14dc380b2
// The SMB_COM_TRANSACTION2 response has two possible formats.
// The standard format is used to return the results of the completed transaction.
// A shortened interim response message is sent following the initial SMB_COM_TRANSACTION2
// request if secondary request messages (SMB_COM_TRANSACTION2_SECONDARY) are pending.
// Whenever a transaction request is split across multiple SMB requests, the server MUST
// evaluate the initial SMB_COM_TRANSACTION2 request to determine whether or not it has
// the resources necessary to process the transaction. It MUST also check for any other
// errors it can detect based upon the initial request, and then send back an interim
// response. The interim response advises the client as to whether it can send the rest
// of the transaction to the server.
// TODO: Implement the two possible formats
type Transaction2Response struct {
	command_interface.Command
}

// NewTransaction2Response creates a new Transaction2Response structure
//
// Returns:
// - A pointer to the new Transaction2Response structure
func NewTransaction2Response() *Transaction2Response {
	c := &Transaction2Response{}

	c.Command.SetCommandCode(codes.SMB_COM_TRANSACTION2)

	return c
}

// Marshal marshals the Transaction2Response structure into a byte array
//
// Returns:
// - A byte array representing the Transaction2Response structure
// - An error if the marshaling fails
func (c *Transaction2Response) Marshal() ([]byte, error) {
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

	// Then marshal the parameters
	rawParametersContent := []byte{}

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
func (c *Transaction2Response) Unmarshal(data []byte) (int, error) {
	offset := 0

	// First unmarshal the two structures
	bytesRead, err := c.GetParameters().Unmarshal(data)
	if err != nil {
		return 0, err
	}
	_ = c.GetParameters().GetBytes()
	bytesRead, err = c.GetData().Unmarshal(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	_ = c.GetData().GetBytes()

	// First unmarshal the parameters
	offset = 0
	// No parameters are sent in this message

	// Then unmarshal the data
	offset = 0
	// No data is sent in this message

	return offset, nil
}
