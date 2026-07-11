package kerberos

import (
	"encoding/asn1"
	"time"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// WithPostdate configures GetTGT to request a postdated Ticket Granting Ticket
// with the given (future) start time (RFC 4120 §3.3). The AS-REQ is sent with
// the allow-postdate and postdated KDC options and a from field carrying start;
// its endtime is derived from start (a 24h window). The KDC — if postdating is
// permitted by policy — returns a TGT flagged POSTDATED and INVALID whose start
// time lies in the future. Call Validate once that start time is reached (or
// where the KDC allows, immediately) to clear the INVALID flag and turn it into
// a usable TGT.
//
// Many production KDCs (default Active Directory policy included) disable
// postdating, in which case GetTGT surfaces the KDC's policy error
// (e.g. KDC_ERR_POLICY / KDC_ERR_CANNOT_POSTDATE).
//
// Returns the client to allow fluent chaining. Passing the zero time clears the
// request, restoring an ordinary immediate-start TGT.
func (c *KerberosClient) WithPostdate(start time.Time) *KerberosClient {
	c.postdateFrom = start.UTC()
	return c
}

// asReqKDCOptions returns the KDCOptions BitString for the AS-REQ. When a
// postdated TGT has been requested (WithPostdate) it adds the allow-postdate and
// postdated bits to the usual forwardable/proxiable/renewable set: postdated
// tells the KDC to honor the future from time (issuing the ticket INVALID until
// validated), and allow-postdate marks the ticket MAY-POSTDATE so it may itself
// authorize further postdated tickets (RFC 4120 §3.3). Otherwise it returns the
// standard AS-REQ options.
func (c *KerberosClient) asReqKDCOptions() asn1.BitString {
	if c.postdateFrom.IsZero() {
		return kdcOptionsForASReq()
	}
	return encodeKDCOptions(
		kdcOptionForwardable,
		kdcOptionProxiable,
		kdcOptionRenewable,
		kdcOptionAllowPostdate,
		kdcOptionPostdated,
	)
}

// TGTFlags returns the ticket-flags BitString of the currently held TGT (the
// decrypted AS-REP/TGS-REP enc-part flags). It is the zero BitString before
// GetTGT succeeds. See TGTMayPostdate / TGTPostdated / TGTInvalid for the
// postdating-relevant decodes.
func (c *KerberosClient) TGTFlags() asn1.BitString {
	return c.tgtEnc.Flags
}

// TGTMayPostdate reports whether the held TGT carries the MAY-POSTDATE flag,
// i.e. it is authorized to obtain postdated tickets (RFC 4120 §2.4).
func (c *KerberosClient) TGTMayPostdate() bool {
	return c.tgtEnc.Flags.At(iana.TicketFlagMayPostdate) == 1
}

// TGTPostdated reports whether the held TGT carries the POSTDATED flag, i.e. it
// was issued with a start time in the future (RFC 4120 §2.4).
func (c *KerberosClient) TGTPostdated() bool {
	return c.tgtEnc.Flags.At(iana.TicketFlagPostdated) == 1
}

// TGTInvalid reports whether the held TGT carries the INVALID flag. A postdated
// ticket is issued INVALID and must be turned valid by a VALIDATE exchange (see
// Validate) once its start time is reached (RFC 4120 §2.4, §3.3.3).
func (c *KerberosClient) TGTInvalid() bool {
	return c.tgtEnc.Flags.At(iana.TicketFlagInvalid) == 1
}
