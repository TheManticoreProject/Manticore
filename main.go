package main

import (
	"fmt"
	"net"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/goopts/parser"
)

var (
	// Configuration
	debug bool

	// Network settings
	target string

	// Share / file under test
	shareName    string
	filePath     string
	writeContent string

	// Authentication details
	authDomain   string
	authUsername string
	authPassword string
	authHashes   string
	authNoPass   bool
)

func parseArgs() {
	ap := parser.ArgumentsParser{
		Banner: "poc - by Remi GASCOU (Podalirius) @ TheManticoreProject - v1.0.0",
	}

	ap.NewBoolArgument(&debug, "", "--debug", false, "Enable debug mode.")

	// Network settings
	subparser_list_group_network, err := ap.NewArgumentGroup("Network")
	if err != nil {
		fmt.Printf("[error] Error creating ArgumentGroup: %s\n", err)
	} else {
		subparser_list_group_network.NewStringArgument(&target, "", "--target", "", false, "IP Address of the target host.")
		subparser_list_group_network.NewStringArgument(&shareName, "", "--share", "share", false, "Name of the share to connect to.")
		subparser_list_group_network.NewStringArgument(&filePath, "", "--file", "file.txt", false, "Path of the file to read at the share root.")
		subparser_list_group_network.NewStringArgument(&writeContent, "", "--write", "", false, "If set, write this content to --file (create/overwrite) before reading it back.")
	}
	// Authentication
	subparser_list_group_auth, err := ap.NewArgumentGroup("Authentication")
	if err != nil {
		fmt.Printf("[error] Error creating ArgumentGroup: %s\n", err)
	} else {
		subparser_list_group_auth.NewStringArgument(&authDomain, "-d", "--domain", "", false, "(FQDN) domain to authenticate to.")
		subparser_list_group_auth.NewStringArgument(&authUsername, "-u", "--user", "", false, "User to authenticate with.")
	}
	// Secret
	subparser_list_group_secret, err := ap.NewRequiredMutuallyExclusiveArgumentGroup("Secret")
	if err != nil {
		fmt.Printf("[error] Error creating ArgumentGroup: %s\n", err)
	} else {
		subparser_list_group_secret.NewBoolArgument(&authNoPass, "", "--no-pass", false, "Don't ask for password (useful for -k).")
		subparser_list_group_secret.NewStringArgument(&authPassword, "-p", "--password", "", false, "Password to authenticate with.")
		subparser_list_group_secret.NewStringArgument(&authHashes, "-H", "--hashes", "", false, "NT/LM hashes, format is LMhash:NThash.")
		subparser_list_group_secret.NewStringArgument(&authHashes, "", "--aes-key", "", false, "AES key to use for Kerberos Authentication (128 or 256 bits).")
	}

	// Parse arguments
	ap.Parse()
}

func main() {
	parseArgs()

	client.FileIODebug = debug

	creds, err := credentials.NewCredentials(authDomain, authUsername, authPassword, authHashes)
	if err != nil {
		logger.Warn(fmt.Sprintf("Error creating credentials: %s", err))
		return
	}

	// Create a new SMB client
	c := client.NewClientUsingTCPTransport(net.ParseIP(target), 445)

	// Connect to the SMB server
	err = c.Connect(net.ParseIP(target), 445)
	if err != nil {
		fmt.Printf("[error] Error connecting to SMB server: %s\n", err)
		return
	} else {
		fmt.Printf("[info] Connected to SMB server\n")
	}

	// Set the native OS and LAN Manager
	c.NativeOS = "Unix"
	c.NativeLanMan = "Samba"

	fmt.Printf("[info] Dialect         : [%s]\n", c.Connection.SelectedDialect)
	fmt.Printf("[info] Capabilities    : [0x%08x] (%s)\n", uint32(c.Connection.Server.Capabilities), c.Connection.Server.Capabilities.String())
	fmt.Printf("[info] Session key     : [0x%08x]\n", c.Connection.Server.SessionKey)
	fmt.Printf("[info] Max buffer size : [0x%08x] (%d)\n", c.Connection.Server.MaxBufferSize, c.Connection.Server.MaxBufferSize)
	fmt.Printf("[info] Max mpx count   : [0x%04x] (%d)\n", c.Connection.MaxMpxCount, c.Connection.MaxMpxCount)
	fmt.Printf("[info] System time     : [0x%08x] (%s)\n", c.Connection.Server.SystemTime.ToInt64(), c.Connection.Server.SystemTime.GetTime())
	timezoneShift := (-1 * c.Connection.Server.TimeZone) / 60.0
	if timezoneShift > 0 {
		fmt.Printf("[info] Time zone       : [%d] (UTC+%d)\n", c.Connection.Server.TimeZone, timezoneShift)
	} else {
		fmt.Printf("[info] Time zone       : [%d] (UTC-%d)\n", c.Connection.Server.TimeZone, timezoneShift)
	}
	fmt.Printf("[info] SecurityMode    : [0x%02x] (%s)\n", uint8(c.Connection.Server.SecurityMode), c.Connection.Server.SecurityMode.String())
	if c.Connection.Server.SecurityMode.SupportsChallengeResponseAuth() {
		fmt.Printf("[info] Server GUID     : [%s]\n", c.Connection.Server.ServerGUID.ToFormatD())
	} else {
		fmt.Printf("[info] Domain name     : [%s]\n", c.Connection.Server.DomainName)
		fmt.Printf("[info] Server name     : [%s]\n", c.Connection.Server.Name)
	}

	err = c.SessionSetup(creds)
	if err != nil {
		fmt.Printf("[error] Error setting up session: %s\n", err)
		return
	}

	err = c.TreeConnect(shareName)
	if err != nil {
		fmt.Printf("[error] Error connecting to share %q: %s\n", shareName, err)
		return
	}
	fmt.Printf("[info] Connected to share [%s] (TID 0x%04x)\n", shareName, c.Session.TreeID)

	// Optionally write content to the file first (creating or overwriting it).
	if writeContent != "" {
		wfid, err := c.OpenFile(
			filePath,
			client.GENERIC_READ|client.GENERIC_WRITE,
			client.FILE_SHARE_READ|client.FILE_SHARE_WRITE,
			client.FILE_OVERWRITE_IF,
			client.FILE_NON_DIRECTORY_FILE,
		)
		if err != nil {
			fmt.Printf("[error] Error opening file %q for write: %s\n", filePath, err)
			return
		}
		n, err := c.WriteFile(wfid, 0, []byte(writeContent))
		if err != nil {
			fmt.Printf("[error] Error writing file %q: %s\n", filePath, err)
			_ = c.CloseFile(wfid)
			return
		}
		fmt.Printf("[info] Wrote %d bytes to [%s]\n", n, filePath)
		if err = c.CloseFile(wfid); err != nil {
			fmt.Printf("[warn] Error closing file %q after write: %s\n", filePath, err)
		}
	}

	// Open the file at the root of the share for reading.
	fid, err := c.OpenFile(
		filePath,
		client.GENERIC_READ,
		client.FILE_SHARE_READ|client.FILE_SHARE_WRITE,
		client.FILE_OPEN,
		client.FILE_NON_DIRECTORY_FILE,
	)
	if err != nil {
		fmt.Printf("[error] Error opening file %q: %s\n", filePath, err)
		return
	}
	fmt.Printf("[info] Opened file [%s] (FID 0x%04x)\n", filePath, uint16(fid))

	// Read up to 1 MiB from the file.
	contents, err := c.ReadFile(fid, 0, 1024*1024)
	if err != nil {
		fmt.Printf("[error] Error reading file %q: %s\n", filePath, err)
		_ = c.CloseFile(fid)
		return
	}
	fmt.Printf("[info] Read %d bytes from [%s]\n", len(contents), filePath)

	if err = c.CloseFile(fid); err != nil {
		fmt.Printf("[warn] Error closing file %q: %s\n", filePath, err)
	}

	fmt.Printf("----- %s -----\n%s\n----- end -----\n", filePath, string(contents))
}
