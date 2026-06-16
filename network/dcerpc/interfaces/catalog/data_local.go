package catalog

// local holds well-known local/host RPC interfaces — the per-service NCALRPC
// and named-pipe interfaces a Windows host registers with the endpoint mapper
// (NGC/Windows Hello, the firewall, the biometric service, …). Most are not
// documented MS-* protocols, so the title is the friendly name the service
// itself reports to the endpoint mapper; the implementing image and hosting
// service are cross-checked against public Windows service references and are
// left empty where not well-established.
var local = []Interface{
	// Application Information / UAC elevation (appinfo.dll, Appinfo service).
	{UUID: mustGUID("201ef99a-7fa0-444c-9399-19ba84f12a1a"), Version: v(1, 0), Name: "AppInfo", Title: "Application Information (UAC)", Description: "Application Information service (UAC elevation).", Executable: "appinfo.dll", Service: "Appinfo"},
	{UUID: mustGUID("58e604e8-9adb-4d2e-a464-3b0683fb1480"), Version: v(1, 0), Name: "AppInfo", Title: "Application Information (UAC)", Description: "Application Information service (UAC elevation).", Executable: "appinfo.dll", Service: "Appinfo"},
	{UUID: mustGUID("5f54ce7d-5b79-4175-8584-cb65313a0e98"), Version: v(1, 0), Name: "AppInfo", Title: "Application Information (UAC)", Description: "Application Information service (UAC elevation).", Executable: "appinfo.dll", Service: "Appinfo"},
	{UUID: mustGUID("fb9a3757-cff0-4db0-b9fc-bd6c131612fd"), Version: v(1, 0), Name: "AppInfo", Title: "Application Information (UAC)", Description: "Application Information service (UAC elevation).", Executable: "appinfo.dll", Service: "Appinfo"},
	{UUID: mustGUID("fd7a0523-dc70-43dd-9b2e-9c5ed48225b1"), Version: v(1, 0), Name: "AppInfo", Title: "Application Information (UAC)", Description: "Application Information service (UAC elevation).", Executable: "appinfo.dll", Service: "Appinfo"},

	// Windows Defender Firewall: Base Filtering Engine + the firewall APIs.
	{UUID: mustGUID("dd490425-5325-4565-b774-7e27d6c09c24"), Version: v(1, 0), Name: "BFE", Title: "Base Firewall Engine API", Description: "Base Filtering Engine (Windows Filtering Platform).", Executable: "bfe.dll", Service: "BFE"},
	{UUID: mustGUID("2fb92682-6599-42dc-ae13-bd2ca89bd11c"), Version: v(1, 0), Name: "FwApis", Title: "Windows Firewall APIs", Description: "Windows Defender Firewall management.", Executable: "mpssvc.dll", Service: "MpsSvc"},
	{UUID: mustGUID("7f9d11bf-7fb9-436b-a812-b2d50c5d4c03"), Version: v(1, 0), Name: "FwApis", Title: "Windows Firewall APIs", Description: "Windows Defender Firewall management.", Executable: "mpssvc.dll", Service: "MpsSvc"},
	{UUID: mustGUID("f47433c3-3e9d-4157-aad4-83aa1f5c2d4c"), Version: v(1, 0), Name: "FwApis", Title: "Windows Firewall APIs", Description: "Windows Defender Firewall management.", Executable: "mpssvc.dll", Service: "MpsSvc"},

	// DHCP client.
	{UUID: mustGUID("3c4728c5-f0ab-448b-bda1-6ce01eb0a6d5"), Version: v(1, 0), Name: "DHCPClient", Title: "DHCP Client LRPC Endpoint", Description: "DHCP client service.", Executable: "dhcpcsvc.dll", Service: "Dhcp"},
	{UUID: mustGUID("3c4728c5-f0ab-448b-bda1-6ce01eb0a6d6"), Version: v(1, 0), Name: "DHCPv6Client", Title: "DHCPv6 Client LRPC Endpoint", Description: "DHCPv6 client service.", Executable: "dhcpcsvc6.dll", Service: "Dhcp"},

	// EFS and DFS-R.
	{UUID: mustGUID("04eeb297-cbf4-466b-8a2a-bfd6a2f10bba"), Version: v(1, 0), Name: "EfsK", Title: "EFSK RPC Interface", Description: "Encrypting File System key service interface.", Executable: "efssvc.dll"},
	{UUID: mustGUID("897e2e5f-93f3-4376-9c9c-fd2277495c27"), Version: v(1, 0), Name: "Frs2", Title: "DFS Replication (FRS2)", Description: "Distributed File System Replication.", Executable: "dfsrs.exe", Service: "DFSR", Protocol: "MS-FRS2"},

	// Group Policy client.
	{UUID: mustGUID("2eb08e3e-639f-4fba-97b1-14f878961076"), Version: v(1, 0), Name: "GroupPolicy", Title: "Group Policy RPC Interface", Description: "Group Policy client service.", Executable: "gpsvc.dll", Service: "gpsvc"},

	// CNG Key Isolation.
	{UUID: mustGUID("b25a52bf-e5dd-4f4a-aea6-8ca7272a0e86"), Version: v(2, 0), Name: "KeyIso", Title: "CNG Key Isolation", Description: "Isolates private keys in the LSA process.", Executable: "keyiso.dll", Service: "KeyIso"},

	// Microsoft Passport / Windows Hello (NGC), all hosted by ngcsvc.dll.
	{UUID: mustGUID("0e3ae095-8a23-48f4-9782-03c1594a890e"), Version: v(1, 0), Name: "NgcKsp", Title: "NGC Service KSP RPC Interface", Description: "Windows Hello (NGC) key storage provider.", Executable: "ngcsvc.dll", Service: "NgcSvc"},
	{UUID: mustGUID("9cbc9d3a-7586-4814-8d70-18737dcbe523"), Version: v(1, 0), Name: "NgcLocalAccountVault", Title: "NGC Service LocalAccount Vault Interface", Description: "Windows Hello (NGC) local account vault.", Executable: "ngcsvc.dll", Service: "NgcSvc"},
	{UUID: mustGUID("c225e799-29de-42af-bc05-1e2127cc056e"), Version: v(1, 0), Name: "NgcManagement", Title: "NGC Service Management RPC Interface", Description: "Windows Hello (NGC) management.", Executable: "ngcsvc.dll", Service: "NgcSvc"},
	{UUID: mustGUID("2c2fd034-05cb-4f44-a75d-2a5cf81499c1"), Version: v(1, 0), Name: "NgcRemotePairing", Title: "NGC Service Remote Pairing RPC Interface", Description: "Windows Hello (NGC) remote pairing.", Executable: "ngcsvc.dll", Service: "NgcSvc"},
	{UUID: mustGUID("2b70bed6-1757-4d22-9f39-448589fbebf5"), Version: v(1, 0), Name: "NgcTicket", Title: "NGC Service Ticket RPC Interface", Description: "Windows Hello (NGC) ticket service.", Executable: "ngcsvc.dll", Service: "NgcSvc"},
	{UUID: mustGUID("51a227ae-825b-41f2-b4a9-1ac9557a1018"), Version: v(1, 0), Name: "NgcPopKey", Title: "Ngc Pop Key Service", Description: "Windows Hello (NGC) proof-of-possession key.", Executable: "ngcsvc.dll", Service: "NgcSvc"},
	{UUID: mustGUID("8fb74744-b2ff-4c00-be0d-9ef9a191fe1b"), Version: v(1, 0), Name: "NgcPopKey", Title: "Ngc Pop Key Service", Description: "Windows Hello (NGC) proof-of-possession key.", Executable: "ngcsvc.dll", Service: "NgcSvc"},

	// NGC handler interfaces (LRPC), ngcsvc.dll.
	{UUID: mustGUID("e6f89680-fc98-11e3-80d4-10604b681cfa"), Version: v(1, 0), Name: "INgcGidsHandler", Title: "INgcGidsHandler", Description: "Windows Hello (NGC) GIDS smart-card handler.", Executable: "ngcsvc.dll", Service: "NgcSvc"},
	{UUID: mustGUID("2d24ff0b-1bab-404c-a0fd-42c85577bf68"), Version: v(1, 0), Name: "INgcHandler", Title: "INgcHandler", Description: "Windows Hello (NGC) handler.", Executable: "ngcsvc.dll", Service: "NgcSvc"},
	{UUID: mustGUID("30034843-029d-46ec-8fff-5d12987f85c4"), Version: v(1, 0), Name: "INgcProvisioningHandler", Title: "INgcProvisioningHandler", Description: "Windows Hello (NGC) provisioning handler.", Executable: "ngcsvc.dll", Service: "NgcSvc"},
	{UUID: mustGUID("8bef2320-f308-4720-b913-0129cecfa6b9"), Version: v(1, 0), Name: "IVscProvisioningHandler", Title: "IVscProvisioningHandler", Description: "Virtual smart card provisioning handler."},

	// Network services.
	{UUID: mustGUID("7ea70bcf-48af-4f6a-8968-6a440754d5fa"), Version: v(1, 0), Name: "NSI", Title: "Network Store Interface", Description: "Network Store Interface service.", Executable: "nsisvc.dll", Service: "nsi"},
	{UUID: mustGUID("552d076a-cb29-4e44-8b6a-d15e59e2c0af"), Version: v(1, 0), Name: "IPTransition", Title: "IP Transition Configuration endpoint", Description: "IP Helper IP transition configuration.", Executable: "iphlpsvc.dll", Service: "iphlpsvc"},
	{UUID: mustGUID("e40f7b57-7a25-4cd3-a135-7f7d3df9d16b"), Version: v(1, 0), Name: "NetworkConnectionBroker", Title: "Network Connection Broker server endpoint", Description: "Network Connection Broker.", Executable: "ncbservice.dll", Service: "NcbService"},
	{UUID: mustGUID("5222821f-d5e2-4885-84f1-5f6185a0ec41"), Version: v(1, 0), Name: "NetworkConnectionBroker", Title: "Network Connection Broker server endpoint (NCB Reset)", Description: "Network Connection Broker (NCB reset module).", Executable: "ncbservice.dll", Service: "NcbService"},
	{UUID: mustGUID("3473dd4d-2e88-4006-9cba-22570909dd10"), Version: v(5, 1), Name: "WinHttpAutoProxy", Title: "WinHTTP Web Proxy Auto-Discovery Service", Description: "WinHTTP WPAD proxy auto-discovery.", Executable: "winhttp.dll", Service: "WinHttpAutoProxySvc"},

	// Program Compatibility Assistant, biometrics, user manager, software protection.
	{UUID: mustGUID("0767a036-0d22-48aa-ba69-b619480f38cb"), Version: v(1, 0), Name: "PcaSvc", Title: "Program Compatibility Assistant", Description: "Program Compatibility Assistant service.", Executable: "pcasvc.dll", Service: "PcaSvc"},
	{UUID: mustGUID("4be96a0f-9f52-4729-a51d-c70610f118b0"), Version: v(1, 0), Name: "wbiosrvc", Title: "Windows Biometric Service", Description: "Windows Biometric Service.", Executable: "wbiosrvc.dll", Service: "WbioSrvc"},
	{UUID: mustGUID("c0e9671e-33c6-4438-9464-56b2e1b1c7b4"), Version: v(1, 0), Name: "wbiosrvc", Title: "Windows Biometric Service", Description: "Windows Biometric Service.", Executable: "wbiosrvc.dll", Service: "WbioSrvc"},
	{UUID: mustGUID("0d3c7f20-1c8d-4654-a1b3-51563b298bda"), Version: v(1, 0), Name: "UserMgrCli", Title: "User Manager", Description: "User Manager service client interface.", Executable: "usermgr.dll", Service: "UserManager"},
	{UUID: mustGUID("b18fbab6-56f8-4702-84e0-41053293a869"), Version: v(1, 0), Name: "UserMgrCli", Title: "User Manager", Description: "User Manager service client interface.", Executable: "usermgr.dll", Service: "UserManager"},
	{UUID: mustGUID("9435cc56-1d9c-4924-ac7d-b60a2c3520e1"), Version: v(1, 0), Name: "SPPSVC", Title: "Software Protection Platform", Description: "Software licensing / activation service.", Executable: "sppsvc.exe", Service: "sppsvc"},

	// Remote registry performance-counter interface and the legacy LanMan API.
	{UUID: mustGUID("da5a86c5-12c2-4943-ab30-7f74a813d853"), Version: v(1, 0), Name: "RemoteRegistryPerflib", Title: "RemoteRegistry Perflib Interface", Description: "Remote performance-counter (perflib) interface.", Executable: "regsvc.dll", Service: "RemoteRegistry"},
	{UUID: mustGUID("98716d03-89ac-44c7-bb8c-285824e51c4a"), Version: v(1, 0), Name: "XactSrv", Title: "XactSrv (LanMan API)", Description: "Legacy LanMan transaction (XactSrv) interface.", Executable: "srvsvc.dll", Service: "LanmanServer"},

	// Task Scheduler SASec (legacy 'at', documented as MS-TSCH alongside ATSvc).
	{UUID: mustGUID("378e52b0-c0a9-11cf-822d-00aa0051e40f"), Version: v(1, 0), Name: "SASec", Title: "Task Scheduler (SASec)", Description: "Task Scheduler SASec interface.", Executable: "schedsvc.dll", Service: "Schedule", Protocol: "MS-TSCH", Pipes: []string{`\pipe\atsvc`}},

	// Identified by name only; implementing image not well-established.
	{UUID: mustGUID("c49a5a70-8a7f-4e70-ba16-1e8f1f193ef1"), Version: v(1, 0), Name: "AdhApis", Title: "Adh APIs", Description: "ADH (Application Deployment Helper) APIs."},
	{UUID: mustGUID("8337aebc-5564-46fd-bc41-7798f18d2e4b"), Version: v(1, 0), Name: "DeviceCredentialManager", Title: "Device Credential Manager RPC Interface", Description: "Device credential manager."},
	{UUID: mustGUID("4e25f4a2-21e8-40ce-b401-32050413143a"), Version: v(1, 0), Name: "DeviceCredential", Title: "Device Credential RPC Interface", Description: "Device credential interface."},
	{UUID: mustGUID("1a0d010f-1c33-432c-b0f5-8cf4e8053099"), Version: v(1, 0), Name: "IdSegSrv", Title: "IdSegSrv service", Description: "Identity segmentation service."},
	{UUID: mustGUID("880fd55e-43b9-11e0-b1a8-cf4edfd72085"), Version: v(1, 0), Name: "KAPI", Title: "KAPI Service endpoint", Description: "KAPI service endpoint."},
	{UUID: mustGUID("a111f1c5-5923-47c0-9a68-d0bafb577901"), Version: v(1, 0), Name: "NetSetup", Title: "NetSetup API", Description: "Network setup / domain-join API."},
	{UUID: mustGUID("30adc50c-5cbc-46ce-9a0e-91914789e23c"), Version: v(1, 0), Name: "NRP", Title: "NRP server endpoint", Description: "NRP server endpoint."},
	{UUID: mustGUID("c36be077-e14b-4fe9-8abc-e856ef4f048b"), Version: v(1, 0), Name: "ProxyManagerClient", Title: "Proxy Manager client server endpoint", Description: "Proxy manager (client)."},
	{UUID: mustGUID("2e6035b2-e8f1-41a7-a044-656b439c4c34"), Version: v(1, 0), Name: "ProxyManagerProvider", Title: "Proxy Manager provider server endpoint", Description: "Proxy manager (provider)."},
	{UUID: mustGUID("0b1c2170-5732-4e0e-8cd3-d9b16f3b84d7"), Version: v(0, 0), Name: "RemoteAccessCheck", Title: "RemoteAccessCheck", Description: "Remote access check interface (LSA-hosted)."},
	{UUID: mustGUID("12e65dd8-887f-41ef-91bf-8d816c42c2e7"), Version: v(1, 0), Name: "SecureDesktop", Title: "Secure Desktop LRPC interface", Description: "Secure desktop (UAC) LRPC interface."},
	{UUID: mustGUID("df4df73a-c52d-4e3a-8003-8437fdf8302a"), Version: v(0, 0), Name: "WindowManagerRPC", Title: "Window Manager RPC server", Description: "Window manager RPC interface."},
}
