package fingerprint

import (
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/capabilities"
)

// TestDefaultNamesNothingIdentifying is the guard this package exists for. Each
// of these strings reaches the wire, and one naming the project describes the
// implementation to an observer before any behaviour is examined. "Manticore"
// and "MANTICORE" were both emitted before this package existed.
func TestDefaultNamesNothingIdentifying(t *testing.T) {
	profile := Default()

	fields := map[string]string{
		"NativeOS":            profile.NativeOS,
		"NativeLanMan":        profile.NativeLanMan,
		"CallingNameFallback": profile.CallingNameFallback,
	}
	banned := []string{"manticore", "samba", "unix", "golang", "gopher"}

	for name, value := range fields {
		if value == "" {
			t.Errorf("%s is empty; strict peers reject an empty native string", name)
		}
		lowered := strings.ToLower(value)
		for _, word := range banned {
			if strings.Contains(lowered, word) {
				t.Errorf("%s = %q contains %q", name, value, word)
			}
		}
	}
}

// TestDefaultIdentityIsSelfConsistent checks the native pair describes one
// product. Two independently plausible strings that disagree about what is
// running are their own tell.
func TestDefaultIdentityIsSelfConsistent(t *testing.T) {
	profile := Default()

	if !strings.HasPrefix(profile.NativeOS, "Windows ") || !strings.HasPrefix(profile.NativeLanMan, "Windows ") {
		t.Fatalf("native strings do not both name a Windows product: %q / %q", profile.NativeOS, profile.NativeLanMan)
	}
	// The product name is everything up to the trailing version token.
	product := func(s string) string {
		i := strings.LastIndex(s, " ")
		if i < 0 {
			return s
		}
		return s[:i]
	}
	if product(profile.NativeOS) != product(profile.NativeLanMan) {
		t.Errorf("native strings describe different products: %q vs %q",
			product(profile.NativeOS), product(profile.NativeLanMan))
	}
}

// TestDefaultNTLMVersionIsPlausible checks the VERSION structure could belong to
// a real host. A build number of zero, which this implementation emitted for a
// long time, corresponds to no Windows release that has ever shipped.
func TestDefaultNTLMVersionIsPlausible(t *testing.T) {
	v := Default().NTLMVersion

	if v.ProductBuild == 0 {
		t.Error("NTLMVersion.ProductBuild is 0, which is not a real Windows build")
	}
	if v.ProductMajorVersion == 0 {
		t.Error("NTLMVersion.ProductMajorVersion is 0")
	}
	if v.NTLMRevision == 0 {
		t.Error("NTLMVersion.NTLMRevision is 0")
	}
}

// TestDefaultDialectsAscend checks the offered list is ordered as [MS-SMB2]
// 2.2.3 requires, and that the list is not empty.
func TestDefaultDialectsAscend(t *testing.T) {
	dialects := Default().Dialects
	if len(dialects) == 0 {
		t.Fatal("no dialects offered")
	}
	for i := 1; i < len(dialects); i++ {
		if dialects[i] <= dialects[i-1] {
			t.Errorf("dialect %d (%s) does not follow %s in ascending order",
				i, dialects[i], dialects[i-1])
		}
	}
}

// TestDefaultClaimsNoUnimplementedCapability is the counterweight to any future
// urge to widen the capability mask toward what Windows advertises. Claiming a
// capability that is then refused is a louder difference than not claiming it,
// so each bit here must be backed by an implementation.
func TestDefaultClaimsNoUnimplementedCapability(t *testing.T) {
	unimplemented := map[string]capabilities.Capabilities{
		"SMB2_GLOBAL_CAP_DFS":                capabilities.SMB2_GLOBAL_CAP_DFS,
		"SMB2_GLOBAL_CAP_MULTI_CHANNEL":      capabilities.SMB2_GLOBAL_CAP_MULTI_CHANNEL,
		"SMB2_GLOBAL_CAP_PERSISTENT_HANDLES": capabilities.SMB2_GLOBAL_CAP_PERSISTENT_HANDLES,
	}
	claimed := Default().Capabilities
	for name, bit := range unimplemented {
		if claimed&bit != 0 {
			t.Errorf("the default profile claims %s, which this client does not implement", name)
		}
	}
}

// TestCloneIsIndependent checks a caller adjusting a profile cannot reach back
// into the package default, which every other connection uses.
func TestCloneIsIndependent(t *testing.T) {
	first := Default()
	first.Dialects[0] = 0xFFFF
	first.Ciphers[0] = 0xFFFF
	first.NativeOS = "changed"

	second := Default()
	if second.Dialects[0] == 0xFFFF {
		t.Error("mutating a profile's dialects changed the package default")
	}
	if second.Ciphers[0] == 0xFFFF {
		t.Error("mutating a profile's ciphers changed the package default")
	}
	if second.NativeOS == "changed" {
		t.Error("mutating a profile's NativeOS changed the package default")
	}
}

// TestMachineGuidIsStableAndNonZero checks the ClientGuid source. Zero is what
// the client emitted before this package; a value that changed per call would be
// as wrong, since the field identifies the client.
func TestMachineGuidIsStableAndNonZero(t *testing.T) {
	first := MachineGuid()
	if first == ([16]byte{}) {
		t.Fatal("MachineGuid is all zero")
	}
	if MachineGuid() != first {
		t.Error("MachineGuid changed between calls")
	}
}

// TestProvenanceIsRecorded checks each profile says where its values come from.
// A value nobody can source is a value nobody should pin a test to.
func TestProvenanceIsRecorded(t *testing.T) {
	if strings.TrimSpace(Default().Provenance) == "" {
		t.Error("the default profile records no provenance")
	}
	if strings.TrimSpace(Default().Name) == "" {
		t.Error("the default profile has no name")
	}
}
