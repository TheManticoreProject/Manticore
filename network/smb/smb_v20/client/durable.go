package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/createcontext"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// DurableHandle holds the state returned from a CREATE that requested a V2
// durable handle (DH2Q). The CreateGuid and FileId are needed to reconnect
// after a transient disconnect.
type DurableHandle struct {
	FileId     types.SMB2_FILEID
	CreateGuid [16]byte
	Timeout    uint32
	Flags      uint32
}

// IsPersistent reports whether the server granted a persistent handle.
func (dh *DurableHandle) IsPersistent() bool {
	return dh.Flags&createcontext.SMB2_DHANDLE_FLAG_PERSISTENT != 0
}

// CreateFileWithDurableHandleV2 opens or creates a file with a V2 durable
// handle request (DH2Q). A random CreateGuid is generated for the open; the
// caller uses the returned DurableHandle to reconnect after a disconnect.
//
// Set persistent to true to request a persistent handle (the server must be
// continuously available and the share must support it).
func (c *Client) CreateFileWithDurableHandleV2(
	path string,
	desiredAccess, shareAccess, createDisposition, createOptions uint32,
	timeout uint32,
	persistent bool,
) (*DurableHandle, error) {
	if c.Session == nil || c.Session.TreeId == 0 {
		return nil, fmt.Errorf("no tree connect established")
	}

	var createGuid [16]byte
	copy(createGuid[:], guid.NewGUID().ToBytes())

	flags := uint32(0)
	if persistent {
		flags = createcontext.SMB2_DHANDLE_FLAG_PERSISTENT
	}

	dh2q := createcontext.MarshalDurableHandleRequestV2(timeout, flags, createGuid)
	ctxBuf, err := createcontext.Marshal([]createcontext.CreateContext{dh2q})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DH2Q create context: %w", err)
	}

	req := commands.NewCreateRequest()
	req.RequestedOplockLevel = commands.SMB2_OPLOCK_LEVEL_BATCH
	req.ImpersonationLevel = 0x00000002
	req.DesiredAccess = desiredAccess
	req.ShareAccess = shareAccess
	req.CreateDisposition = createDisposition
	req.CreateOptions = createOptions
	req.Name = path
	req.CreateContexts = ctxBuf

	response, err := c.sendReceive(c.newRequest(req), "Create(DH2Q)")
	if err != nil {
		return nil, err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		return nil, fmt.Errorf("create with durable handle V2 failed: %s", formatNTStatus(status))
	}

	createResp, ok := response.Command.(*commands.CreateResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected create response: %T", response.Command)
	}

	dh := &DurableHandle{
		FileId:     createResp.FileId,
		CreateGuid: createGuid,
	}

	if len(createResp.CreateContexts) > 0 {
		contexts, perr := createcontext.Parse(createResp.CreateContexts)
		if perr == nil {
			for _, ctx := range contexts {
				if string(ctx.Name) == string(createcontext.NameDurableHandleReqV2) {
					resp, rerr := createcontext.ParseDurableHandleResponseV2(ctx.Data)
					if rerr == nil {
						dh.Timeout = resp.Timeout
						dh.Flags = resp.Flags
					}
				}
			}
		}
	}

	return dh, nil
}

// ReconnectDurableHandleV2 reconnects a durable handle after a transient
// disconnect. The caller must have re-established the SMB2 session and tree
// connect before calling this. The original DurableHandle's FileId and
// CreateGuid identify the open to the server.
func (c *Client) ReconnectDurableHandleV2(dh *DurableHandle) (types.SMB2_FILEID, error) {
	if c.Session == nil || c.Session.TreeId == 0 {
		return types.SMB2_FILEID{}, fmt.Errorf("no tree connect established")
	}

	flags := uint32(0)
	if dh.IsPersistent() {
		flags = createcontext.SMB2_DHANDLE_FLAG_PERSISTENT
	}

	dh2c := createcontext.MarshalDurableHandleReconnectV2(
		dh.FileId.Persistent, dh.FileId.Volatile,
		dh.CreateGuid, flags,
	)
	ctxBuf, err := createcontext.Marshal([]createcontext.CreateContext{dh2c})
	if err != nil {
		return types.SMB2_FILEID{}, fmt.Errorf("failed to marshal DH2C create context: %w", err)
	}

	req := commands.NewCreateRequest()
	req.ImpersonationLevel = 0x00000002
	req.Name = ""
	req.CreateContexts = ctxBuf

	response, err := c.sendReceive(c.newRequest(req), "Create(DH2C)")
	if err != nil {
		return types.SMB2_FILEID{}, err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		return types.SMB2_FILEID{}, fmt.Errorf("durable handle V2 reconnect failed: %s", formatNTStatus(status))
	}

	createResp, ok := response.Command.(*commands.CreateResponse)
	if !ok {
		return types.SMB2_FILEID{}, fmt.Errorf("unexpected create response: %T", response.Command)
	}
	return createResp.FileId, nil
}
