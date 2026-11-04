package ms_tsch

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/86d35949-83c9-4044-b424-db363231fd0c/1.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// Task creation disposition flags for SchRpcRegisterTask ([MS-TSCH] 2.3.9).
const (
	TaskValidateOnly   uint32 = 0x01
	TaskCreate         uint32 = 0x02
	TaskUpdate         uint32 = 0x04
	TaskCreateOrUpdate uint32 = TaskCreate | TaskUpdate
)

// Logon types for SchRpcRegisterTask ([MS-TSCH] 2.3.9 TASK_LOGON_TYPE). TaskLogonNone
// leaves the logon method to the task's XML principal.
const TaskLogonNone uint32 = 0

// taskXMLTemplate is a minimal, hidden, run-on-demand task that executes a single command
// as LocalSystem (S-1-5-18) at the highest run level. It carries no trigger: the task is
// fired explicitly by SchRpcRun, the way atexec runs a command. The placeholders are, in
// order: registration date, author, command, arguments.
const taskXMLTemplate = `<?xml version="1.0" encoding="UTF-16"?>
<Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <RegistrationInfo>
    <Date>%s</Date>
    <Author>%s</Author>
  </RegistrationInfo>
  <Principals>
    <Principal id="Author">
      <UserId>S-1-5-18</UserId>
      <RunLevel>HighestAvailable</RunLevel>
    </Principal>
  </Principals>
  <Settings>
    <MultipleInstancesPolicy>IgnoreNew</MultipleInstancesPolicy>
    <DisallowStartIfOnBatteries>false</DisallowStartIfOnBatteries>
    <StopIfGoingOnBatteries>false</StopIfGoingOnBatteries>
    <AllowHardTerminate>true</AllowHardTerminate>
    <RunOnlyIfNetworkAvailable>false</RunOnlyIfNetworkAvailable>
    <IdleSettings>
      <StopOnIdleEnd>true</StopOnIdleEnd>
      <RestartOnIdle>false</RestartOnIdle>
    </IdleSettings>
    <AllowStartOnDemand>true</AllowStartOnDemand>
    <Enabled>true</Enabled>
    <Hidden>true</Hidden>
    <RunOnlyIfIdle>false</RunOnlyIfIdle>
    <WakeToRun>false</WakeToRun>
    <ExecutionTimeLimit>PT0S</ExecutionTimeLimit>
    <Priority>7</Priority>
  </Settings>
  <Actions Context="Author">
    <Exec>
      <Command>%s</Command>
      <Arguments>%s</Arguments>
    </Exec>
  </Actions>
</Task>`

// BuildTaskXML returns a registration XML for a hidden, run-on-demand LocalSystem task
// that executes command with args. command and args are XML-escaped.
func BuildTaskXML(command, args string) string {
	return fmt.Sprintf(taskXMLTemplate,
		time.Now().UTC().Format("2006-01-02T15:04:05"),
		xmlEscape("Manticore"),
		xmlEscape(command),
		xmlEscape(args),
	)
}

// xmlEscape escapes s for use as XML element text.
func xmlEscape(s string) string {
	var b bytes.Buffer
	if err := xml.EscapeText(&b, []byte(s)); err != nil {
		return s
	}
	return b.String()
}

// wstrPtr returns a *ndr.WSTR for s, or nil for the empty string so an optional [unique]
// pointer parameter is sent as a null pointer.
func wstrPtr(s string) *ndr.WSTR {
	if s == "" {
		return nil
	}
	w := ndr.WSTR(s)
	return &w
}

// RegisterTask registers (creates) a scheduled task at the given path (for example
// "\MyTask") from a task-definition XML. It returns the task path the server recorded.
func (c *Client) RegisterTask(path, taskXML string) (string, error) {
	if err := c.ensure(); err != nil {
		return "", err
	}
	actual, _, err := functions.SchRpcRegisterTask(
		c.rpc,
		wstrPtr(path),
		ndr.WSTR(taskXML),
		ndr.DWORD(TaskCreate),
		nil, // sddl
		ndr.DWORD(TaskLogonNone),
		0,   // cCreds
		nil, // pCreds
	)
	if err != nil {
		return "", fmt.Errorf("ms_tsch: RegisterTask %q: %w", path, err)
	}
	if actual != nil {
		return string(*actual), nil
	}
	return path, nil
}

// Run starts the task at path immediately and returns the GUID of the running instance.
func (c *Client) Run(path string) (guid.GUID, error) {
	if err := c.ensure(); err != nil {
		return guid.GUID{}, err
	}
	g, err := functions.SchRpcRun(c.rpc, ndr.WSTR(path), 0, nil, 0, 0, nil)
	if err != nil {
		return guid.GUID{}, fmt.Errorf("ms_tsch: Run %q: %w", path, err)
	}
	return g, nil
}

// Delete removes the task (or empty folder) at path.
func (c *Client) Delete(path string) error {
	if err := c.ensure(); err != nil {
		return err
	}
	if err := functions.SchRpcDelete(c.rpc, ndr.WSTR(path), 0); err != nil {
		return fmt.Errorf("ms_tsch: Delete %q: %w", path, err)
	}
	return nil
}

// RegisterRunDelete is the scheduled-task command-execution primitive: it registers a
// hidden LocalSystem task named taskName that runs command with args, runs it, then
// deletes it. It is how atexec and an NTLM relay run a command over SMB.
//
// The task runs asynchronously: Run returns once the instance has been launched, before
// command has necessarily finished, so capturing output is the caller's job (for example
// a command that redirects to a share). Deletion is attempted even if the run reports an
// error, so a transient run failure does not leave the task behind; a run error is still
// surfaced, and a cleanup failure is reported only once the command has already run.
func (c *Client) RegisterRunDelete(taskName, command, args string) error {
	if err := c.ensure(); err != nil {
		return err
	}
	path := `\` + taskName

	if _, err := c.RegisterTask(path, BuildTaskXML(command, args)); err != nil {
		return err
	}

	_, runErr := c.Run(path)
	delErr := c.Delete(path)

	if runErr != nil {
		return runErr
	}
	if delErr != nil {
		return fmt.Errorf("ms_tsch: RegisterRunDelete: the task ran but %q was not removed: %w", path, delErr)
	}
	return nil
}
