package client

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/createcontext"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// Lease holds the server-granted lease state returned from a CREATE that
// requested a lease.
type Lease struct {
	LeaseKey       [16]byte
	LeaseState     uint32
	LeaseFlags     uint32
	Epoch          uint16
	ParentLeaseKey [16]byte
}

// LeaseBreak is a server-initiated lease break notification.
type LeaseBreak struct {
	LeaseKey          [16]byte
	CurrentLeaseState uint32
	NewLeaseState     uint32
	NewEpoch          uint16
	Flags             uint32
}

// AckRequired reports whether the server expects a Lease Break Acknowledgment
// for this notification.
func (lb *LeaseBreak) AckRequired() bool {
	return lb.Flags&commands.SMB2_NOTIFY_BREAK_LEASE_FLAG_ACK_REQUIRED != 0
}

// BreakNotification is the result of WaitBreakNotification: either an oplock
// break or a lease break, distinguished by which field is non-nil.
type BreakNotification struct {
	OplockBreak *OplockBreak
	LeaseBreak  *LeaseBreak
}

// CreateFileWithLease opens or creates a file with a V2 lease request. The
// leaseKey identifies the lease across reconnects; leaseState is a combination
// of SMB2_LEASE_READ_CACHING, SMB2_LEASE_HANDLE_CACHING, and
// SMB2_LEASE_WRITE_CACHING. parentLeaseKey is non-zero for directory leases.
//
// Returns the server-assigned FileId, the granted oplock level (always
// SMB2_OPLOCK_LEVEL_LEASE when a lease is granted), and the lease the server
// granted.
func (c *Client) CreateFileWithLease(
	path string,
	desiredAccess, shareAccess, createDisposition, createOptions uint32,
	leaseKey [16]byte,
	leaseState uint32,
	parentLeaseKey [16]byte,
	epoch uint16,
) (types.SMB2_FILEID, *Lease, error) {
	if c.Session == nil || c.Session.TreeId == 0 {
		return types.SMB2_FILEID{}, nil, fmt.Errorf("no tree connect established")
	}

	leaseCtx := createcontext.MarshalLeaseV2(leaseKey, leaseState, parentLeaseKey, epoch)
	ctxBuf, err := createcontext.Marshal([]createcontext.CreateContext{leaseCtx})
	if err != nil {
		return types.SMB2_FILEID{}, nil, fmt.Errorf("failed to marshal lease create context: %w", err)
	}

	req := commands.NewCreateRequest()
	req.RequestedOplockLevel = commands.SMB2_OPLOCK_LEVEL_LEASE
	req.ImpersonationLevel = 0x00000002
	req.DesiredAccess = desiredAccess
	req.ShareAccess = shareAccess
	req.CreateDisposition = createDisposition
	req.CreateOptions = createOptions
	req.Name = path
	req.CreateContexts = ctxBuf

	response, err := c.sendReceive(c.newRequest(req), "Create(lease)")
	if err != nil {
		return types.SMB2_FILEID{}, nil, err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		return types.SMB2_FILEID{}, nil, fmt.Errorf("create with lease failed: %s", formatNTStatus(status))
	}

	createResp, ok := response.Command.(*commands.CreateResponse)
	if !ok {
		return types.SMB2_FILEID{}, nil, fmt.Errorf("unexpected create response: %T", response.Command)
	}

	lease := &Lease{LeaseKey: leaseKey}
	if len(createResp.CreateContexts) > 0 {
		contexts, err := createcontext.Parse(createResp.CreateContexts)
		if err == nil {
			for _, ctx := range contexts {
				if string(ctx.Name) == string(createcontext.NameRequestLease) {
					parsed, perr := createcontext.ParseLeaseResponse(ctx.Data)
					if perr == nil {
						switch v := parsed.(type) {
						case *createcontext.LeaseV2:
							lease.LeaseState = v.LeaseState
							lease.LeaseFlags = v.LeaseFlags
							lease.Epoch = v.Epoch
							lease.ParentLeaseKey = v.ParentLeaseKey
						case *createcontext.LeaseV1:
							lease.LeaseState = v.LeaseState
							lease.LeaseFlags = v.LeaseFlags
						}
					}
				}
			}
		}
	}

	return createResp.FileId, lease, nil
}

// AcknowledgeLeaseBreak replies to a lease break notification, accepting the
// new lease state. Wire: SMB2 Lease Break Acknowledgment (MS-SMB2 2.2.24.2).
func (c *Client) AcknowledgeLeaseBreak(leaseKey [16]byte, leaseState uint32) error {
	if c.Session == nil || c.Session.TreeId == 0 {
		return fmt.Errorf("no tree connect established")
	}

	req := commands.NewLeaseBreakAckRequest()
	req.LeaseKey = leaseKey
	req.LeaseState = leaseState

	response, err := c.sendReceive(c.newRequest(req), "LeaseBreakAck")
	if err != nil {
		return err
	}
	if status := statusFromResponse(response); status != 0x00000000 {
		return fmt.Errorf("lease break acknowledgment failed: %s", formatNTStatus(status))
	}
	return nil
}

// WaitBreakNotification reads the next unsolicited break notification from the
// server — either an oplock break (StructureSize 24) or a lease break
// (StructureSize 44). It blocks until a notification arrives and must not run
// concurrently with another operation that reads the connection.
func (c *Client) WaitBreakNotification() (*BreakNotification, error) {
	if !c.Transport.IsConnected() {
		return nil, fmt.Errorf("transport is not connected")
	}

	raw, err := c.Transport.Receive()
	if err != nil {
		return nil, fmt.Errorf("failed to receive break notification: %w", err)
	}

	// When per-session encryption is active, the server wraps unsolicited
	// notifications in an SMB2 TRANSFORM_HEADER (MS-SMB2 3.2.5.19).
	wasEncrypted := false
	if isTransformHeader(raw) {
		plaintext, derr := c.decryptMessage(raw)
		if derr != nil {
			return nil, fmt.Errorf("break notification: %w", derr)
		}
		raw = plaintext
		wasEncrypted = true
	}

	if c.Session != nil && c.Session.EncryptData && !wasEncrypted {
		return nil, fmt.Errorf("break notification: session requires encryption but notification was not encrypted")
	}

	msg := message.NewMessage()
	if _, err := msg.Header.Unmarshal(raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal break notification header: %w", err)
	}
	if !msg.Header.HasValidProtocolId() {
		return nil, fmt.Errorf("break notification is not an SMB2 message (ProtocolId % x)", msg.Header.ProtocolId)
	}
	if !wasEncrypted && c.Session != nil && c.Session.SigningActive && signatureRequired(msg) {
		if !verifySignatureForDialect(c.Connection.Dialect, c.Connection.SigningAlgorithmId, c.Session.SigningKey, raw) {
			return nil, fmt.Errorf("break notification failed SMB2 signature verification")
		}
	}
	if msg.Header.Command != codes.SMB2_OPLOCK_BREAK {
		return nil, fmt.Errorf("expected a break notification, got command 0x%04x", uint16(msg.Header.Command))
	}

	body := raw[header.SMB2_HEADER_SIZE:]
	if len(body) < 2 {
		return nil, fmt.Errorf("break notification body too short")
	}
	structureSize := binary.LittleEndian.Uint16(body[0:2])

	switch structureSize {
	case commands.LeaseBreakNotificationStructureSize:
		notify := commands.NewLeaseBreakNotification()
		if _, err := notify.Unmarshal(body); err != nil {
			return nil, fmt.Errorf("failed to unmarshal lease break notification: %w", err)
		}
		return &BreakNotification{
			LeaseBreak: &LeaseBreak{
				LeaseKey:          notify.LeaseKey,
				CurrentLeaseState: notify.CurrentLeaseState,
				NewLeaseState:     notify.NewLeaseState,
				NewEpoch:          notify.NewEpoch,
				Flags:             notify.Flags,
			},
		}, nil

	case commands.OplockBreakResponseStructureSize:
		notify := commands.NewOplockBreakResponse()
		if _, err := notify.Unmarshal(body); err != nil {
			return nil, fmt.Errorf("failed to unmarshal oplock break notification: %w", err)
		}
		return &BreakNotification{
			OplockBreak: &OplockBreak{
				FileId:   notify.FileId,
				NewLevel: uint8(notify.OplockLevel),
			},
		}, nil

	default:
		return nil, fmt.Errorf("unrecognised break notification StructureSize %d", structureSize)
	}
}
