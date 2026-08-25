package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
	"github.com/TheManticoreProject/winacl/securitydescriptor"
)

// echoPipe is a PipeHandler that returns whatever is written to it, which is the
// smallest handler that exercises the whole path a request-response pipe takes.
type echoPipe struct {
	mutex sync.Mutex

	// names are the pipes this handler serves; anything else is unknown.
	names map[string]bool

	// opened and closed record the calls, so a test can assert the lifecycle.
	opened []string
	closed []string

	// prefix is prepended to every answer, so a test can tell the handler's
	// output from its input.
	prefix string

	// truncateAt, when positive, caps an answer and reports that more remains.
	truncateAt int
}

func newEchoPipe(names ...string) *echoPipe {
	served := map[string]bool{}
	for _, name := range names {
		served[strings.ToLower(name)] = true
	}
	return &echoPipe{names: served, prefix: "echo:"}
}

func (p *echoPipe) OpenPipe(name string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if !p.names[strings.ToLower(name)] {
		return fmt.Errorf("pipe %q not found", name)
	}
	p.opened = append(p.opened, name)
	return nil
}

func (p *echoPipe) Transact(name string, input []byte, maxOutput int) ([]byte, bool, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if !p.names[strings.ToLower(name)] {
		return nil, false, fmt.Errorf("pipe %q not found", name)
	}

	answer := append([]byte(p.prefix), input...)
	if p.truncateAt > 0 && len(answer) > p.truncateAt {
		return answer[:p.truncateAt], true, nil
	}
	if len(answer) > maxOutput {
		return answer[:maxOutput], true, nil
	}
	return answer, false, nil
}

func (p *echoPipe) ClosePipe(name string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.closed = append(p.closed, name)
	return nil
}

// openedNames and closedNames report the lifecycle calls the handler saw.
func (p *echoPipe) openedNames() []string {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return append([]string(nil), p.opened...)
}

func (p *echoPipe) closedNames() []string {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return append([]string(nil), p.closed...)
}

// pipeServer stands up a server with an IPC share served by the given handler, and
// returns a client connected to it.
func pipeServer(t *testing.T, pipes PipeHandler) (*Server, *smb1client.Client) {
	t.Helper()

	srv, transportEnd := pipedServer(t, conformanceConfig(SigningDisabled))
	if err := srv.AddShare(&Share{Name: "IPC$", Type: ShareTypeNamedPipe, Pipes: pipes}); err != nil {
		t.Fatalf("AddShare() error = %v", err)
	}

	client := smb1client.NewFromTransport(transportEnd, net.IPv4(127, 0, 0, 1), 445)
	if err := client.Negotiate(); err != nil {
		t.Fatalf("Negotiate() error = %v", err)
	}
	creds, err := credentials.NewCredentials(captureDomain, captureUsername, capturePassword, "")
	if err != nil {
		t.Fatalf("NewCredentials() error = %v", err)
	}
	if err := client.SessionSetup(creds); err != nil {
		t.Fatalf("SessionSetup() error = %v", err)
	}
	if err := client.TreeConnect("IPC$"); err != nil {
		t.Fatalf("TreeConnect() error = %v", err)
	}

	return srv, client
}

// openPipeHandle opens a pipe over SMB and returns the FID the server assigned.
//
// This is the first half of the exchange the phase exists for: the FID is what a
// transaction names, per [MS-CIFS] section 3.3.5.57.7.
func openPipeHandle(t *testing.T, client *smb1client.Client, pipeName string) (smb1client.FID, error) {
	t.Helper()

	fid, err := client.OpenFile(pipeName, fileflags.GENERIC_READ|fileflags.GENERIC_WRITE,
		fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE,
		fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		return 0, err
	}
	return fid, nil
}

// transactPipe sends a TRANS_TRANSACT_NMPIPE against an open pipe handle and
// returns the response's data block and status.
//
// The pipe is named by the FID in Setup[1]; the Name field carries "\PIPE\" as
// the subcommand specifies, which is exactly what a real client sends.
func transactPipe(t *testing.T, client *smb1client.Client, fid smb1client.FID, payload []byte) ([]byte, uint32) {
	t.Helper()
	return transactPipeNamed(t, client, fid, `\PIPE\`, payload)
}

// transactPipeNamed is transactPipe with control over the Name field, so a test
// can send a request that names the pipe rather than passing a handle.
func transactPipeNamed(t *testing.T, client *smb1client.Client, fid smb1client.FID, name string, payload []byte) ([]byte, uint32) {
	t.Helper()

	setup := []types.USHORT{types.USHORT(subcommands.TRANS_TRANSACT_NMPIPE)}
	if fid != 0 {
		setup = append(setup, types.USHORT(fid))
	}

	response, status := sendTransaction(t, client, setup, name, nil, payload)
	if status != 0 {
		return nil, status
	}
	return []byte(response.Trans_Data), 0
}

// sendTransaction sends one SMB_COM_TRANSACTION and returns the decoded response,
// or the status it was refused with.
func sendTransaction(
	t *testing.T,
	client *smb1client.Client,
	setup []types.USHORT,
	name string,
	parameters, data []byte,
) (*commands.TransactionResponse, uint32) {
	t.Helper()

	request := newRequest(codes.SMB_COM_TRANSACTION)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID
	// The name is sent OEM, so the server must read it that way.
	request.Header.Flags2 &^= flags2.Flags2(flags2.FLAGS2_UNICODE)

	transaction := commands.NewTransactionRequest()
	transaction.Setup = setup
	transaction.SetupCount = types.UCHAR(len(setup))
	transaction.MaxParameterCount = 1024
	transaction.MaxDataCount = 4096
	if err := transaction.Name.SetString(name); err != nil {
		t.Fatalf("Name.SetString() error = %v", err)
	}
	transaction.TotalParameterCount = types.USHORT(len(parameters))
	transaction.ParameterCount = types.USHORT(len(parameters))
	transaction.Trans_Parameters = parameters
	transaction.TotalDataCount = types.USHORT(len(data))
	transaction.DataCount = types.USHORT(len(data))
	transaction.Trans_Data = data
	request.AddCommand(transaction)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the transaction: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}
	response := message.NewMessage()
	if err := response.Unmarshal(raw); err != nil {
		t.Fatalf("failed to decode the response: %v", err)
	}
	if response.Header.Status != 0 && response.Header.Status != uint32(nt_status.NT_STATUS_BUFFER_OVERFLOW) {
		return nil, response.Header.Status
	}

	transactionResponse, ok := response.Command.(*commands.TransactionResponse)
	if !ok {
		t.Fatalf("response command is %T", response.Command)
	}
	return transactionResponse, response.Header.Status
}

// TestNamedPipeTransact is what this phase exists for: a request-response pipe
// reached over SMB, which is the path MS-RPC travels.
//
// It is the whole exchange, in the order a real client performs it: open the pipe
// on IPC$, transact on the handle, close it.
func TestNamedPipeTransact(t *testing.T) {
	pipes := newEchoPipe("srvsvc", "lsarpc")
	_, client := pipeServer(t, pipes)

	fid, err := openPipeHandle(t, client, `\PIPE\srvsvc`)
	if err != nil {
		t.Fatalf("opening the pipe failed: %v", err)
	}

	payload := []byte("an rpc bind would go here")
	answer, status := transactPipe(t, client, fid, payload)
	if status != 0 {
		t.Fatalf("the pipe transaction answered 0x%08X, want success", status)
	}

	want := append([]byte("echo:"), payload...)
	if !bytes.Equal(answer, want) {
		t.Fatalf("the pipe returned %q, want %q", answer, want)
	}

	if err := client.CloseFile(fid); err != nil {
		t.Fatalf("closing the pipe failed: %v", err)
	}

	// The handler's lifecycle calls are what prove the handle was a pipe handle
	// and not something the file path happened to accept.
	if got := pipes.openedNames(); len(got) != 1 || got[0] != "srvsvc" {
		t.Errorf("the handler was asked to open %v, want [srvsvc]", got)
	}
	if got := pipes.closedNames(); len(got) != 1 || got[0] != "srvsvc" {
		t.Errorf("the handler was asked to close %v, want [srvsvc]", got)
	}
}

// TestNamedPipeNameForms asserts the forms a client names a pipe in all reach the
// same handler, since clients differ in how much of the path they send.
func TestNamedPipeNameForms(t *testing.T) {
	for _, form := range []string{`\PIPE\srvsvc`, `\pipe\srvsvc`, `\srvsvc`, "srvsvc", "/PIPE/srvsvc"} {
		form := form
		t.Run(form, func(t *testing.T) {
			pipes := newEchoPipe("srvsvc")
			_, client := pipeServer(t, pipes)

			fid, err := openPipeHandle(t, client, form)
			if err != nil {
				t.Fatalf("naming the pipe %q failed to open: %v", form, err)
			}
			answer, status := transactPipe(t, client, fid, []byte("x"))
			if status != 0 {
				t.Fatalf("naming the pipe %q answered 0x%08X", form, status)
			}
			if !bytes.Equal(answer, []byte("echo:x")) {
				t.Fatalf("naming the pipe %q returned %q", form, answer)
			}
		})
	}
}

// TestNamedPipeUnknownIsRefused asserts a pipe the handler does not serve is
// refused, and that a share with no handler refuses every pipe operation.
func TestNamedPipeUnknownIsRefused(t *testing.T) {
	t.Run("unknown pipe", func(t *testing.T) {
		_, client := pipeServer(t, newEchoPipe("srvsvc"))
		if _, err := openPipeHandle(t, client, `\PIPE\nosuchpipe`); err == nil {
			t.Fatal("a pipe the handler does not serve was opened")
		}
	})

	t.Run("share with no handler", func(t *testing.T) {
		srv, transportEnd := pipedServer(t, conformanceConfig(SigningDisabled))
		if err := srv.AddShare(&Share{Name: "IPC$", Type: ShareTypeNamedPipe}); err != nil {
			t.Fatalf("AddShare() error = %v", err)
		}
		client := smb1client.NewFromTransport(transportEnd, net.IPv4(127, 0, 0, 1), 445)
		if err := client.Negotiate(); err != nil {
			t.Fatalf("Negotiate() error = %v", err)
		}
		creds, _ := credentials.NewCredentials(captureDomain, captureUsername, capturePassword, "")
		if err := client.SessionSetup(creds); err != nil {
			t.Fatalf("SessionSetup() error = %v", err)
		}
		if err := client.TreeConnect("IPC$"); err != nil {
			t.Fatalf("TreeConnect() error = %v", err)
		}
		if _, err := openPipeHandle(t, client, `\PIPE\srvsvc`); err == nil {
			t.Fatal("a share with no pipe handler opened a pipe")
		}
	})
}

// TestNamedPipeTransactRequiresAPipeHandle asserts a transaction on something that
// is not an open pipe is refused rather than guessed at.
//
// The FID is what identifies the pipe, so a request that carries the wrong one
// must fail: answering it from the Name field instead would let a client act on a
// pipe it never opened.
func TestNamedPipeTransactRequiresAPipeHandle(t *testing.T) {
	t.Run("a FID that is not open", func(t *testing.T) {
		_, client := pipeServer(t, newEchoPipe("srvsvc"))
		if _, status := transactPipe(t, client, 0x0999, []byte("x")); status == 0 {
			t.Fatal("a transaction on an unopened FID was answered")
		}
	})

	t.Run("a FID that is a file", func(t *testing.T) {
		fs := NewMemoryFileSystem("FILES")
		if err := fs.AddFile("notapipe.txt", []byte("x")); err != nil {
			t.Fatalf("AddFile() error = %v", err)
		}
		srv, transportEnd := pipedServer(t, conformanceConfig(SigningDisabled))
		if err := srv.AddShare(&Share{Name: fileShareName, Type: ShareTypeDisk, FS: fs}); err != nil {
			t.Fatalf("AddShare() error = %v", err)
		}
		if err := srv.AddShare(&Share{Name: "IPC$", Type: ShareTypeNamedPipe, Pipes: newEchoPipe("srvsvc")}); err != nil {
			t.Fatalf("AddShare() error = %v", err)
		}

		client := smb1client.NewFromTransport(transportEnd, net.IPv4(127, 0, 0, 1), 445)
		if err := client.Negotiate(); err != nil {
			t.Fatalf("Negotiate() error = %v", err)
		}
		creds, _ := credentials.NewCredentials(captureDomain, captureUsername, capturePassword, "")
		if err := client.SessionSetup(creds); err != nil {
			t.Fatalf("SessionSetup() error = %v", err)
		}
		if err := client.TreeConnect(fileShareName); err != nil {
			t.Fatalf("TreeConnect() error = %v", err)
		}
		fileFID, err := client.OpenFile("notapipe.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ,
			fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE)
		if err != nil {
			t.Fatalf("OpenFile() error = %v", err)
		}
		fileTID := client.Session.TreeID

		// Move onto the pipe share, but hand the transaction the file's handle.
		if err := client.TreeConnect("IPC$"); err != nil {
			t.Fatalf("TreeConnect(IPC$) error = %v", err)
		}
		if client.Session.TreeID == fileTID {
			t.Fatal("the second tree connect returned the same TID")
		}
		if _, status := transactPipe(t, client, fileFID, []byte("x")); status == 0 {
			t.Fatal("a transaction on a file handle was answered as a pipe")
		}
	})
}

// TestNamedPipeOnDiskShareIsRefused asserts a pipe operation on a disk share is
// refused, rather than being attempted against a file system.
func TestNamedPipeOnDiskShareIsRefused(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	_, client := fileServer(t, fs, false)

	if _, status := transactPipeNamed(t, client, 0, `\PIPE\srvsvc`, []byte("x")); status == 0 {
		t.Fatal("a pipe transaction on a disk share was answered")
	}
}

// TestNamedPipeTruncatedAnswer asserts an answer larger than the client's budget is
// cut to fit AND reported with STATUS_BUFFER_OVERFLOW.
//
// The status is the assertion that matters. [MS-CIFS] section 3.3.5.57.7 requires
// the response to be constructed on overflow as well as on success, and the
// overflow status is the only thing telling the client its answer is incomplete —
// reporting plain success would leave an RPC client parsing a truncated response
// as a whole one.
func TestNamedPipeTruncatedAnswer(t *testing.T) {
	pipes := newEchoPipe("srvsvc")
	pipes.truncateAt = 8
	_, client := pipeServer(t, pipes)

	fid, err := openPipeHandle(t, client, `\PIPE\srvsvc`)
	if err != nil {
		t.Fatalf("opening the pipe failed: %v", err)
	}

	setup := []types.USHORT{types.USHORT(subcommands.TRANS_TRANSACT_NMPIPE), types.USHORT(fid)}
	response, status := sendTransaction(t, client, setup, `\PIPE\`, nil, bytes.Repeat([]byte("A"), 64))
	if status != uint32(nt_status.NT_STATUS_BUFFER_OVERFLOW) {
		t.Fatalf("a truncated answer reported 0x%08X, want STATUS_BUFFER_OVERFLOW (0x%08X)",
			status, uint32(nt_status.NT_STATUS_BUFFER_OVERFLOW))
	}
	if response == nil {
		t.Fatal("no response accompanied the overflow status")
	}
	if len(response.Trans_Data) != 8 {
		t.Fatalf("the answer is %d bytes, want the 8 the handler capped it at", len(response.Trans_Data))
	}
}

// TestPipeStateQuery asserts a state query is answered, and that it reports a
// message-mode pipe.
//
// The mode is not cosmetic: [MS-CIFS] section 3.3.5.57.7 requires a transacted
// exchange to be refused on a pipe that is not message mode, so a server that
// reported byte mode here would be contradicting the transactions it goes on to
// answer.
func TestPipeStateQuery(t *testing.T) {
	_, client := pipeServer(t, newEchoPipe("srvsvc"))

	fid, err := openPipeHandle(t, client, `\PIPE\srvsvc`)
	if err != nil {
		t.Fatalf("opening the pipe failed: %v", err)
	}

	setup := []types.USHORT{types.USHORT(subcommands.TRANS_QUERY_NMPIPE_STATE), types.USHORT(fid)}
	response, status := sendTransaction(t, client, setup, `\PIPE\`, nil, nil)
	if status != 0 {
		t.Fatalf("the state query answered 0x%08X", status)
	}
	if len(response.Trans_Parameters) < 2 {
		t.Fatalf("the state query returned %d parameter bytes, want 2", len(response.Trans_Parameters))
	}

	state := binary.LittleEndian.Uint16([]byte(response.Trans_Parameters))
	if state != pipeStateMessageMode {
		t.Fatalf("the reported pipe state is 0x%04X, want 0x%04X", state, pipeStateMessageMode)
	}
	// The bits, spelled out: read mode and pipe type both message.
	if readMode := (state >> 8) & 0x03; readMode != uint16(types.SMB_NMPIPE_STATUS_READ_MODE_MESSAGE) {
		t.Errorf("the reported read mode is %d, want message mode", readMode)
	}
}

// TestSecurityDescriptorQuery asserts a descriptor is returned and that it parses
// back through winacl into what the provider meant, which is what a client will do
// with it.
func TestSecurityDescriptorQuery(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("described.txt", []byte("x")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}

	srv, transportEnd := pipedServer(t, conformanceConfig(SigningDisabled))
	if err := srv.AddShare(&Share{
		Name: fileShareName, Type: ShareTypeDisk, FS: fs,
		Security: NewReflectiveSecurityProvider(false),
	}); err != nil {
		t.Fatalf("AddShare() error = %v", err)
	}

	client := smb1client.NewFromTransport(transportEnd, net.IPv4(127, 0, 0, 1), 445)
	if err := client.Negotiate(); err != nil {
		t.Fatalf("Negotiate() error = %v", err)
	}
	creds, _ := credentials.NewCredentials(captureDomain, captureUsername, capturePassword, "")
	if err := client.SessionSetup(creds); err != nil {
		t.Fatalf("SessionSetup() error = %v", err)
	}
	if err := client.TreeConnect(fileShareName); err != nil {
		t.Fatalf("TreeConnect() error = %v", err)
	}

	fid, err := client.OpenFile("described.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer client.CloseFile(fid)

	descriptor, status := querySecurityDescriptor(t, client, uint16(fid),
		OwnerSecurityInformation|GroupSecurityInformation|DaclSecurityInformation)
	if status != 0 {
		t.Fatalf("the security-descriptor query answered 0x%08X", status)
	}
	if len(descriptor) == 0 {
		t.Fatal("the query returned an empty descriptor")
	}

	// The proof: winacl parses back what the provider built.
	parsed := &securitydescriptor.NtSecurityDescriptor{}
	if _, err := parsed.Unmarshal(descriptor); err != nil {
		t.Fatalf("the descriptor does not parse: %v", err)
	}
	if parsed.Owner == nil || parsed.Owner.SID.ToString() != sidLocalSystem {
		t.Errorf("the owner is %v, want %s", parsed.Owner, sidLocalSystem)
	}
	if parsed.DACL == nil || len(parsed.DACL.Entries) != 1 {
		t.Fatalf("the DACL has %v entries, want 1", parsed.DACL)
	}
	entry := parsed.DACL.Entries[0]
	if entry.Identity.SID.ToString() != sidAuthenticatedUsers {
		t.Errorf("the DACL grants %s, want %s", entry.Identity.SID.ToString(), sidAuthenticatedUsers)
	}
	// A writable share describes write access, because that is what it enforces.
	if entry.Mask.RawValue&accessWrite == 0 {
		t.Errorf("the DACL mask is 0x%08X, which does not include write on a writable share", entry.Mask.RawValue)
	}
}

// TestSecurityDescriptorReflectsReadOnly asserts the descriptor describes what the
// server actually enforces: a read-only share does not advertise write access.
//
// That is the whole justification for deriving a descriptor rather than refusing:
// a client uses one to predict what it will be allowed to do, so a descriptor that
// disagreed with the handlers would make the client wrong.
func TestSecurityDescriptorReflectsReadOnly(t *testing.T) {
	for _, readOnly := range []bool{false, true} {
		readOnly := readOnly
		t.Run(fmt.Sprintf("readOnly=%t", readOnly), func(t *testing.T) {
			provider := NewReflectiveSecurityProvider(readOnly)
			descriptor, err := provider.SecurityDescriptor("anything", DaclSecurityInformation)
			if err != nil {
				t.Fatalf("SecurityDescriptor() error = %v", err)
			}

			parsed := &securitydescriptor.NtSecurityDescriptor{}
			if _, err := parsed.Unmarshal(descriptor); err != nil {
				t.Fatalf("the descriptor does not parse: %v", err)
			}
			if parsed.DACL == nil || len(parsed.DACL.Entries) != 1 {
				t.Fatal("the descriptor has no single-entry DACL")
			}

			mask := parsed.DACL.Entries[0].Mask.RawValue
			if mask&accessRead == 0 {
				t.Error("the descriptor does not describe read access, which every share grants")
			}
			if readOnly && mask&accessWrite != 0 {
				t.Error("a read-only share's descriptor describes write access")
			}
			if !readOnly && mask&accessWrite == 0 {
				t.Error("a writable share's descriptor does not describe write access")
			}
		})
	}
}

// TestSecurityDescriptorWithoutProviderIsRefused asserts a share with no security
// model says so rather than inventing a descriptor.
func TestSecurityDescriptorWithoutProviderIsRefused(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("plain.txt", []byte("x")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	fid, err := client.OpenFile("plain.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer client.CloseFile(fid)

	if _, status := querySecurityDescriptor(t, client, uint16(fid), DaclSecurityInformation); status == 0 {
		t.Fatal("a share with no security provider returned a descriptor")
	}
}

// TestReflectiveProviderRefusesSet asserts the derived provider refuses a change
// rather than accepting one it could not store: a client that believed it had
// applied a descriptor would be wrong about what the server enforces.
func TestReflectiveProviderRefusesSet(t *testing.T) {
	provider := NewReflectiveSecurityProvider(false)
	if err := provider.SetSecurityDescriptor("anything", DaclSecurityInformation, []byte{0x01}); err == nil {
		t.Fatal("the derived provider accepted a descriptor change")
	}
}

// TestNtTransactUnimplementedFunctions asserts the NT_TRANSACT functions this
// server does not serve are refused, rather than answered with something invented.
func TestNtTransactUnimplementedFunctions(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	_, client := fileServer(t, fs, false)

	unimplemented := []subcommands.NtTransactSubcommand{
		subcommands.NT_TRANSACT_CREATE,
		subcommands.NT_TRANSACT_RENAME,
		subcommands.NT_TRANSACT_NOTIFY_CHANGE,
		subcommands.NT_TRANSACT_QUERY_QUOTA,
		subcommands.NT_TRANSACT_SET_QUOTA,
	}
	for _, function := range unimplemented {
		function := function
		t.Run(fmt.Sprintf("0x%04X", uint16(function)), func(t *testing.T) {
			status := sendNtTransact(t, client, function, nil, nil, nil)
			if status != uint32(nt_status.NT_STATUS_NOT_IMPLEMENTED) {
				t.Fatalf("function 0x%04X answered 0x%08X, want NT_STATUS_NOT_IMPLEMENTED",
					uint16(function), status)
			}
		})
	}
}

// TestNtTransactRejectsOversizeTotals asserts a declared total beyond the limit is
// refused rather than allocated. The NT_TRANSACT counts are 32-bit, so unlike
// TRANSACTION2 there is no natural ceiling and a client could otherwise ask for
// gigabytes.
func TestNtTransactRejectsOversizeTotals(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	_, client := fileServer(t, fs, false)

	request := newRequest(codes.SMB_COM_NT_TRANSACT)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID

	transaction := commands.NewNtTransactRequest()
	transaction.Function = types.USHORT(subcommands.NT_TRANSACT_IOCTL)
	transaction.SetupCount = types.UCHAR(0)
	transaction.Setup = []types.USHORT{}
	// Far beyond the imposed limit.
	transaction.TotalDataCount = types.ULONG(1 << 28)
	request.AddCommand(transaction)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}
	status := binary.LittleEndian.Uint32(raw[5:9])
	if status != uint32(nt_status.NT_STATUS_INVALID_PARAMETER) {
		t.Fatalf("an oversize NT_TRANSACT answered 0x%08X, want NT_STATUS_INVALID_PARAMETER", status)
	}
}

// TestNtCancelIsAcceptedSilently asserts a cancel produces no response at all and
// leaves the connection usable.
//
// The silence is the assertion: a client that received a response to a cancel
// would read it as the answer to whatever it sends next, and every exchange after
// that would be off by one. The following echo is what proves the connection is
// still in step.
func TestNtCancelIsAcceptedSilently(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	_, client := fileServer(t, fs, false)

	request := newRequest(codes.SMB_COM_NT_CANCEL)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID
	request.AddCommand(commands.NewNtCancelRequest())

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the cancel: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	// If the cancel had been answered, this echo would read that answer instead of
	// its own and the payload would not match.
	payload := []byte("still in step")
	echoed, err := client.Echo(payload)
	if err != nil {
		t.Fatalf("the connection is unusable after a cancel: %v", err)
	}
	if !bytes.Equal(echoed, payload) {
		t.Fatalf("the echo returned %q, want %q: the cancel was answered", echoed, payload)
	}
}

// querySecurityDescriptor sends an NT_TRANSACT_QUERY_SECURITY_DESC and returns the
// descriptor and status.
func querySecurityDescriptor(t *testing.T, client *smb1client.Client, fid uint16, information SecurityInformation) ([]byte, uint32) {
	t.Helper()

	parameters := make([]byte, 8)
	binary.LittleEndian.PutUint16(parameters[0:2], fid)
	binary.LittleEndian.PutUint32(parameters[4:8], uint32(information))

	request := newRequest(codes.SMB_COM_NT_TRANSACT)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID

	transaction := commands.NewNtTransactRequest()
	transaction.Function = types.USHORT(subcommands.NT_TRANSACT_QUERY_SECURITY_DESC)
	transaction.SetupCount = types.UCHAR(0)
	transaction.Setup = []types.USHORT{}
	transaction.MaxParameterCount = 1024
	transaction.MaxDataCount = 8192
	transaction.TotalParameterCount = types.ULONG(len(parameters))
	transaction.ParameterCount = types.ULONG(len(parameters))
	transaction.NT_Trans_Parameters = parameters
	request.AddCommand(transaction)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}
	response := message.NewMessage()
	if err := response.Unmarshal(raw); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if response.Header.Status != 0 {
		return nil, response.Header.Status
	}
	ntResponse, ok := response.Command.(*commands.NtTransactResponse)
	if !ok {
		t.Fatalf("response command is %T", response.Command)
	}
	return []byte(ntResponse.Data), 0
}

// sendNtTransact sends an NT_TRANSACT and returns the status it was answered with.
func sendNtTransact(
	t *testing.T,
	client *smb1client.Client,
	function subcommands.NtTransactSubcommand,
	setup []types.USHORT,
	parameters, data []byte,
) uint32 {
	t.Helper()

	request := newRequest(codes.SMB_COM_NT_TRANSACT)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID

	transaction := commands.NewNtTransactRequest()
	transaction.Function = types.USHORT(function)
	transaction.Setup = setup
	transaction.SetupCount = types.UCHAR(len(setup))
	transaction.MaxParameterCount = 1024
	transaction.MaxDataCount = 4096
	transaction.TotalParameterCount = types.ULONG(len(parameters))
	transaction.ParameterCount = types.ULONG(len(parameters))
	transaction.NT_Trans_Parameters = parameters
	transaction.TotalDataCount = types.ULONG(len(data))
	transaction.DataCount = types.ULONG(len(data))
	transaction.NT_Trans_Data = data
	request.AddCommand(transaction)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}
	if len(raw) < 9 {
		t.Fatalf("the response is %d bytes", len(raw))
	}
	return binary.LittleEndian.Uint32(raw[5:9])
}
