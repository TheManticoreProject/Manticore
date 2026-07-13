package serviceprincipalname

import (
	"sort"
	"strings"
)

// HostServiceClass is the special "HOST" service class. A HOST SPN registered on
// a computer account implicitly substitutes for every built-in service class SPN
// (see BuiltInServiceClasses): unless a built-in SPN is explicitly registered on
// another object, a client requesting it is satisfied by the computer's HOST SPN.
// HOST itself is not one of the built-in service classes; it is the class that
// stands in for them.
const HostServiceClass = "HOST"

// builtInServiceClasses is the set of service class names that Active Directory
// recognizes as built-in SPNs for computer accounts. The names are stored
// lowercased; SPNs are not case sensitive, so all lookups fold case.
//
// Source (verbatim list): https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/setspn
// ("The built-in SPNs that are recognized for computer accounts are:").
var builtInServiceClasses = map[string]struct{}{
	"alerter":          {},
	"appmgmt":          {},
	"browser":          {},
	"cifs":             {},
	"cisvc":            {},
	"clipsrv":          {},
	"dcom":             {},
	"dhcp":             {},
	"dmserver":         {},
	"dns":              {},
	"dnscache":         {},
	"eventlog":         {},
	"eventsystem":      {},
	"fax":              {},
	"http":             {},
	"ias":              {},
	"iisadmin":         {},
	"mcsvc":            {},
	"messenger":        {},
	"msiserver":        {},
	"netdde":           {},
	"netddedsm":        {},
	"netlogon":         {},
	"netman":           {},
	"nmagent":          {},
	"oakley":           {},
	"plugplay":         {},
	"policyagent":      {},
	"protectedstorage": {},
	"rasman":           {},
	"remoteaccess":     {},
	"replicator":       {},
	"rpc":              {},
	"rpclocator":       {},
	"rpcss":            {},
	"rsvp":             {},
	"samss":            {},
	"scardsvr":         {},
	"scesrv":           {},
	"schedule":         {},
	"scm":              {},
	"seclogon":         {},
	"snmp":             {},
	"spooler":          {},
	"tapisrv":          {},
	"time":             {},
	"trksvr":           {},
	"trkwks":           {},
	"ups":              {},
	"w3svc":            {},
	"wins":             {},
	"www":              {},
}

// BuiltInServiceClasses returns the built-in SPN service class names recognized
// for computer accounts, lowercased and sorted. The returned slice is a fresh
// copy the caller may modify. HOST is not included (see HostServiceClass).
func BuiltInServiceClasses() []string {
	out := make([]string, 0, len(builtInServiceClasses))
	for class := range builtInServiceClasses {
		out = append(out, class)
	}
	sort.Strings(out)
	return out
}

// IsBuiltInServiceClass reports whether class is one of the built-in SPN service
// classes recognized for computer accounts. The comparison is case-insensitive.
func IsBuiltInServiceClass(class string) bool {
	_, ok := builtInServiceClasses[strings.ToLower(class)]
	return ok
}

// IsBuiltInServiceClass reports whether the SPN's service class is one of the
// built-in service classes recognized for computer accounts (case-insensitive).
func (s *ServicePrincipalName) IsBuiltInServiceClass() bool {
	if s == nil {
		return false
	}
	return IsBuiltInServiceClass(s.ServiceClass)
}

// IsHostServiceClass reports whether the SPN's service class is the special
// "HOST" class (case-insensitive).
func (s *ServicePrincipalName) IsHostServiceClass() bool {
	if s == nil {
		return false
	}
	return strings.EqualFold(s.ServiceClass, HostServiceClass)
}

// CoveredByHostSPN reports whether a HOST SPN on the same computer account would
// implicitly satisfy a request for this SPN. This is true exactly when the
// service class is a built-in one: per the setspn documentation, a computer's
// HOST SPN substitutes for any built-in service class SPN that is not explicitly
// registered on another object. It does not account for such explicit
// registrations, which are not knowable from the SPN string alone.
func (s *ServicePrincipalName) CoveredByHostSPN() bool {
	return s.IsBuiltInServiceClass()
}
