package ms_scmr

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msscmr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-scmr"
)

// Service Control Manager access rights ([MS-SCMR] 2.2.15) requested when opening the SCM
// database.
const (
	SCManagerConnect       uint32 = 0x0001
	SCManagerCreateService uint32 = 0x0002
	SCManagerEnumerate     uint32 = 0x0004
	SCManagerAllAccess     uint32 = 0xF003F
)

// Service object access rights ([MS-SCMR] 2.2.15) requested when opening or creating a
// service.
const (
	ServiceQueryStatus uint32 = 0x0004
	ServiceStart       uint32 = 0x0010
	ServiceStop        uint32 = 0x0020
	ServiceDelete      uint32 = 0x00010000
	ServiceAllAccess   uint32 = 0xF01FF
)

// Service type, start type and error control for CreateService ([MS-SCMR] 2.2.15), and
// the control codes for ControlService ([MS-SCMR] 3.1.4.2).
const (
	ServiceWin32OwnProcess uint32 = 0x00000010
	ServiceDemandStart     uint32 = 0x00000003
	ServiceErrorIgnore     uint32 = 0x00000000
	ServiceControlStop     uint32 = 0x00000001
)

// wstrPtr returns a *ndr.WSTR for s, or nil for the empty string so an optional [unique]
// pointer parameter is sent as a null pointer.
func wstrPtr(s string) *ndr.WSTR {
	if s == "" {
		return nil
	}
	w := ndr.WSTR(s)
	return &w
}

// OpenSCManager opens the local SCM's active-services database with the given access
// rights (one of the SCManager* constants). The returned handle is closed with
// CloseHandle.
func (c *Client) OpenSCManager(desiredAccess uint32) (Handle, error) {
	if err := c.ensure(); err != nil {
		return Handle{}, err
	}
	h, err := functions.ROpenSCManagerW(c.rpc, nil, nil, ndr.DWORD(desiredAccess))
	if err != nil {
		return Handle{}, fmt.Errorf("ms_scmr: OpenSCManager: %w", err)
	}
	return h, nil
}

// OpenService opens an existing service by name under an SCM handle with the given access
// rights. The returned handle is closed with CloseHandle.
func (c *Client) OpenService(scm Handle, name string, desiredAccess uint32) (Handle, error) {
	if err := c.ensure(); err != nil {
		return Handle{}, err
	}
	h, err := functions.ROpenServiceW(c.rpc, scm, ndr.WSTR(name), ndr.DWORD(desiredAccess))
	if err != nil {
		return Handle{}, fmt.Errorf("ms_scmr: OpenService %q: %w", name, err)
	}
	return h, nil
}

// CreateService creates a Win32 own-process, demand-start service named name under the
// SCM handle, whose image path is binPath (which may be an arbitrary command line). It
// runs as LocalSystem. The returned handle is closed with CloseHandle.
func (c *Client) CreateService(scm Handle, name, displayName, binPath string, desiredAccess uint32) (Handle, error) {
	if err := c.ensure(); err != nil {
		return Handle{}, err
	}
	if displayName == "" {
		displayName = name
	}
	_, h, err := functions.RCreateServiceW(
		c.rpc,
		scm,
		ndr.WSTR(name),
		wstrPtr(displayName),
		ndr.DWORD(desiredAccess),
		ndr.DWORD(ServiceWin32OwnProcess),
		ndr.DWORD(ServiceDemandStart),
		ndr.DWORD(ServiceErrorIgnore),
		ndr.WSTR(binPath),
		nil,       // lpLoadOrderGroup
		nil,       // lpdwTagId
		[]uint8{}, // lpDependencies ([ref], empty = none)
		0,         // dwDependSize
		nil,       // lpServiceStartName (nil = LocalSystem)
		[]uint8{}, // lpPassword ([ref], empty = none)
		0,         // dwPwSize
	)
	if err != nil {
		return Handle{}, fmt.Errorf("ms_scmr: CreateService %q: %w", name, err)
	}
	return h, nil
}

// StartService starts a service, passing args as the service's command-line arguments.
func (c *Client) StartService(svc Handle, args ...string) error {
	if err := c.ensure(); err != nil {
		return err
	}
	argv := make([]msscmr.STRING_PTRSW, len(args))
	for i, a := range args {
		w := ndr.WSTR(a)
		argv[i] = msscmr.STRING_PTRSW{StringPtr: &w}
	}
	if err := functions.RStartServiceW(c.rpc, svc, ndr.DWORD(len(argv)), argv); err != nil {
		return fmt.Errorf("ms_scmr: StartService: %w", err)
	}
	return nil
}

// ControlService sends a control code (a ServiceControl* constant) to a service and
// returns its reported status.
func (c *Client) ControlService(svc Handle, control uint32) (msscmr.SERVICE_STATUS, error) {
	if err := c.ensure(); err != nil {
		return msscmr.SERVICE_STATUS{}, err
	}
	status, err := functions.RControlService(c.rpc, svc, ndr.DWORD(control))
	if err != nil {
		return status, fmt.Errorf("ms_scmr: ControlService: %w", err)
	}
	return status, nil
}

// DeleteService marks a service for deletion. The service record is removed once its last
// handle is closed.
func (c *Client) DeleteService(svc Handle) error {
	if err := c.ensure(); err != nil {
		return err
	}
	if err := functions.RDeleteService(c.rpc, svc); err != nil {
		return fmt.Errorf("ms_scmr: DeleteService: %w", err)
	}
	return nil
}

// CloseHandle closes an SCM or service handle.
func (c *Client) CloseHandle(h Handle) error {
	if err := c.ensure(); err != nil {
		return err
	}
	if _, err := functions.RCloseServiceHandle(c.rpc, h); err != nil {
		return fmt.Errorf("ms_scmr: CloseHandle: %w", err)
	}
	return nil
}

// CreateAndRun is the service-based command-execution primitive: it opens the SCM, creates
// a demand-start service whose image path is command, starts it, then deletes the service
// and closes every handle. It is how the psexec/smbexec family and an NTLM relay run a
// command over SMB.
//
// Starting such a service makes the SCM launch command and wait for it to check in as a
// service, which a plain command never does: the SCM returns an error
// (ERROR_SERVICE_REQUEST_TIMEOUT, or that the process is not a valid service) AFTER the
// command has already run. That start error is therefore expected and not returned;
// CreateAndRun reports only failures to create the service or to clean it up. The caller
// drives output capture (for example a binPath that redirects to a share) itself; use the
// lower-level methods for finer control.
func (c *Client) CreateAndRun(name, command string) error {
	scm, err := c.OpenSCManager(SCManagerCreateService)
	if err != nil {
		return err
	}
	defer c.CloseHandle(scm)

	svc, err := c.CreateService(scm, name, name, command, ServiceAllAccess)
	if err != nil {
		return err
	}
	// The start triggers execution; its error is expected (the command is not a service)
	// and ignored — the command has already run by the time the SCM reports it.
	_ = c.StartService(svc)

	delErr := c.DeleteService(svc)
	closeErr := c.CloseHandle(svc)
	if delErr != nil {
		return fmt.Errorf("ms_scmr: CreateAndRun: the command ran but the service %q was not removed: %w", name, delErr)
	}
	if closeErr != nil {
		return fmt.Errorf("ms_scmr: CreateAndRun: the command ran but the service handle did not close: %w", closeErr)
	}
	return nil
}
