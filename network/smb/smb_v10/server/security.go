package server

import (
	"fmt"

	"github.com/TheManticoreProject/winacl/ace"
	aceflags "github.com/TheManticoreProject/winacl/ace/aceflags"
	acetype "github.com/TheManticoreProject/winacl/ace/acetype"
	aceheader "github.com/TheManticoreProject/winacl/ace/header"
	acemask "github.com/TheManticoreProject/winacl/ace/mask"
	"github.com/TheManticoreProject/winacl/acl"
	aclrevision "github.com/TheManticoreProject/winacl/acl/revision"
	"github.com/TheManticoreProject/winacl/identity"
	"github.com/TheManticoreProject/winacl/securitydescriptor"
	sdcontrol "github.com/TheManticoreProject/winacl/securitydescriptor/control"
	"github.com/TheManticoreProject/winacl/sid"
)

// SecurityInformation selects which parts of a security descriptor a query or a set
// applies to. The values are the SECURITY_INFORMATION bits from [MS-DTYP] 2.4.7.
type SecurityInformation uint32

const (
	// OwnerSecurityInformation selects the owner.
	OwnerSecurityInformation SecurityInformation = 0x00000001
	// GroupSecurityInformation selects the primary group.
	GroupSecurityInformation SecurityInformation = 0x00000002
	// DaclSecurityInformation selects the discretionary ACL.
	DaclSecurityInformation SecurityInformation = 0x00000004
	// SaclSecurityInformation selects the system ACL.
	SaclSecurityInformation SecurityInformation = 0x00000008
)

// SecurityProvider supplies the security descriptors for a share.
//
// A share without one refuses the security-descriptor subcommands with
// STATUS_NOT_SUPPORTED, which is the honest answer: a client that receives a
// descriptor believes it describes the access the server enforces, so inventing one
// is worse than admitting there is no model to describe.
type SecurityProvider interface {
	// SecurityDescriptor returns the self-relative descriptor for a path, with
	// only the parts information selects populated.
	SecurityDescriptor(path string, information SecurityInformation) ([]byte, error)

	// SetSecurityDescriptor applies a descriptor to a path.
	SetSecurityDescriptor(path string, information SecurityInformation, descriptor []byte) error
}

// Well-known SIDs used by the reflective provider below.
const (
	// sidAuthenticatedUsers is S-1-5-11, every identity that authenticated.
	sidAuthenticatedUsers = "S-1-5-11"
	// sidLocalSystem is S-1-5-18, which owns what the server itself created.
	sidLocalSystem = "S-1-5-18"
	// sidAdministrators is S-1-5-32-544, the local administrators group.
	sidAdministrators = "S-1-5-32-544"
)

// Access-mask values used in the descriptor built below.
const (
	// accessRead is what reading needs: the data, the extended attributes and the
	// attributes, plus READ_CONTROL and SYNCHRONIZE, which every open carries.
	accessRead uint32 = 0x00120089
	// accessWrite is the rights writing adds on top. It holds no bit that
	// accessRead holds, so a mask can be tested for one without matching the
	// other — which a combined constant could not be, since both would carry
	// READ_CONTROL and SYNCHRONIZE.
	accessWrite uint32 = 0x00000116
)

// ReflectiveSecurityProvider describes the access the server actually enforces,
// rather than an access-control model it does not have.
//
// The descriptor it returns says: authenticated users may read, and may also write
// unless the share is read-only. That is exactly the rule the handlers apply, so
// the descriptor is truthful — which matters because a client uses a descriptor to
// predict what it will be allowed to do, and a descriptor describing rights the
// server does not honour, or withholding ones it does, makes the client wrong in
// one direction or the other.
//
// It is deliberately not a stand-in for real per-file ACLs. A caller with a real
// model implements SecurityProvider itself.
type ReflectiveSecurityProvider struct {
	// readOnly mirrors the share's own flag, which is what decides whether write
	// access is described.
	readOnly bool
}

// NewReflectiveSecurityProvider creates a provider describing a share whose
// read-only state is given.
//
// Parameters:
//   - readOnly: whether the share refuses modification
//
// Returns:
//   - The provider
func NewReflectiveSecurityProvider(readOnly bool) *ReflectiveSecurityProvider {
	return &ReflectiveSecurityProvider{readOnly: readOnly}
}

// SecurityDescriptor builds the descriptor for a path.
//
// The path does not affect the answer: the server's access rule is per share, not
// per file, and pretending otherwise by varying the descriptor would suggest a
// granularity that does not exist.
func (p *ReflectiveSecurityProvider) SecurityDescriptor(path string, information SecurityInformation) ([]byte, error) {
	descriptor := &securitydescriptor.NtSecurityDescriptor{}
	descriptor.Header.Revision = 1
	// SE_DACL_PRESENT plus SE_SELF_RELATIVE: the form a client asks for, where
	// every component is reached by an offset from the start rather than by a
	// pointer.
	descriptor.Header.Control = sdcontrol.NtSecurityDescriptorControl{
		RawValue: sdcontrol.NT_SECURITY_DESCRIPTOR_CONTROL_DP |
			sdcontrol.NT_SECURITY_DESCRIPTOR_CONTROL_SR,
	}

	if information&OwnerSecurityInformation != 0 {
		owner, err := identityFor(sidLocalSystem)
		if err != nil {
			return nil, err
		}
		descriptor.Owner = owner
	}
	if information&GroupSecurityInformation != 0 {
		group, err := identityFor(sidAdministrators)
		if err != nil {
			return nil, err
		}
		descriptor.Group = group
	}

	if information&DaclSecurityInformation != 0 {
		mask := accessRead
		if !p.readOnly {
			mask |= accessWrite
		}
		entry, err := allowEntry(sidAuthenticatedUsers, mask)
		if err != nil {
			return nil, err
		}
		entries := []ace.AccessControlEntry{*entry}
		descriptor.DACL = &acl.DiscretionaryAccessControlList{
			Header: acl.DiscretionaryAccessControlListHeader{
				// ACL_REVISION, the revision for an ACL holding no
				// object-specific ACEs.
				Revision: aclrevision.AccessControlListRevision{Value: 2},
				// The marshaller derives AclSize from what it writes but takes
				// AceCount as given, so a count left at zero produces an ACL
				// whose entries every parser skips.
				AceCount: uint16(len(entries)),
			},
			Entries: entries,
		}
	}

	// A SACL is never produced: nothing here audits, so an empty one would claim
	// auditing is configured and a populated one would be a fabrication.

	marshalled, err := descriptor.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal the security descriptor: %v", err)
	}
	return marshalled, nil
}

// SetSecurityDescriptor refuses every change.
//
// The descriptor this provider returns is derived from the share's configuration,
// so there is nowhere to store a different one. Accepting a change and then
// continuing to report the derived descriptor would be worse than refusing: a
// client would believe it had applied something it had not.
func (p *ReflectiveSecurityProvider) SetSecurityDescriptor(path string, information SecurityInformation, descriptor []byte) error {
	return fmt.Errorf("this share's security descriptor is derived from its configuration: %w", ErrAccessDenied)
}

// identityFor builds an identity from a SID string.
func identityFor(value string) (*identity.Identity, error) {
	parsed := &sid.SID{}
	if err := parsed.FromString(value); err != nil {
		return nil, fmt.Errorf("failed to parse the SID %q: %v", value, err)
	}
	return &identity.Identity{SID: *parsed}, nil
}

// allowEntry builds an access-allowed ACE granting a mask to a SID.
func allowEntry(value string, mask uint32) (*ace.AccessControlEntry, error) {
	holder, err := identityFor(value)
	if err != nil {
		return nil, err
	}

	entry := &ace.AccessControlEntry{
		Header: aceheader.AccessControlEntryHeader{
			Type:  acetype.AccessControlEntryType{Value: acetype.ACE_TYPE_ACCESS_ALLOWED},
			Flags: aceflags.AccessControlEntryFlag{RawValue: aceflags.ACE_FLAG_NONE},
		},
		Mask:     acemask.AccessControlMask{RawValue: mask},
		Identity: *holder,
	}
	// The header size is the fixed part plus the SID, which the marshaller reads
	// rather than deriving, so it has to be right or the entry is unparseable.
	entry.Header.Size = uint16(4 + 4 + holder.SID.RawBytesSize)
	return entry, nil
}

// Compile-time assurance that the provider satisfies the contract.
var _ SecurityProvider = (*ReflectiveSecurityProvider)(nil)
