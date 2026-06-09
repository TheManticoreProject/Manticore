package client_test

import (
	"fmt"
	"log"

	"github.com/TheManticoreProject/Manticore/network/smb"
	smbclient "github.com/TheManticoreProject/Manticore/network/smb/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// ExampleDial shows a full session: negotiate the best mutually-supported
// dialect, authenticate, connect to a share, and read a file — without the
// caller having to know whether SMB1 or SMB2 was selected.
func ExampleDial() {
	creds, err := credentials.NewCredentials("", "Administrator", "secret", "")
	if err != nil {
		log.Fatal(err)
	}

	// Empty Options: offer all supported versions, best first.
	c, err := smbclient.Dial("fileserver", 445, smbclient.Options{})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Disconnect()

	if err := c.Login(creds); err != nil {
		log.Fatal(err)
	}
	fmt.Println("negotiated", c.Dialect())

	if err := c.TreeConnect("C$"); err != nil {
		log.Fatal(err)
	}
	defer c.TreeDisconnect()

	h, err := c.OpenFile(`Windows\win.ini`, smbclient.OpenOptions{
		DesiredAccess:     0x80000000, // GENERIC_READ
		ShareAccess:       0x00000001, // FILE_SHARE_READ
		CreateDisposition: 0x00000001, // FILE_OPEN
		CreateOptions:     0x00000040, // FILE_NON_DIRECTORY_FILE
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.CloseFile(h)

	data, err := c.ReadFile(h, 0, 4096)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("read %d bytes\n", len(data))
}

// ExampleDial_preferenceOrder forces SMB1, falling back to SMB 2.0.2. With the
// default PolicyStrictOrder the client selects the first version the server
// accepts, honoring the listed order even when the server supports both.
func ExampleDial_preferenceOrder() {
	c, err := smbclient.Dial("fileserver", 445, smbclient.Options{
		Preferred: []smb.SMBProtocolVersion{
			smb.SMB_VERSION_1_0,
			smb.SMB_VERSION_2_0_2,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Disconnect()
	fmt.Println("negotiated", c.Dialect())
}

// ExampleClient_ListDirectory enumerates a directory; the returned FileInfo
// values are the same regardless of the negotiated dialect.
func ExampleClient_ListDirectory() {
	var c *smbclient.Client // obtained from Dial + Login + TreeConnect

	entries, err := c.ListDirectory("Windows", "*")
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range entries {
		fmt.Printf("%s\tdir=%v\tsize=%d\n", e.Name, e.IsDir(), e.Size)
	}
}
