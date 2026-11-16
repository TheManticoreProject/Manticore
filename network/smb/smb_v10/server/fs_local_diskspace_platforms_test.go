package server

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestLocalFileSystemDiskUsageBuildsOnEveryPlatform guards the platform split of
// the LocalFileSystem volume queries (diskUsage, hostVolumeSerial). The Unix
// implementation calls statfs, which golang.org/x/sys/unix does not provide on
// js/wasm, wasip1/wasm, plan9, netbsd, solaris or illumos, and whose Statfs_t
// on openbsd has other field names; those targets need their own fallback so the
// SMB server packages keep compiling there.
func TestLocalFileSystemDiskUsageBuildsOnEveryPlatform(t *testing.T) {
	if testing.Short() {
		t.Skip("cross-compiles the package; skipped in -short mode")
	}
	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go tool not found in PATH")
	}
	targets := []struct{ goos, goarch string }{
		{"linux", "amd64"},
		{"linux", "386"},
		{"darwin", "arm64"},
		{"freebsd", "amd64"},
		{"windows", "amd64"},
		{"js", "wasm"},
		{"wasip1", "wasm"},
		{"plan9", "amd64"},
		{"openbsd", "amd64"},
		{"netbsd", "amd64"},
		{"solaris", "amd64"},
		{"illumos", "amd64"},
		{"aix", "ppc64"},
		{"dragonfly", "amd64"},
	}
	for _, tc := range targets {
		tc := tc
		t.Run(tc.goos+"_"+tc.goarch, func(t *testing.T) {
			cmd := exec.Command(goTool, "build", "-o", os.DevNull, ".")
			cmd.Env = append(os.Environ(), "GOOS="+tc.goos, "GOARCH="+tc.goarch, "CGO_ENABLED=0")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("GOOS=%s GOARCH=%s go build failed: %v\n%s", tc.goos, tc.goarch, err, strings.TrimSpace(string(out)))
			}
		})
	}
}
