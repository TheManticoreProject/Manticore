package catalog

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// mustGUID parses a hyphenated (format D) UUID, panicking on a malformed literal.
// It keeps the seed table below readable and transcription-safe; a bad literal
// fails the package's init/test rather than corrupting a lookup.
func mustGUID(s string) guid.GUID {
	g, err := guid.FromFormatD(s)
	if err != nil {
		panic(fmt.Sprintf("catalog: invalid UUID %q: %v", s, err))
	}
	return *g
}

// v is shorthand for an interface version.
func v(major, minor uint16) Version { return Version{Major: major, Minor: minor} }

// builtin is the curated catalog of well-known DCE/RPC interfaces. UUIDs are
// taken from the MS-* protocol documents and cross-checked against public
// endpoint-mapper references; Executable is the implementing image (an .exe, or
// a .dll loaded into svchost.exe / lsass.exe). Fields that are not
// well-established are left empty rather than guessed.
var builtin = []Interface{
	// --- interfaces with a client implementation in this repo ---
	{
		UUID: mustGUID("12345778-1234-abcd-ef00-0123456789ab"), Version: v(0, 0),
		Name: "lsarpc", Title: "Local Security Authority (domain policy)",
		Description: "LSA policy and translation (also MS-LSAT, MS-DSSP).",
		Executable:  "lsasrv.dll", Protocol: "MS-LSAD", Pipes: []string{`\pipe\lsarpc`},
	},
	{
		UUID: mustGUID("12345778-1234-abcd-ef00-0123456789ac"), Version: v(1, 0),
		Name: "samr", Title: "Security Account Manager",
		Description: "Remote SAM database access (users, groups, domains).",
		Executable:  "samsrv.dll", Protocol: "MS-SAMR", Pipes: []string{`\pipe\samr`},
	},
	{
		UUID: mustGUID("338cd001-2244-31f1-aaaa-900038001003"), Version: v(1, 0),
		Name: "winreg", Title: "Windows Remote Registry",
		Description: "Remote registry query and manipulation.",
		Executable:  "regsvc.dll", Service: "RemoteRegistry", Protocol: "MS-RRP", Pipes: []string{`\pipe\winreg`},
	},
	{
		UUID: mustGUID("367abb81-9844-35f1-ad32-98f038001003"), Version: v(2, 0),
		Name: "svcctl", Title: "Service Control Manager",
		Description: "Remote creation/control of Windows services.",
		Executable:  "services.exe", Protocol: "MS-SCMR", Pipes: []string{`\pipe\svcctl`},
	},
	{
		UUID: mustGUID("3dde7c30-165d-11d1-ab8f-00805f14db40"), Version: v(1, 0),
		Name: "backupkey", Title: "BackupKey Remote Protocol",
		Description: "DPAPI domain backup key wrap/unwrap; abused to steal the domain backup key.",
		Executable:  "lsass.exe", Protocol: "MS-BKRP", Pipes: []string{`\pipe\protected_storage`, `\pipe\ntsvcs`},
	},
	{
		UUID: mustGUID("4b324fc8-1670-01d3-1278-5a47bf6ee188"), Version: v(3, 0),
		Name: "srvsvc", Title: "Server Service",
		Description: "Server service (shares, sessions, files).",
		Executable:  "srvsvc.dll", Service: "LanmanServer", Protocol: "MS-SRVS", Pipes: []string{`\pipe\srvsvc`},
	},
	{
		UUID: mustGUID("c681d488-d850-11d0-8c52-00c04fd90f7e"), Version: v(1, 0),
		Name: "efsrpc", Title: "Encrypting File System Remote",
		Description: "EFSRPC; abused for coercion (PetitPotam).",
		Executable:  "efslsaext.dll", Protocol: "MS-EFSR", Pipes: []string{`\pipe\efsrpc`, `\pipe\lsarpc`},
	},
	{
		UUID: mustGUID("e1af8308-5d1f-11c9-91a4-08002b14a0fa"), Version: v(3, 0),
		Name: "epm", Title: "Endpoint Mapper",
		Description: "Maps interface UUIDs to transport endpoints (port 135).",
		Executable:  "rpcss.dll", Service: "RpcEptMapper", Protocol: "C706", Pipes: []string{`\pipe\epmapper`},
	},
	{
		UUID: mustGUID("e3514235-4b06-11d1-ab04-00c04fc2dcd2"), Version: v(4, 0),
		Name: "drsuapi", Title: "Directory Replication Service",
		Description: "AD replication; abused for DCSync. Bound over ncacn_ip_tcp (no named pipe).",
		Executable:  "ntdsai.dll", Service: "NTDS", Protocol: "MS-DRSR",
	},
	{
		UUID: mustGUID("7c44d7d4-31d5-424c-bd5e-2b3e1f323d22"), Version: v(1, 0),
		Name: "dsaop", Title: "Directory Service Administration",
		Description: "RODC demotion scripting (MS-DRSR). Bound over ncacn_ip_tcp (no named pipe).",
		Executable:  "ntdsai.dll", Service: "NTDS", Protocol: "MS-DRSR",
	},

	// --- other well-known interfaces ---
	{
		UUID: mustGUID("12345678-1234-abcd-ef00-0123456789ab"), Version: v(1, 0),
		Name: "spoolss", Title: "Print System Remote Protocol",
		Description: "Print spooler; abused for coercion (PrinterBug).",
		Executable:  "spoolsv.exe", Service: "Spooler", Protocol: "MS-RPRN", Pipes: []string{`\pipe\spoolss`},
	},
	{
		UUID: mustGUID("76f03f96-cdfd-44fc-a22c-64950a001209"), Version: v(1, 0),
		Name: "IRemoteWinspool", Title: "Print System Asynchronous Remote Protocol",
		Description: "Asynchronous print spooler interface.",
		Executable:  "spoolsv.exe", Service: "Spooler", Protocol: "MS-PAR", Pipes: []string{`\pipe\spoolss`},
	},
	{
		UUID: mustGUID("0b6edbfa-4a24-4fc6-8a23-942b1eca65d1"), Version: v(1, 0),
		Name: "IRPCAsyncNotify", Title: "Print System Asynchronous Notification Protocol",
		Description: "Registers print clients and exchanges asynchronous print notifications.",
		Executable:  "spoolsv.exe", Service: "Spooler", Protocol: "MS-PAN",
	},
	{
		UUID: mustGUID("ae33069b-a2a8-46ee-a235-ddfd339be281"), Version: v(1, 0),
		Name: "IRPCRemoteObject", Title: "Print System Asynchronous Notification Remote Object",
		Description: "Creates and destroys remote objects that refer to printers.",
		Executable:  "spoolsv.exe", Service: "Spooler", Protocol: "MS-PAN",
	},
	{
		UUID: mustGUID("d95afe70-a6d5-4259-822e-2c84da1ddb0d"), Version: v(1, 0),
		Name: "WindowsShutdown", Title: "Remote Shutdown Protocol (WindowsShutdown)",
		Description: "Initiates or aborts a remote shutdown through a dynamic RPC-over-TCP endpoint.",
		Protocol:    "MS-RSP",
	},
	{
		UUID: mustGUID("906b0ce0-c70b-1067-b317-00dd010662da"), Version: v(1, 0),
		Name: "IXnRemote", Title: "MSDTC Connection Manager: OleTx Transports Protocol",
		Description: "Establishes peer-to-peer MSDTC partner communication over dynamic RPC endpoints.",
		Protocol:    "MS-CMPO",
	},
	{
		UUID: mustGUID("12345678-1234-abcd-ef00-01234567cffb"), Version: v(1, 0),
		Name: "netlogon", Title: "Netlogon Remote Protocol",
		Description: "Domain authentication; abused via Zerologon.",
		Executable:  "netlogon.dll", Service: "Netlogon", Protocol: "MS-NRPC", Pipes: []string{`\pipe\netlogon`},
	},
	{
		UUID: mustGUID("6bffd098-a112-3610-9833-46c3f87e345a"), Version: v(1, 0),
		Name: "wkssvc", Title: "Workstation Service",
		Description: "Workstation service (sessions, domain join).",
		Executable:  "wkssvc.dll", Service: "LanmanWorkstation", Protocol: "MS-WKST", Pipes: []string{`\pipe\wkssvc`},
	},
	{
		UUID: mustGUID("1ff70682-0a51-30e8-076d-740be8cee98b"), Version: v(1, 0),
		Name: "atsvc", Title: "Task Scheduler (ATSvc)",
		Description: "Legacy 'at' job scheduling.",
		Executable:  "schedsvc.dll", Service: "Schedule", Protocol: "MS-TSCH", Pipes: []string{`\pipe\atsvc`},
	},
	{
		UUID: mustGUID("86d35949-83c9-4044-b424-db363231fd0c"), Version: v(1, 0),
		Name: "ITaskSchedulerService", Title: "Task Scheduler Service",
		Description: "Modern task scheduler remoting.",
		Executable:  "schedsvc.dll", Service: "Schedule", Protocol: "MS-TSCH",
	},
	{
		UUID: mustGUID("82273fdc-e32a-18c3-3f78-827929dc23ea"), Version: v(0, 0),
		Name: "eventlog", Title: "EventLog Remoting Protocol",
		Description: "Classic event log access.",
		Executable:  "wevtsvc.dll", Service: "EventLog", Protocol: "MS-EVEN", Pipes: []string{`\pipe\eventlog`},
	},
	{
		UUID: mustGUID("f6beaff7-1e19-4fbb-9f8f-b89e2018337c"), Version: v(1, 0),
		Name: "IEventService", Title: "EventLog Remoting Protocol v6",
		Description: "Modern event log access.",
		Executable:  "wevtsvc.dll", Service: "EventLog", Protocol: "MS-EVEN6",
	},
	{
		UUID: mustGUID("afa8bd80-7d8a-11c9-bef4-08002b102989"), Version: v(1, 0),
		Name: "mgmt", Title: "Remote Management",
		Description: "DCE/RPC management interface (inq_if_ids, etc.).",
		Executable:  "rpcss.dll", Service: "RpcSs", Protocol: "C706",
	},
	{
		UUID: mustGUID("99fcfec4-5260-101b-bbcb-00aa0021347a"), Version: v(0, 0),
		Name: "IObjectExporter", Title: "DCOM OXID Resolver",
		Description: "DCOM object exporter / OXID resolution (port 135).",
		Executable:  "rpcss.dll", Service: "RpcSs", Protocol: "MS-DCOM",
	},
	{
		UUID: mustGUID("000001a0-0000-0000-c000-000000000046"), Version: v(0, 0),
		Name: "ISystemActivator", Title: "DCOM Remote Activation",
		Description: "DCOM remote object activation.",
		Executable:  "rpcss.dll", Service: "RpcSs", Protocol: "MS-DCOM",
	},
	{
		UUID: mustGUID("4d9f4ab8-7d1c-11cf-861e-0020af6e7c57"), Version: v(0, 0),
		Name: "IActivation", Title: "DCOM Activation (legacy)",
		Description: "Legacy DCOM activation interface.",
		Executable:  "rpcss.dll", Service: "RpcSs", Protocol: "MS-DCOM",
	},
	{
		UUID: mustGUID("f309ad18-d86a-11d0-a075-00c04fb68820"), Version: v(0, 0),
		Name: "IWbemLevel1Login", Title: "Windows Management Instrumentation",
		Description: "WMI remoting login interface (over DCOM).",
		Service:     "Winmgmt", Protocol: "MS-WMI",
	},
	{
		UUID: mustGUID("50abc2a4-574d-40b3-9d66-ee4fd5fba076"), Version: v(5, 0),
		Name: "dnsserver", Title: "DNS Server Management",
		Description: "Remote DNS server administration.",
		Executable:  "dns.exe", Service: "DNS", Protocol: "MS-DNSP",
	},
	{
		UUID: mustGUID("a8e0653c-2744-4389-a61d-7373df8b2292"), Version: v(1, 0),
		Name: "FssagentRpc", Title: "File Server Remote VSS",
		Description: "File Server VSS agent; abused for coercion.",
		Service:     "FileServerVssAgent", Protocol: "MS-FSRVP", Pipes: []string{`\pipe\FssagentRpc`},
	},
	{
		UUID: mustGUID("3919286a-b10c-11d0-9ba8-00c04fd92ef5"), Version: v(0, 0),
		Name: "dssetup", Title: "Directory Services Setup",
		Description: "Domain/role information (DsRoleGetPrimaryDomainInformation).",
		Executable:  "lsasrv.dll", Protocol: "MS-DSSP", Pipes: []string{`\pipe\lsarpc`},
	},
	{
		UUID: mustGUID("df1941c5-fe89-4e79-bf10-463657acf44d"), Version: v(1, 0),
		Name: "efsrpc", Title: "Encrypting File System Remote (lsarpc binding)",
		Description: "Alternate EFSRPC interface UUID over \\pipe\\lsarpc.",
		Executable:  "efslsaext.dll", Protocol: "MS-EFSR", Pipes: []string{`\pipe\lsarpc`},
	},
}
