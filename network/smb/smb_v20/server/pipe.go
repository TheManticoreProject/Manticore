package server

import (
	"sort"
	"strings"

	"github.com/TheManticoreProject/Manticore/logger"
	smb1 "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/server"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/server/rpcpipe"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// maxPipeAnswerSize bounds the answer taken from a pipe session in one exchange,
// the same ceiling the SMB1 server uses. A session hands over larger answers in
// parts, and the server collects each further part as the client reads.
const maxPipeAnswerSize = 64 * 1024

// pipeCommands are the commands that act on a pipe handle. Any other command
// naming one is refused, since a pipe has no file behind it.
var pipeCommands = map[codes.CommandCode]bool{
	codes.SMB2_READ:  true,
	codes.SMB2_WRITE: true,
	codes.SMB2_IOCTL: true,
	codes.SMB2_CLOSE: true,
}

// ipcShareComment is the remark Windows reports for IPC$ in a share enumeration.
const ipcShareComment = "Remote IPC"

// pipeOpen is the pipe side of an Open on IPC$: the handler session the client's
// writes reach, and the part of the session's answer the client has not read.
type pipeOpen struct {
	name    string
	session smb1.PipeSession

	// output is the rest of the current message, and more reports that the
	// session holds further parts of the answer.
	output []byte
	more   bool
}

// busy reports whether part of the previous answer is still unread. A pipe takes
// no new message until the client has collected it.
func (p *pipeOpen) busy() bool {
	return len(p.output) > 0 || p.more
}

// transact hands one message to the session and keeps its answer for the reads.
func (p *pipeOpen) transact(input []byte) nt_status.NT_STATUS {
	output, more, err := p.session.Transact(input, maxPipeAnswerSize)
	if err != nil {
		logger.Debugf("SMB2 server: pipe %q refused a message: %v", p.name, err)
		return statusForPipeError(err)
	}
	p.output, p.more = output, more && len(output) > 0
	return nt_status.NT_STATUS_SUCCESS
}

// nextMessage removes up to limit bytes of the current message, fetching the
// next part of the answer from the session once the current one is drained, and
// reports whether the message continues past them. The pipe is message mode, so
// a read never crosses from one message (one RPC PDU) into the next.
func (p *pipeOpen) nextMessage(limit int) ([]byte, bool) {
	if len(p.output) == 0 && p.more {
		output, more, err := p.session.Transact(nil, maxPipeAnswerSize)
		if err != nil {
			logger.Debugf("SMB2 server: pipe %q failed to return the rest of its answer: %v", p.name, err)
			output, more = nil, false
		}
		p.output, p.more = output, more && len(output) > 0
	}
	limit = max(min(limit, len(p.output)), 0)
	chunk := p.output[:limit:limit]
	p.output = p.output[limit:]
	return chunk, len(p.output) > 0
}

// statusForPipeError maps a pipe handler's failure onto the status a client
// expects, following the PipeHandler contract of the SMB1 server.
func statusForPipeError(err error) nt_status.NT_STATUS {
	if strings.Contains(err.Error(), "not found") {
		return nt_status.NT_STATUS_OBJECT_NAME_NOT_FOUND
	}
	return nt_status.NT_STATUS_UNSUCCESSFUL
}

// defaultPipes builds the pipe handler a server uses when its Config names none:
// srvsvc reporting the server's own shares and wkssvc reporting its names, which
// is what a client listing shares opens on IPC$.
func defaultPipes(s *Server) smb1.PipeHandler {
	return rpcpipe.New(rpcpipe.Options{
		ServerName: s.config.ServerName,
		DomainName: s.config.DomainName,
		Shares:     s.shareEntries,
	})
}

// shareEntries lists the registered shares and IPC$ for a share enumeration,
// sorted by name so the listing is stable.
func (s *Server) shareEntries() []rpcpipe.ShareEntry {
	s.mutex.RLock()
	entries := make([]rpcpipe.ShareEntry, 0, len(s.shares)+1)
	for _, share := range s.shares {
		entries = append(entries, rpcpipe.ShareEntry{
			Name: share.Name,
			Type: rpcpipe.ShareTypeOf(share.Name, smb1.ShareTypeDisk),
		})
	}
	s.mutex.RUnlock()
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToUpper(entries[i].Name) < strings.ToUpper(entries[j].Name)
	})
	return append(entries, rpcpipe.ShareEntry{
		Name:    ipcShareName,
		Comment: ipcShareComment,
		Type:    rpcpipe.ShareTypeOf(ipcShareName, smb1.ShareTypeNamedPipe),
	})
}

// createPipe answers a CREATE on IPC$ by opening a named pipe (MS-SMB2 3.3.5.9).
// A name the pipe handler does not serve is refused with
// STATUS_OBJECT_NAME_NOT_FOUND.
func (c *Connection) createPipe(r *request, resp *response, req *commands.CreateRequest) nt_status.NT_STATUS {
	if c.server.pipes == nil {
		return nt_status.NT_STATUS_OBJECT_NAME_NOT_FOUND
	}
	if len(c.opens) >= c.server.config.MaxOpensPerConnection {
		return nt_status.NT_STATUS_INSUFFICIENT_RESOURCES
	}
	name := strings.Trim(strings.ReplaceAll(req.Name, "/", `\`), `\`)
	if upper := strings.ToUpper(name); strings.HasPrefix(upper, `PIPE\`) {
		name = name[len(`PIPE\`):]
	}
	if name == "" {
		return nt_status.NT_STATUS_OBJECT_NAME_INVALID
	}

	session, err := c.server.pipes.OpenPipe(name)
	if err != nil {
		logger.Debugf("SMB2 server: %s could not open pipe %q: %v", c.remote, name, err)
		return statusForPipeError(err)
	}

	open := &Open{
		session: r.session, tree: r.tree, path: name,
		access: req.DesiredAccess, shareAccess: req.ShareAccess,
		pipe: &pipeOpen{name: name, session: session},
	}
	open.id, open.persistent = c.newFileID()
	c.opens[open.id] = open
	r.fileID = open.fileID()

	answer := commands.NewCreateResponse()
	answer.CreateAction = fileOpened
	answer.FileAttributes = fileflags.FILE_ATTRIBUTE_NORMAL
	answer.FileId = open.fileID()
	resp.command = answer
	logger.Debugf("SMB2 server: %s opened pipe %q", c.remote, name)
	return nt_status.NT_STATUS_SUCCESS
}

// readPipe answers a READ on a pipe handle with the next part of the current
// message. A read shorter than the message returns its first part with
// STATUS_BUFFER_OVERFLOW, and the client reads the rest next.
func (c *Connection) readPipe(open *Open, req *commands.ReadRequest, resp *response) nt_status.NT_STATUS {
	if req.Length > c.server.config.MaxReadSize {
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}
	chunk, continues := open.pipe.nextMessage(int(req.Length))
	answer := commands.NewReadResponse()
	answer.Data = chunk
	resp.command = answer
	if continues {
		return nt_status.NT_STATUS_BUFFER_OVERFLOW
	}
	return nt_status.NT_STATUS_SUCCESS
}

// writePipe answers a WRITE on a pipe handle by handing the data to the session
// as one message. Its answer is kept on the handle for the reads that follow.
func (c *Connection) writePipe(open *Open, req *commands.WriteRequest, resp *response) nt_status.NT_STATUS {
	if uint32(len(req.Data)) > c.server.config.MaxWriteSize {
		return nt_status.NT_STATUS_INVALID_PARAMETER
	}
	if len(req.Data) > 0 {
		if open.pipe.busy() {
			return nt_status.NT_STATUS_PIPE_BUSY
		}
		if status := open.pipe.transact(req.Data); !status.IsSuccess() {
			return status
		}
	}
	answer := commands.NewWriteResponse()
	answer.Count = uint32(len(req.Data))
	resp.command = answer
	return nt_status.NT_STATUS_SUCCESS
}

// transceivePipe answers FSCTL_PIPE_TRANSCEIVE (MS-SMB2 3.3.5.15.1): the input
// is written to the pipe and the first message of the answer is returned, cut to
// MaxOutputResponse with STATUS_BUFFER_OVERFLOW when it does not fit. The rest
// is read with READ.
func (c *Connection) transceivePipe(r *request, req *commands.IoctlRequest, resp *response) nt_status.NT_STATUS {
	open, status := c.openFor(r, req.FileId)
	if !status.IsSuccess() {
		return status
	}
	if open.pipe == nil {
		return nt_status.NT_STATUS_NOT_SUPPORTED
	}
	if open.pipe.busy() {
		return nt_status.NT_STATUS_PIPE_BUSY
	}
	if status := open.pipe.transact(req.Input); !status.IsSuccess() {
		return status
	}
	output, continues := open.pipe.nextMessage(int(req.MaxOutputResponse))

	answer := commands.NewIoctlResponse()
	answer.CtlCode = req.CtlCode
	answer.FileId = open.fileID()
	answer.Output = output
	resp.command = answer
	if continues {
		return nt_status.NT_STATUS_BUFFER_OVERFLOW
	}
	return nt_status.NT_STATUS_SUCCESS
}
