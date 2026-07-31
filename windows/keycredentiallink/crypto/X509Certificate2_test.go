package crypto

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// newTestCertificate builds a small certificate for the export tests. 1024 bits
// keeps key generation fast; the tests only exercise file handling.
func newTestCertificate(t *testing.T) *X509Certificate {
	t.Helper()

	notBefore := time.Now()
	cert, err := NewX509Certificate("TESTUSER$", 1024, notBefore, notBefore.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("NewX509Certificate() error = %v", err)
	}
	return cert
}

func TestExportRSAPrivateKeyPEMUsesRestrictivePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes are not enforced on Windows")
	}

	cert := newTestCertificate(t)

	// A nested path so that the directory is created by the export itself.
	dir := filepath.Join(t.TempDir(), "keys", "nested")
	path := filepath.Join(dir, "private.pem")

	if err := cert.ExportRSAPrivateKeyPEM(path); err != nil {
		t.Fatalf("ExportRSAPrivateKeyPEM() error = %v", err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat(%q) error = %v", path, err)
	}
	if got := fi.Mode().Perm(); got != privateKeyFileMode {
		t.Errorf("private key file mode = %04o, want %04o", got, privateKeyFileMode)
	}

	di, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("os.Stat(%q) error = %v", dir, err)
	}
	if got := di.Mode().Perm(); got != privateKeyDirMode {
		t.Errorf("private key directory mode = %04o, want %04o", got, privateKeyDirMode)
	}
}

func TestExportRSAPrivateKeyPEMNarrowsExistingFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes are not enforced on Windows")
	}

	cert := newTestCertificate(t)
	path := filepath.Join(t.TempDir(), "private.pem")

	// O_CREATE does not change the mode of a file that already exists, so an
	// export over a world-readable file has to narrow it explicitly.
	if err := os.WriteFile(path, []byte("stale"), 0o666); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
	if err := os.Chmod(path, 0o666); err != nil {
		t.Fatalf("os.Chmod() error = %v", err)
	}

	if err := cert.ExportRSAPrivateKeyPEM(path); err != nil {
		t.Fatalf("ExportRSAPrivateKeyPEM() error = %v", err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat(%q) error = %v", path, err)
	}
	if got := fi.Mode().Perm(); got != privateKeyFileMode {
		t.Errorf("private key file mode = %04o, want %04o", got, privateKeyFileMode)
	}
}

func TestExportRSAPrivateKeyPEMWritesParsablePEM(t *testing.T) {
	cert := newTestCertificate(t)
	path := filepath.Join(t.TempDir(), "private.pem")

	if err := cert.ExportRSAPrivateKeyPEM(path); err != nil {
		t.Fatalf("ExportRSAPrivateKeyPEM() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	if len(data) == 0 {
		t.Fatal("exported private key is empty")
	}
	if want := "-----BEGIN RSA PRIVATE KEY-----"; string(data[:len(want)]) != want {
		t.Errorf("exported private key does not start with %q", want)
	}
}
