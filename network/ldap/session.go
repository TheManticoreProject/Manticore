package ldap

import (
	"crypto/tls"
	"fmt"

	"github.com/TheManticoreProject/Manticore/windows/credentials"

	"github.com/TheManticoreProject/Manticore/network/kerberos"
	"github.com/go-ldap/ldap/v3"
	"github.com/go-ldap/ldap/v3/gssapi"
	krb5client "github.com/jcmturner/gokrb5/v8/client"
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
//	session := &Session{}
//	err := session.InitSession("ldap.example.com", 389, false, true, "EXAMPLE", "user", "password", false)
//	if err != nil {
//		log.Fatalf("Failed to initialize session: %s", err)
//	}
//	success := session.Connect()
//	if !success {
//		log.Fatalf("Failed to connect to LDAP server")
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
}

// InitSession initializes the LDAP session with the provided configuration and credentials.
//
// Parameters:
//
//	host (string): The hostname or IP address of the LDAP server.
//	port (int): The port number to connect to on the LDAP server. Must be in the range 1-65535.
//	useldaps (bool): A flag indicating whether to use LDAPS (LDAP over SSL).
//	usekerberos (bool): A flag indicating whether to use Kerberos for authentication.
//	domain (string): The domain name for the LDAP server.
//	username (string): The username for authentication.
//	password (string): The password for authentication.
//	debug (bool): A flag indicating whether to enable debug mode.
//
// Returns:
//
//	error: An error object if the initialization fails, otherwise nil.
//
// Example:
//
//	session := &Session{}
//	err := session.InitSession("ldap.example.com", 389, false, true, "EXAMPLE", "user", "password", false)
//	if err != nil {
//		log.Fatalf("Failed to initialize session: %s", err)
//	}
//
// Note:
//
//	The function validates the provided port number to ensure it is within the valid range (1-65535).
//	It then sets the network, credentials, and configuration fields of the Session struct accordingly.
func (s *Session) InitSession(host string, port int, credentials *credentials.Credentials, useldaps bool, usekerberos bool) error {
	// Check if TCP port is valid
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port number. Port must be in the range 1-65535")
	}

	// Network
	s.host = host
	s.port = port

	// Credentials
	s.credentials = credentials

	// Config
	s.useldaps = useldaps
	s.usekerberos = usekerberos

	return nil
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
//	session := &Session{}
//	err := session.InitSession("ldap.example.com", 389, false, true, "EXAMPLE", "user", "password", false)
//	if err != nil {
//		logger.Error(fmt.Sprintf("Failed to initialize session: %s", err))
//		return
//	}
//	success := session.Connect()
//	if success {
//		logger.Info("Successfully connected to the LDAP server")
//	} else {
//		logger.Warn("Failed to connect to the LDAP server")
//	}
//
// Note:
//
//	The function uses the configuration set in the Session struct to determine the connection parameters.
//	If useldaps is true, it will attempt to establish an LDAPS connection. If usekerberos is true, it will
//	use Kerberos for authentication.
func (s *Session) Connect() (bool, error) {
	// Set up LDAP connection
	var ldapConnection *ldap.Conn
	var err error

	// Connect to remote server
	if s.useldaps {
		// LDAPS connection
		ldapConnection, err = ldap.DialURL(
			fmt.Sprintf("ldaps://%s:%d", s.host, s.port),
			ldap.DialWithTLSConfig(
				&tls.Config{
					InsecureSkipVerify: true,
				},
			),
		)
		if err != nil {
			return false, fmt.Errorf("error connecting to LDAPS server: %s", err)
		}
	} else {
		// Regular LDAP connection
		ldapConnection, err = ldap.DialURL(
			fmt.Sprintf("ldap://%s:%d", s.host, s.port),
		)
		if err != nil {
			return false, fmt.Errorf("error connecting to LDAP server: %s", err)
		}
	}

	// Use Kerberos
	if s.usekerberos {
		servicePrincipalName, krb5Conf := kerberos.KerberosInit(s.host, s.credentials.GetDomain())

		// Initialize kerberos client
		// Inspired from: https://github.com/go-ldap/ldap/blob/06d50d1ad03bcd323e48f2fe174d95ceb31b8b90/v3/gssapi/client.go#L51
		kerberosClient := gssapi.Client{
			Client: krb5client.NewWithPassword(
				s.credentials.GetUsername(),
				s.credentials.GetDomain(),
				s.credentials.GetPassword(),
				krb5Conf,
				// Active Directory does not commonly support FAST negotiation so you will need to disable this on the client.
				// If this is the case you will see this error: KDC did not respond appropriately to FAST negotiation
				// https://github.com/jcmturner/gokrb5/blob/master/USAGE.md#active-directory-kdc-and-fast-negotiation
				krb5client.DisablePAFXFAST(true),
			),
		}
		defer kerberosClient.Close()

		err = ldapConnection.GSSAPIBindRequest(
			&kerberosClient,
			&ldap.GSSAPIBindRequest{
				ServicePrincipalName: servicePrincipalName,
				AuthZID:              "",
			},
		)
		if err != nil {
			return false, fmt.Errorf("error binding with Kerberos: %w", err)
		}
	} else {
		// Use NTLM authentification or null auth
		if s.credentials.CanPassTheHash() {
			// Bind with Pass the NT Hash
			if len(s.credentials.GetPassword()) > 0 {
				// Binding with credentials
				err = ldapConnection.Bind(fmt.Sprintf("%s@%s", s.credentials.GetUsername(), s.credentials.GetDomain()), s.credentials.GetPassword())
				if err != nil {
					return false, fmt.Errorf("error binding with Pass the NT Hash: %w", err)
				}
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
	s.connection.Close()
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
	s.connection.Close()
}
