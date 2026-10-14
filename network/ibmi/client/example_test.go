package client_test

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/TheManticoreProject/Manticore/ibmi/api/users"
	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/client"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd"
)

// Example lists IBM i user profiles: it builds a credential, connects over
// verified TLS (or explicit cleartext on an isolated lab), runs DSPJOB,
// lists profiles in format AUTU0150, and closes. It needs a real system, so
// it is compiled but not run by go test.
func Example() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 1. A credential owns the profile and password; formatting it prints a
	// redacted marker. Use a least-privileged profile, never QSECOFR.
	password := []byte(os.Getenv("IBMI_PASSWORD"))
	cred, err := ibminet.NewPasswordCredential(os.Getenv("IBMI_USER"), password)
	clear(password)
	if err != nil {
		log.Fatal(err)
	}
	defer cred.Destroy()

	// 2. Choose the transport explicitly. TLS is always verified and never
	// falls back to cleartext.
	cfg := client.Config{Host: "ibmi.example.com", Transport: ibminet.TransportTLSRequired}
	if pemFile := os.Getenv("IBMI_CA_FILE"); pemFile != "" {
		pem, err := os.ReadFile(pemFile)
		if err != nil {
			log.Fatal(err)
		}
		roots := x509.NewCertPool()
		roots.AppendCertsFromPEM(pem)
		cfg.TLS = hostserver.TLSOptions{RootCAs: roots}
	}
	if os.Getenv("IBMI_LAB_CLEARTEXT") == "1" {
		cfg.Transport = ibminet.TransportCleartextExplicit // isolated lab only
	}

	// 3. Connect: Sign-on, then an authenticated Remote Command server job.
	// A rejected credential is returned, never retried.
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := c.Connect(ctx, cred); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			log.Printf("close: %v", err)
		}
	}()
	info := c.SessionInfo()
	fmt.Printf("%s, password level %d, system CCSID %d, job CCSID %d\n", info.VRM, info.PasswordLevel, info.SystemCCSID, info.JobCCSID)

	// 4. Run DSPJOB and inspect the structured IBM messages. RunCommand
	// executes raw CL: never build commands from untrusted input.
	res, err := c.RunCommand(ctx, "DSPJOB OUTPUT(*PRINT)", rmtcmd.RunCommandOptions{})
	switch {
	case errors.Is(err, rmtcmd.ErrUncertainOutcome):
		log.Fatalf("DSPJOB may or may not have run: %v", err) // never replayed
	case err != nil && !errors.Is(err, rmtcmd.ErrNative):
		log.Fatal(err)
	}
	for _, m := range res.Messages {
		fmt.Println(m) // ID and severity only; m.Text may hold sensitive data
	}

	// 5. List profiles with the typed users client, over the same session.
	u, err := users.New(c, users.Options{SystemCCSID: info.SystemCCSID, JobCCSID: info.JobCCSID})
	if err != nil {
		log.Fatal(err)
	}
	list, err := u.List(ctx, users.ListOptions{PageSize: 100, MaxUsers: 5000})
	if err != nil && (list == nil || !list.Partial) {
		log.Fatal(err)
	}

	// 6. Inspect the result: partial data comes only with an error, and
	// IBM i returns only the profiles the caller may read.
	if list.Partial {
		log.Printf("partial list (%d of %d): %v", len(list.Users), list.TotalReported, err)
	}
	for _, w := range list.Warnings {
		log.Print(w)
	}
	for _, p := range list.Users {
		fmt.Printf("%-10s group=%t %s\n", p.Name, p.GroupProfile, p.Description)
	}
}
