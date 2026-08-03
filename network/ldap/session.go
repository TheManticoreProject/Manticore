package ldap

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/gssapi"
	"github.com/TheManticoreProject/Manticore/windows/credentials"

	"github.com/go-ldap/ldap/v3"
)

// Session represents an LDAP session with configuration and connection details.
//
// Fields:
//
//	host (string): The hostname or IP address of the LDAP server.
//	port (int): The port number to connect to on the LDAP server.
//	connection (*ldap.Conn): The LDAP connection object.
//	domain (string): The domain name for the LDAP server.
//	username (string): The username for authentication.
//	password (string): The password for authentication.
//	debug (bool): A flag indicating whether to enable debug mode.
//	useldaps (bool): A flag indicating whether to use LDAPS (LDAP over SSL).
//	usekerberos (bool): A flag indicating whether to use Kerberos for authentication.
//
// Example:
//
//	session, err := NewSession("ldap.example.com", 389, credentials, false, false)
//	if err != nil {
//		log.Fatalf("Failed to create session: %s", err)
//	}
//	success, err := session.Connect()
//	if !success {
//		log.Fatalf("Failed to connect to LDAP server: %s", err)
//	}
type Session struct {
	// Network
	host       string
	port       int
	connection *ldap.Conn
	// Credentials
	credentials *credentials.Credentials
	// Config
	useldaps    bool
	usekerberos bool
	// tlsSkipVerify controls whether the server certificate is verified for
	// LDAPS connections. It defaults to true (verification disabled) to preserve
	// the historical behavior; use SetTLSSkipVerify to enable verification.
	tlsSkipVerify bool
	// gssapiLayer is the SASL GSSAPI security layer to negotiate for Kerberos
	// binds (RFC 4752 §3.1). It defaults to no security layer (auth only); use
	// SetGSSAPISigning / SetGSSAPISealing to request integrity or confidentiality.
	gssapiLayer byte
	// spnHostname overrides the hostname used to build the ldap/<host> service
	// principal name for the Kerberos bind. It is empty by default, in which case
	// the SPN is built from host. Set it (via SetKerberosSPNHostname) when host is
	// an IP address: Active Directory registers the SPN under the DC's FQDN, so a
	// Kerberos bind by IP needs the FQDN here while the TCP connection and the KDC
	// exchanges still use host.
	spnHostname string
	// gssClient retains the GSSAPI client for the lifetime of the session so the
	// negotiated per-message security context survives past the bind and is torn
	// down on Close.
	gssClient *nativeGSSAPIClient
}

// NewSession creates a new LDAP session with the provided configuration and credentials.
//
// Parameters:
//
//	host (string): The hostname or IP address of the LDAP server.
//	port (int): The port number to connect to on the LDAP server. Must be in the range 1-65535.
//	credentials (*credentials.Credentials): The credentials to use for authentication.
//	useldaps (bool): A flag indicating whether to use LDAPS (LDAP over SSL).
//	usekerberos (bool): A flag indicating whether to use Kerberos for authentication.
//
// Returns:
//
//	*Session: A new LDAP session object.
//	error: An error object if the creation fails, otherwise nil.
//
// Example:
//
//	session, err := NewSession("ldap.example.com", 389, credentials, false, false)
//	if err != nil {
//		logger.Warn(fmt.Sprintf("Error creating LDAP session: %s", err))
//		return
//	}
func NewSession(host string, port int, credentials *credentials.Credentials, useldaps bool, usekerberos bool) (*Session, error) {
	s := &Session{}

	// Check if TCP port is valid
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid port number. Port must be in the range 1-65535")
	}

	// Network
	s.host = host
	s.port = port

	// Credentials
	s.credentials = credentials

	// Config
	s.useldaps = useldaps
	s.usekerberos = usekerberos
	// Preserve the historical default of not verifying the server certificate.
	s.tlsSkipVerify = true
	// Preserve the historical default of an auth-only GSSAPI bind (no ongoing
	// sign/seal); callers opt into a security layer explicitly.
	s.gssapiLayer = saslLayerNone

	return s, nil
}

// SetGSSAPISigning requests the RFC 4752 integrity (signing) security layer for
// Kerberos binds: after the bind, every LDAP PDU is GSS-signed and its integrity
// verified on receipt. It must be called before Connect.
func (s *Session) SetGSSAPISigning() {
	s.gssapiLayer = saslLayerIntegrity
}

// SetGSSAPISealing requests the RFC 4752 confidentiality (sealing) security layer
// for Kerberos binds: after the bind, every LDAP PDU is GSS-encrypted (which also
// implies integrity). It must be called before Connect.
func (s *Session) SetGSSAPISealing() {
	s.gssapiLayer = saslLayerConfidentiality
}

// SetTLSSkipVerify controls whether the server certificate is verified when
// establishing an LDAPS connection.
//
// Parameters:
//
//	skip (bool): When true (the default), the server certificate is not verified.
//	             When false, the certificate chain and hostname are validated and
//	             the connection fails if validation does not succeed.
//
// This must be called before Connect to take effect.
func (s *Session) SetTLSSkipVerify(skip bool) {
	s.tlsSkipVerify = skip
}

// SetKerberosSPNHostname sets the hostname used to build the ldap/<host> service
// principal name for a Kerberos bind, overriding the connection host.
//
// Use it when the session connects to a domain controller by IP address: Active
// Directory registers the ldap SPN under the DC's FQDN, so a Kerberos bind by IP
// fails with KDC_ERR_S_PRINCIPAL_UNKNOWN unless the SPN is built from the FQDN. The
// TCP connection and the KDC exchanges continue to use the connection host, so the
// IP is still what is dialled.
//
// Parameters:
//
//	hostname (string): The FQDN to use in the SPN, or empty to use the connection host.
func (s *Session) SetKerberosSPNHostname(hostname string) {
	s.spnHostname = hostname
}

// kerberosSPNHostname returns the hostname the ldap SPN is built from: the override
// when one was set, otherwise the connection host.
func (s *Session) kerberosSPNHostname() string {
	if s.spnHostname != "" {
		return s.spnHostname
	}
	return s.host
}

// Connect establishes a connection to the LDAP server. It supports both regular LDAP and LDAPS connections,
// and can optionally use Kerberos for authentication.
//
// Returns:
//
//	bool: True if the connection is successful, false otherwise.
//
// Example:
//
//	session, err := NewSession("ldap.example.com", 389, credentials, false, false)
//	if err != nil {
//		logger.Warn(fmt.Sprintf("Error creating LDAP session: %s", err))
//		return
//	}
//	success := session.Connect()
//	if success {
//		logger.Info(fmt.Sprintf("Successfully connected to LDAP server: %s", s.host))
//	} else {
//		logger.Warn(fmt.Sprintf("Error connecting to LDAP server: %s", err))
//	}
func (s *Session) Connect() (bool, error) {
	var err error

	// Dial the transport ourselves and interpose a SASL wrapper connection so a
	// negotiated GSSAPI security layer (sign/seal) can be applied to every LDAP
	// PDU after the bind. The server certificate is captured for LDAPS so a
	// tls-server-end-point channel-binding token can be sent during the bind.
	var rawConn net.Conn
	var serverCert *x509.Certificate
	if s.useldaps {
		tlsConn, derr := tls.DialWithDialer(
			&net.Dialer{Timeout: ldap.DefaultTimeout},
			"tcp",
			fmt.Sprintf("%s:%d", s.host, s.port),
			&tls.Config{InsecureSkipVerify: s.tlsSkipVerify},
		)
		if derr != nil {
			return false, fmt.Errorf("error connecting to LDAPS server: %s", derr)
		}
		if certs := tlsConn.ConnectionState().PeerCertificates; len(certs) > 0 {
			serverCert = certs[0]
		}
		rawConn = tlsConn
	} else {
		rawConn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", s.host, s.port), ldap.DefaultTimeout)
		if err != nil {
			return false, fmt.Errorf("error connecting to LDAP server: %s", err)
		}
	}

	sc := newSASLConn(rawConn)
	ldapConnection := ldap.NewConn(sc, s.useldaps)
	ldapConnection.Start()

	// Use Kerberos
	if s.usekerberos {
		// Native (stdlib-only) GSSAPI SASL bind; the DC is both the LDAP server
		// and the KDC. Realm is the credentials' domain (upper-cased by the client).
		// The SPN is built from the DC's FQDN when host is an IP (see spnHostname).
		servicePrincipalName := fmt.Sprintf("ldap/%s", s.kerberosSPNHostname())
		gssClient, gerr := newNativeGSSAPIClient(s.host, s.credentials.GetDomain(), s.credentials)
		if gerr != nil {
			ldapConnection.Close()
			return false, fmt.Errorf("error initializing Kerberos: %w", gerr)
		}
		gssClient.desiredLayer = s.gssapiLayer
		// Over LDAPS, bind the GSSAPI context to the TLS channel (RFC 5929
		// tls-server-end-point), which AD requires when channel binding is enforced.
		if serverCert != nil {
			gssClient.channelBindings = gssapi.GSSChannelBindings(tlsServerEndPointCBT(serverCert))
		}
		s.gssClient = gssClient

		err = ldapConnection.GSSAPIBindRequest(
			gssClient,
			&ldap.GSSAPIBindRequest{
				ServicePrincipalName: servicePrincipalName,
				AuthZID:              "",
			},
		)
		if err != nil {
			ldapConnection.Close()
			return false, fmt.Errorf("error binding with Kerberos: %w", err)
		}

		// Install the negotiated security layer on the connection so subsequent
		// PDUs are signed and/or sealed. A no-security-layer bind leaves the
		// connection as a transparent pass-through.
		if cipher := gssClient.securityLayer(); cipher != nil {
			sc.activate(cipher)
		}
	} else {
		// Use NTLM authentification or null auth
		if s.credentials.CanPassTheHash() {
			// Bind with Pass the NT Hash using an NTLMSSP bind keyed on the NT hash,
			// not a cleartext simple bind keyed on the password.
			err = ldapConnection.NTLMBindWithHash(s.credentials.GetDomain(), s.credentials.GetUsername(), s.credentials.GetNTHash())
			if err != nil {
				return false, fmt.Errorf("error binding with Pass the NT Hash: %w", err)
			}
		} else if len(s.credentials.GetPassword()) > 0 {
			// Binding with credentials
			err = ldapConnection.Bind(fmt.Sprintf("%s@%s", s.credentials.GetUsername(), s.credentials.GetDomain()), s.credentials.GetPassword())
			if err != nil {
				return false, fmt.Errorf("error binding with credentials: %w", err)
			}
		} else {
			// Unauthenticated Bind
			bindDN := ""
			if s.credentials.GetUsername() != "" {
				bindDN = fmt.Sprintf("%s@%s", s.credentials.GetUsername(), s.credentials.GetDomain())
			}

			err = ldapConnection.UnauthenticatedBind(bindDN)
			if err != nil {
				return false, fmt.Errorf("error binding with unauthenticated bind: %w", err)
			}
		}
	}

	s.connection = ldapConnection

	return true, nil
}

// ReConnect attempts to re-establish the LDAP connection by closing the current connection and calling the Connect method again.
//
// Returns:
//
//	bool: True if the reconnection is successful, false otherwise.
//
// Example:
//
//	session := &Session{}
//	success := session.ReConnect()
//	if success {
//		fmt.Println("Reconnected successfully")
//	} else {
//		fmt.Println("Failed to reconnect")
//	}
//
// Note:
//
//	This function assumes that the Session struct has a valid connection object and that the Connect method is implemented correctly.
func (s *Session) ReConnect() (bool, error) {
	if s.connection != nil {
		s.connection.Close()
	}
	return s.Connect()
}

// Close terminates the LDAP session by closing the current connection.
//
// Example:
//
//	session := &Session{}
//	session.Close()
//
// Note:
//
//	This function assumes that the Session struct has a valid connection object.
func (s *Session) Close() {
	if s.connection != nil {
		s.connection.Close()
	}
	if s.gssClient != nil {
		s.gssClient.DeleteSecContext()
		s.gssClient = nil
	}
}
