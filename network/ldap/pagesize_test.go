package ldap

import (
	"net"
	"testing"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/go-ldap/ldap/v3"
)

// pagingRecorder is a minimal LDAP server that answers every SearchRequest with an
// empty SearchResultDone and records the page size carried by the paged-results
// control (1.2.840.113556.1.4.319) the client attached to it. It exists to observe
// what the session actually puts on the wire, which is the only place the page size
// is visible from outside the package.
type pagingRecorder struct {
	listener  net.Listener
	pageSizes []uint32
	unpaged   []bool
	done      chan struct{}
}

func newPagingRecorder(t *testing.T) *pagingRecorder {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %s", err)
	}

	recorder := &pagingRecorder{listener: listener, done: make(chan struct{})}
	go recorder.serve()

	return recorder
}

func (r *pagingRecorder) serve() {
	defer close(r.done)

	conn, err := r.listener.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		packet, err := ber.ReadPacket(conn)
		if err != nil {
			return
		}
		if len(packet.Children) < 2 {
			return
		}

		messageID, ok := packet.Children[0].Value.(int64)
		if !ok {
			return
		}

		switch packet.Children[1].Tag {
		case ldap.ApplicationSearchRequest:
			pageSize, paged := uint32(0), false
			if len(packet.Children) > 2 {
				for _, controlPacket := range packet.Children[2].Children {
					control, err := ldap.DecodeControl(controlPacket)
					if err != nil {
						continue
					}
					if pagingControl, ok := control.(*ldap.ControlPaging); ok {
						pageSize, paged = pagingControl.PagingSize, true
					}
				}
			}
			r.pageSizes = append(r.pageSizes, pageSize)
			r.unpaged = append(r.unpaged, !paged)
			conn.Write(ldapResultDone(messageID, ldap.ApplicationSearchResultDone))
		case ldap.ApplicationUnbindRequest:
			return
		default:
			conn.Write(ldapResultDone(messageID, ldap.ApplicationBindResponse))
		}
	}
}

// ldapResultDone builds an LDAPMessage carrying a success result for the given
// application response tag.
func ldapResultDone(messageID int64, applicationTag ber.Tag) []byte {
	message := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Response")
	message.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, messageID, "MessageID"))

	response := ber.Encode(ber.ClassApplication, ber.TypeConstructed, applicationTag, nil, "Response")
	response.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, 0, "resultCode"))
	response.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", "matchedDN"))
	response.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", "diagnosticMessage"))
	message.AppendChild(response)

	return message.Bytes()
}

// recordPageSizes runs search against a session connected to a recording server and
// returns the page sizes the searches asked for, in order.
func recordPageSizes(t *testing.T, search func(session *Session)) ([]uint32, []bool) {
	t.Helper()

	recorder := newPagingRecorder(t)
	defer recorder.listener.Close()

	conn, err := net.Dial("tcp", recorder.listener.Addr().String())
	if err != nil {
		t.Fatalf("failed to dial the recording server: %s", err)
	}

	connection := ldap.NewConn(conn, false)
	connection.Start()

	session := &Session{connection: connection}
	search(session)

	connection.Close()
	<-recorder.done

	return recorder.pageSizes, recorder.unpaged
}

// TestQueryPageSizeDefault checks that a session which was never told otherwise
// pages at DefaultPageSize.
func TestQueryPageSizeDefault(t *testing.T) {
	pageSizes, unpaged := recordPageSizes(t, func(session *Session) {
		if _, err := session.Query("DC=example,DC=com", "(objectClass=*)", []string{"cn"}, ldap.ScopeWholeSubtree); err != nil {
			t.Errorf("Query returned an error: %s", err)
		}
	})

	if len(pageSizes) != 1 {
		t.Fatalf("expected 1 search, got %d", len(pageSizes))
	}
	if unpaged[0] {
		t.Fatalf("expected the search to carry a paged-results control")
	}
	if pageSizes[0] != DefaultPageSize {
		t.Fatalf("expected a page size of %d, got %d", DefaultPageSize, pageSizes[0])
	}
}

// TestQueryPageSizeSet checks that SetPageSize is what reaches the wire, for Query
// and for the scope wrappers built on it.
func TestQueryPageSizeSet(t *testing.T) {
	pageSizes, unpaged := recordPageSizes(t, func(session *Session) {
		session.SetPageSize(42)

		if _, err := session.Query("DC=example,DC=com", "(objectClass=*)", []string{"cn"}, ldap.ScopeWholeSubtree); err != nil {
			t.Errorf("Query returned an error: %s", err)
		}
		if _, err := session.QueryWholeSubtree("DC=example,DC=com", "(objectClass=*)", []string{"cn"}); err != nil {
			t.Errorf("QueryWholeSubtree returned an error: %s", err)
		}
	})

	if len(pageSizes) != 2 {
		t.Fatalf("expected 2 searches, got %d", len(pageSizes))
	}
	for i, pageSize := range pageSizes {
		if unpaged[i] {
			t.Fatalf("expected search %d to carry a paged-results control", i)
		}
		if pageSize != 42 {
			t.Fatalf("expected search %d to ask for a page size of 42, got %d", i, pageSize)
		}
	}
}

// TestQueryPageSizeZeroIsDefault checks that zero means unset rather than a page
// size of zero, so a session literal and an explicit SetPageSize(0) both page at
// DefaultPageSize.
func TestQueryPageSizeZeroIsDefault(t *testing.T) {
	pageSizes, _ := recordPageSizes(t, func(session *Session) {
		session.SetPageSize(0)

		if _, err := session.Query("DC=example,DC=com", "(objectClass=*)", []string{"cn"}, ldap.ScopeWholeSubtree); err != nil {
			t.Errorf("Query returned an error: %s", err)
		}
	})

	if len(pageSizes) != 1 || pageSizes[0] != DefaultPageSize {
		t.Fatalf("expected a page size of %d, got %v", DefaultPageSize, pageSizes)
	}
}

// TestSecurityDescriptorPageSizeSet checks that GetNtSecurityDescriptorOf, the
// second call site that used to hardcode the page size, honours the session setting
// too.
func TestSecurityDescriptorPageSizeSet(t *testing.T) {
	pageSizes, _ := recordPageSizes(t, func(session *Session) {
		session.SetPageSize(7)

		// The recording server returns no entries, so this reports "no entry
		// returned"; the page size it asked for is what is under test.
		session.GetNtSecurityDescriptorOf("DC=example,DC=com")
	})

	if len(pageSizes) != 1 || pageSizes[0] != 7 {
		t.Fatalf("expected a page size of 7, got %v", pageSizes)
	}
}

// TestGetPageSize checks the accessor against both a session built by NewSession
// and a zero-value session literal.
func TestGetPageSize(t *testing.T) {
	session, err := NewSession("ldap.example.com", 389, nil, false, false)
	if err != nil {
		t.Fatalf("failed to create the session: %s", err)
	}
	if session.GetPageSize() != DefaultPageSize {
		t.Fatalf("expected NewSession to default to %d, got %d", DefaultPageSize, session.GetPageSize())
	}

	session.SetPageSize(500)
	if session.GetPageSize() != 500 {
		t.Fatalf("expected 500, got %d", session.GetPageSize())
	}

	if (&Session{}).GetPageSize() != DefaultPageSize {
		t.Fatalf("expected a zero-value session to report %d, got %d", DefaultPageSize, (&Session{}).GetPageSize())
	}
}
