// Package serviceprincipalname implements a parser, validator and renderer for
// Kerberos Service Principal Names (SPNs).
//
// An SPN is the name a client uses to identify a service for mutual
// authentication. Per [MS-ADTS] it is a string with the format:
//
//	serviceclass "/" hostname [":"port | ":"instancename] ["/" servicename]
//
// An SPN consists of either two parts or three parts, each separated by a
// forward slash ("/"). The first part is the service class, the second part is
// the host name, and the third part (if present) is the service name. The host
// name part can optionally be suffixed with either a ":port" component or a
// ":instancename" component. A port component is distinguished from an
// instancename component by being entirely composed of numeric digits.
//
// For example, "ldap/dc-01.fabrikam.com/fabrikam.com" is a three-part SPN where
// "ldap" is the service class name, "dc-01.fabrikam.com" is the host name, and
// "fabrikam.com" is the service name.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/cd328386-4d97-4666-be33-056545c1cad2
package serviceprincipalname

import (
	"fmt"
	"strconv"
	"strings"
)

// ServicePrincipalName represents a parsed Kerberos SPN of the form:
//
//	serviceclass "/" hostname [":"port | ":"instancename] ["/" servicename]
//
// Exactly one of Port (non-zero) or InstanceName (non-empty) may be set: the
// ":"-suffix on the host name is interpreted as a port when it is composed
// entirely of numeric digits, and as an instance name otherwise.
type ServicePrincipalName struct {
	// ServiceClass is the first part of the SPN (e.g. "ldap", "cifs",
	// "MSSQLSvc", "HTTP"). It is required.
	ServiceClass string

	// Hostname is the host portion of the second part of the SPN, with any
	// ":port" or ":instancename" suffix removed. It is required.
	Hostname string

	// Port is the value of the optional numeric ":port" suffix on the host
	// name. It is 0 when no numeric port suffix is present.
	Port uint16

	// InstanceName is the value of the optional non-numeric ":instancename"
	// suffix on the host name. It is empty when no such suffix is present.
	InstanceName string

	// ServiceName is the optional third part of the SPN. It is empty for a
	// two-part SPN.
	ServiceName string
}

// FromString parses spn into a ServicePrincipalName and validates it.
//
// It returns an error when spn is empty, has fewer than two or more than three
// "/"-separated parts, contains an empty component, or carries a numeric port
// suffix that does not fit in a uint16.
func FromString(spn string) (*ServicePrincipalName, error) {
	if strings.TrimSpace(spn) == "" {
		return nil, fmt.Errorf("invalid SPN: empty string")
	}

	parts := strings.Split(spn, "/")
	if len(parts) < 2 || len(parts) > 3 {
		return nil, fmt.Errorf("invalid SPN %q: expected 2 or 3 \"/\"-separated parts, got %d", spn, len(parts))
	}

	s := &ServicePrincipalName{
		ServiceClass: parts[0],
	}

	if len(parts) == 3 {
		if parts[2] == "" {
			return nil, fmt.Errorf("invalid SPN %q: empty service name after \"/\"", spn)
		}
		s.ServiceName = parts[2]
	}

	// Parse the host portion (parts[1]), splitting on the first ":" to detect
	// an optional port or instance-name suffix.
	host := parts[1]
	if colon := strings.Index(host, ":"); colon >= 0 {
		s.Hostname = host[:colon]
		suffix := host[colon+1:]

		if suffix == "" {
			return nil, fmt.Errorf("invalid SPN %q: empty port/instance-name suffix after \":\"", spn)
		}

		// A suffix composed entirely of digits is a port; anything else is an
		// instance name.
		if isAllDigits(suffix) {
			port, err := strconv.ParseUint(suffix, 10, 16)
			if err != nil {
				return nil, fmt.Errorf("invalid SPN %q: port %q out of range: %w", spn, suffix, err)
			}
			s.Port = uint16(port)
		} else {
			s.InstanceName = suffix
		}
	} else {
		s.Hostname = host
	}

	if err := s.Validate(); err != nil {
		return nil, err
	}

	return s, nil
}

// isAllDigits reports whether s is non-empty and composed entirely of ASCII
// decimal digits. This mirrors the [MS-ADTS] rule for distinguishing a port
// from an instance name.
func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// Validate checks that the ServicePrincipalName is well-formed: the service
// class and host name are present, at most one of Port/InstanceName is set, and
// no component contains a separator ("/" or ":") that would break round-tripping
// through String.
//
// FromString calls Validate before returning, so any value produced by
// FromString is valid. Validate is intended for values constructed by hand.
func (s *ServicePrincipalName) Validate() error {
	if s == nil {
		return fmt.Errorf("invalid SPN: nil")
	}
	if s.ServiceClass == "" {
		return fmt.Errorf("invalid SPN: empty service class")
	}
	if s.Hostname == "" {
		return fmt.Errorf("invalid SPN: empty host name")
	}
	if s.Port != 0 && s.InstanceName != "" {
		return fmt.Errorf("invalid SPN: port and instance name are mutually exclusive")
	}

	// Components must not embed the SPN separators, otherwise String would
	// produce a string that no longer parses back to the same components.
	for label, value := range map[string]string{
		"service class": s.ServiceClass,
		"service name":  s.ServiceName,
		"instance name": s.InstanceName,
	} {
		if strings.ContainsAny(value, "/:") {
			return fmt.Errorf("invalid SPN: %s %q contains a separator", label, value)
		}
	}
	if strings.Contains(s.Hostname, "/") {
		return fmt.Errorf("invalid SPN: host name %q contains a separator", s.Hostname)
	}

	return nil
}

// IsValid reports whether the ServicePrincipalName is well-formed. It is a
// boolean convenience wrapper over Validate.
func (s *ServicePrincipalName) IsValid() bool {
	return s.Validate() == nil
}

// HasServiceName reports whether the SPN is a three-part SPN (i.e. it carries a
// service name).
func (s *ServicePrincipalName) HasServiceName() bool {
	return s.ServiceName != ""
}

// String renders the canonical SPN string. For a value produced by FromString
// the result round-trips back to an equal ServicePrincipalName.
func (s *ServicePrincipalName) String() string {
	var b strings.Builder
	b.WriteString(s.ServiceClass)
	b.WriteByte('/')
	b.WriteString(s.Hostname)

	if s.Port != 0 {
		b.WriteByte(':')
		b.WriteString(strconv.FormatUint(uint64(s.Port), 10))
	} else if s.InstanceName != "" {
		b.WriteByte(':')
		b.WriteString(s.InstanceName)
	}

	if s.ServiceName != "" {
		b.WriteByte('/')
		b.WriteString(s.ServiceName)
	}

	return b.String()
}

// Equal reports whether s and other describe the same SPN. Comparison is
// case-insensitive, matching the case-insensitive handling of SPNs in Kerberos
// and Active Directory.
func (s *ServicePrincipalName) Equal(other *ServicePrincipalName) bool {
	if s == nil || other == nil {
		return s == other
	}
	return strings.EqualFold(s.ServiceClass, other.ServiceClass) &&
		strings.EqualFold(s.Hostname, other.Hostname) &&
		s.Port == other.Port &&
		strings.EqualFold(s.InstanceName, other.InstanceName) &&
		strings.EqualFold(s.ServiceName, other.ServiceName)
}
