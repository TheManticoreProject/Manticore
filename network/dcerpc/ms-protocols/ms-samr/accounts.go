package ms_samr

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/functions"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// MaximumAllowed asks the server to grant every access right the caller is entitled to
// ([MS-DTYP] 2.4.3). It is the simplest mask for read-only enumeration and what the
// high-level helpers request.
const MaximumAllowed uint32 = 0x02000000

// Well-known alias RIDs in the Builtin domain ([MS-DTYP] 2.4.2.4).
const (
	AliasRIDAdministrators uint32 = 544
	AliasRIDUsers          uint32 = 545
)

// BuiltinDomainSID is the SID of the Builtin domain (S-1-5-32), whose aliases include the
// local Administrators and Users groups. SamrOpenDomain takes this SID directly, so no
// SamrLookupDomainInSamServer round trip is needed to reach the local groups.
func BuiltinDomainSID() (msdtyp.RPC_SID, error) {
	return msdtyp.ParseSID("S-1-5-32")
}

// OpenServer connects to the SAM server and returns a server handle (SamrConnect5). The
// desired access is typically MaximumAllowed. The handle is closed with CloseHandle.
func (c *Client) OpenServer(desiredAccess uint32) (Handle, error) {
	if err := c.ensure(); err != nil {
		return Handle{}, err
	}
	// SamrConnect5 is the modern connect call; the server name is advisory and ignored by
	// Windows, so an empty name is sent. The version/revision outputs are not needed.
	h, _, _, err := functions.SamrConnect5(c.rpc, "", desiredAccess)
	if err != nil {
		return Handle{}, fmt.Errorf("ms_samr: OpenServer: %w", err)
	}
	return h, nil
}

// LookupDomain resolves a domain name (for example "Builtin" or the host's account domain)
// to its SID under a server handle (SamrLookupDomainInSamServer).
func (c *Client) LookupDomain(server Handle, name string) (msdtyp.RPC_SID, error) {
	if err := c.ensure(); err != nil {
		return msdtyp.RPC_SID{}, err
	}
	sid, err := functions.SamrLookupDomainInSamServer(c.rpc, server, name)
	if err != nil {
		return msdtyp.RPC_SID{}, fmt.Errorf("ms_samr: LookupDomain %q: %w", name, err)
	}
	if sid == nil {
		return msdtyp.RPC_SID{}, fmt.Errorf("ms_samr: LookupDomain %q: server returned no SID", name)
	}
	return *sid, nil
}

// OpenDomain returns a handle to the domain identified by sid under a server handle
// (SamrOpenDomain). For the local groups, pass BuiltinDomainSID.
func (c *Client) OpenDomain(server Handle, sid msdtyp.RPC_SID, desiredAccess uint32) (Handle, error) {
	if err := c.ensure(); err != nil {
		return Handle{}, err
	}
	h, err := functions.SamrOpenDomain(c.rpc, server, desiredAccess, sid)
	if err != nil {
		return Handle{}, fmt.Errorf("ms_samr: OpenDomain: %w", err)
	}
	return h, nil
}

// OpenAlias returns a handle to the alias with the given RID under a domain handle
// (SamrOpenAlias). RID 544 is Administrators, 545 is Users.
func (c *Client) OpenAlias(domain Handle, aliasRID, desiredAccess uint32) (Handle, error) {
	if err := c.ensure(); err != nil {
		return Handle{}, err
	}
	h, err := functions.SamrOpenAlias(c.rpc, domain, desiredAccess, aliasRID)
	if err != nil {
		return Handle{}, fmt.Errorf("ms_samr: OpenAlias %d: %w", aliasRID, err)
	}
	return h, nil
}

// GetMembersInAlias returns the SIDs of the members of an alias (SamrGetMembersInAlias),
// as their string form (for example "S-1-5-21-...-500").
func (c *Client) GetMembersInAlias(alias Handle) ([]string, error) {
	if err := c.ensure(); err != nil {
		return nil, err
	}
	out, err := functions.SamrGetMembersInAlias(c.rpc, alias)
	if err != nil {
		return nil, fmt.Errorf("ms_samr: GetMembersInAlias: %w", err)
	}
	sids := make([]string, 0, len(out.Sids))
	for _, m := range out.Sids {
		if m.SidPointer == nil {
			continue
		}
		sids = append(sids, m.SidPointer.String())
	}
	return sids, nil
}

// OpenGroup returns a handle to the group with the given RID under a domain handle
// (SamrOpenGroup).
func (c *Client) OpenGroup(domain Handle, groupRID, desiredAccess uint32) (Handle, error) {
	if err := c.ensure(); err != nil {
		return Handle{}, err
	}
	h, err := functions.SamrOpenGroup(c.rpc, domain, desiredAccess, groupRID)
	if err != nil {
		return Handle{}, fmt.Errorf("ms_samr: OpenGroup %d: %w", groupRID, err)
	}
	return h, nil
}

// GetMembersInGroup returns the RIDs of the members of a group (SamrGetMembersInGroup).
// Group members are domain-relative RIDs, not SIDs; resolve them with LookupIds if names
// are needed.
func (c *Client) GetMembersInGroup(group Handle) ([]uint32, error) {
	if err := c.ensure(); err != nil {
		return nil, err
	}
	buf, err := functions.SamrGetMembersInGroup(c.rpc, group)
	if err != nil {
		return nil, fmt.Errorf("ms_samr: GetMembersInGroup: %w", err)
	}
	if buf == nil {
		return nil, nil
	}
	rids := make([]uint32, len(buf.Members))
	for i, m := range buf.Members {
		rids[i] = uint32(m)
	}
	return rids, nil
}

// CloseHandle closes a SAM handle (SamrCloseHandle).
func (c *Client) CloseHandle(h Handle) error {
	if err := c.ensure(); err != nil {
		return err
	}
	if _, err := functions.SamrCloseHandle(c.rpc, h); err != nil {
		return fmt.Errorf("ms_samr: CloseHandle: %w", err)
	}
	return nil
}

// LocalAliasMembers enumerates the members of a Builtin-domain alias and returns their SIDs
// as strings. With aliasRID 544 (AliasRIDAdministrators) it is the enum-local-admins
// primitive: the SIDs that are local administrators of the target. It opens the SAM server,
// the Builtin domain (S-1-5-32), the alias, reads its members, and closes every handle.
func (c *Client) LocalAliasMembers(aliasRID uint32) ([]string, error) {
	if err := c.ensure(); err != nil {
		return nil, err
	}
	server, err := c.OpenServer(MaximumAllowed)
	if err != nil {
		return nil, err
	}
	defer c.CloseHandle(server)

	builtin, err := BuiltinDomainSID()
	if err != nil {
		return nil, err
	}
	domain, err := c.OpenDomain(server, builtin, MaximumAllowed)
	if err != nil {
		return nil, err
	}
	defer c.CloseHandle(domain)

	alias, err := c.OpenAlias(domain, aliasRID, MaximumAllowed)
	if err != nil {
		return nil, err
	}
	defer c.CloseHandle(alias)

	return c.GetMembersInAlias(alias)
}
