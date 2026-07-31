package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"software.sslmate.com/src/go-pkcs12"
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

func TestExportCertificatePEMWritesParsableCertificate(t *testing.T) {
	cert := newTestCertificate(t)
	path := filepath.Join(t.TempDir(), "cert.pem")

	if err := cert.ExportCertificatePEM(path); err != nil {
		t.Fatalf("ExportCertificatePEM() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}

	block, rest := pem.Decode(data)
	if block == nil {
		t.Fatal("exported certificate is not a PEM block")
	}
	if block.Type != "CERTIFICATE" {
		t.Errorf("PEM block type = %q, want %q", block.Type, "CERTIFICATE")
	}
	if len(rest) != 0 {
		t.Errorf("unexpected trailing data after PEM block: %d bytes", len(rest))
	}

	parsed, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("x509.ParseCertificate() error = %v", err)
	}

	original, err := cert.GetCertificate()
	if err != nil {
		t.Fatalf("GetCertificate() error = %v", err)
	}
	if parsed.SerialNumber.Cmp(original.SerialNumber) != 0 {
		t.Errorf("serial number = %v, want %v", parsed.SerialNumber, original.SerialNumber)
	}
	if parsed.Subject.CommonName != "TESTUSER$" {
		t.Errorf("subject CN = %q, want %q", parsed.Subject.CommonName, "TESTUSER$")
	}
}

func TestExportPFXRoundTrips(t *testing.T) {
	cert := newTestCertificate(t)
	path := filepath.Join(t.TempDir(), "keys", "bundle.pfx")
	const password = "Passw0rd!"

	if err := cert.ExportPFX(path, password); err != nil {
		t.Fatalf("ExportPFX() error = %v", err)
	}

	pfxData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}

	decodedKey, decodedCert, err := pkcs12.Decode(pfxData, password)
	if err != nil {
		t.Fatalf("pkcs12.Decode() error = %v", err)
	}

	original, err := cert.GetCertificate()
	if err != nil {
		t.Fatalf("GetCertificate() error = %v", err)
	}
	if decodedCert.SerialNumber.Cmp(original.SerialNumber) != 0 {
		t.Errorf("decoded serial number = %v, want %v", decodedCert.SerialNumber, original.SerialNumber)
	}

	originalKey, err := cert.GetRSAPrivateKey()
	if err != nil {
		t.Fatalf("GetRSAPrivateKey() error = %v", err)
	}
	rsaKey, ok := decodedKey.(*rsa.PrivateKey)
	if !ok {
		t.Fatalf("decoded private key type = %T, want *rsa.PrivateKey", decodedKey)
	}
	if !rsaKey.Equal(originalKey) {
		t.Error("decoded private key does not match the certificate's private key")
	}
}

func TestExportPFXRejectsWrongPassword(t *testing.T) {
	cert := newTestCertificate(t)
	path := filepath.Join(t.TempDir(), "bundle.pfx")

	if err := cert.ExportPFX(path, "correct-horse"); err != nil {
		t.Fatalf("ExportPFX() error = %v", err)
	}

	pfxData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}

	if _, _, err := pkcs12.Decode(pfxData, "wrong-password"); err == nil {
		t.Error("pkcs12.Decode() with a wrong password succeeded, want error")
	}
}

func TestExportPFXUsesRestrictivePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes are not enforced on Windows")
	}

	cert := newTestCertificate(t)
	dir := filepath.Join(t.TempDir(), "keys")
	path := filepath.Join(dir, "bundle.pfx")

	if err := cert.ExportPFX(path, "Passw0rd!"); err != nil {
		t.Fatalf("ExportPFX() error = %v", err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat(%q) error = %v", path, err)
	}
	if got := fi.Mode().Perm(); got != privateKeyFileMode {
		t.Errorf("PFX file mode = %04o, want %04o", got, privateKeyFileMode)
	}

	di, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("os.Stat(%q) error = %v", dir, err)
	}
	if got := di.Mode().Perm(); got != privateKeyDirMode {
		t.Errorf("PFX directory mode = %04o, want %04o", got, privateKeyDirMode)
	}
}
