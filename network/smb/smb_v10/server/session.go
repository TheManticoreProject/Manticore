package server

import (
	"fmt"
	"time"
)

// Session-setup Action bits, returned in the final session-setup response
// ([MS-CIFS] 2.2.4.53.2).
const (
	// SMB_SETUP_GUEST reports that the logon was mapped to the guest account
	// rather than to the identity the client claimed. A client is entitled to
	// treat that as a failure, which is why it must be reported rather than
	// silently granted.
	SMB_SETUP_GUEST = 0x0001

	// SMB_SETUP_USE_LANMAN_KEY reports that the client's LM key is in use for
	// signing rather than a derived session key. This server never does that.
	SMB_SETUP_USE_LANMAN_KEY = 0x0002
)

// Identifier bounds. Zero and 0xFFFF are both reserved: zero means "none yet" on
// a request, and 0xFFFF is the wildcard several commands use.
const (
	minIdentifier = 0x0001
	maxIdentifier = 0xFFFE
)

// Session is an authenticated session on a connection.
//
// A session in the table has been authenticated: either its response verified
// against a credential, or it was admitted under an explicit guest or anonymous
// policy. Which of the three it was is recorded, because it decides what the
// session may go on to do — an anonymous session has no key, so it cannot sign.
type Session struct {
	// UID is the identifier the client sends on every request in this session.
	UID uint16

	// Domain, Username and Workstation are the identity that was authenticated,
	// or the identity that was claimed when the session is a guest one.
	Domain      string
	Username    string
	Workstation string

	// SessionKey is the exported session key derived from the authentication. It
	// is the MAC key for signing, and is nil for a guest or anonymous session,
	// where no key was derived.
	SessionKey []byte

	// IsGuest reports that the claimed identity was not verified and the session
	// was admitted as a guest.
	IsGuest bool

	// IsAnonymous reports a null session: no identity was claimed at all.
	IsAnonymous bool

	// Created is when the session was established.
	Created time.Time
}

// Account renders the session's identity as DOMAIN\user, or just the username
// when no domain was claimed.
func (s *Session) Account() string {
	if s.IsAnonymous {
		return "<anonymous>"
	}
	if s.Domain == "" {
		return s.Username
	}
	return s.Domain + "\\" + s.Username
}

// CanSign reports whether the session has key material to sign with. A guest or
// anonymous session does not, which is why signing cannot be required of one.
func (s *Session) CanSign() bool {
	return len(s.SessionKey) > 0
}

// identifierAllocator hands out 16-bit protocol identifiers and takes them back.
//
// Identifiers are reused once released, because the space is only 16 bits wide and
// a long-lived connection that opened and closed handles without reuse would
// exhaust it. Reuse is safe here because an identifier is only ever released when
// the thing it named is gone.
type identifierAllocator struct {
	next  uint16
	freed []uint16
	inUse int
	limit int
}

// newIdentifierAllocator creates an allocator bounded to limit live identifiers
// at once. A limit of zero or less bounds it only by the 16-bit space.
func newIdentifierAllocator(limit int) *identifierAllocator {
	if limit <= 0 || limit > maxIdentifier-minIdentifier+1 {
		limit = maxIdentifier - minIdentifier + 1
	}
	return &identifierAllocator{next: minIdentifier, limit: limit}
}

// Allocate returns an unused identifier.
//
// Returns:
//   - The identifier
//   - An error when the limit is reached, which a caller answers with a
//     too-many-resources status rather than by handing out a duplicate
func (a *identifierAllocator) Allocate() (uint16, error) {
	if a.inUse >= a.limit {
		return 0, fmt.Errorf("no identifier available: %d already in use", a.inUse)
	}

	// Prefer a released identifier so the counter does not run away.
	if n := len(a.freed); n > 0 {
		identifier := a.freed[n-1]
		a.freed = a.freed[:n-1]
		a.inUse++
		return identifier, nil
	}

	if a.next > maxIdentifier {
		return 0, fmt.Errorf("no identifier available: the 16-bit space is exhausted")
	}
	identifier := a.next
	a.next++
	a.inUse++
	return identifier, nil
}

// Release returns an identifier for reuse. Releasing one that was never
// allocated, or releasing twice, is ignored rather than corrupting the count.
func (a *identifierAllocator) Release(identifier uint16) {
	if identifier < minIdentifier || identifier > maxIdentifier || a.inUse == 0 {
		return
	}
	for _, freed := range a.freed {
		if freed == identifier {
			return
		}
	}
	a.freed = append(a.freed, identifier)
	a.inUse--
}

// InUse reports how many identifiers are currently allocated.
func (a *identifierAllocator) InUse() int {
	return a.inUse
}

// Session returns the session a UID names, or nil when it names none.
func (c *Connection) Session(uid uint16) *Session {
	return c.sessions[uid]
}

// Sessions returns the sessions established on the connection.
func (c *Connection) Sessions() []*Session {
	established := make([]*Session, 0, len(c.sessions))
	for _, session := range c.sessions {
		established = append(established, session)
	}
	return established
}

// addSession records an authenticated session against its UID.
func (c *Connection) addSession(session *Session) {
	c.sessions[session.UID] = session
}

// removeSession drops a session and releases its UID for reuse.
func (c *Connection) removeSession(uid uint16) *Session {
	session, ok := c.sessions[uid]
	if !ok {
		return nil
	}
	delete(c.sessions, uid)
	c.uids.Release(uid)
	return session
}
