package ldap

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Monday, January 1, 1601 12:00:00 AM in 100-nanosecond intervals
const UnixTimestampStart int64 = 116444736000000000

// GetDomainFromDistinguishedName extracts the domain name from a distinguished name (DN).
// A distinguished name is a string that uniquely identifies an entry in the LDAP directory.
// The function splits the DN into its components and concatenates the domain components (DC) to form the domain name.
//
// Parameters:
//
//	distinguishedName (string): The distinguished name from which to extract the domain name.
//
// Returns:
//
//	string: The extracted domain name.
//
// Example:
//
//	distinguishedName := "CN=John Doe,OU=Users,DC=example,DC=com"
//	domain := GetDomainFromDistinguishedName(distinguishedName)
//	// domain will be "example.com"
//
// Note:
//
//	The function assumes that the distinguished name is well-formed and contains domain components (DC).
func GetDomainFromDistinguishedName(distinguishedName string) string {
	domainParts := strings.Split(distinguishedName, ",")

	domain := ""
	for _, part := range domainParts {
		if strings.HasPrefix(part, "DC=") {
			domain += strings.TrimPrefix(part, "DC=") + "."
		}
	}

	domain = strings.TrimSuffix(domain, ".")

	return domain
}

// ConvertLDAPTimeStampToUnixTimeStamp converts an LDAP timestamp to a Unix timestamp.
// LDAP timestamps are represented as the number of 100-nanosecond intervals since January 1, 1601 (UTC).
// Unix timestamps are represented as the number of seconds since January 1, 1970 (UTC).
//
// Parameters:
//
//	value (string): The LDAP timestamp to be converted.
//
// Returns:
//
//	int64: The converted Unix timestamp.
//	error: An error object if the conversion fails, otherwise nil.
//
// Example:
//
//	ldapTimestamp := "132537600000000000"
//	unixTimestamp, err := ConvertLDAPTimeStampToUnixTimeStamp(ldapTimestamp)
//	if err != nil {
//		logger.Warn(fmt.Sprintf("Error converting LDAP timestamp to Unix timestamp: %s", err))
//		return
//	}
//	// unixTimestamp will be 1640995200 (corresponding to January 1, 2022)
//
// Note:
//
//	If the LDAP timestamp is invalid or cannot be converted, the function returns 0 and an error.
func ConvertLDAPTimeStampToUnixTimeStamp(value string) (int64, error) {
	convertedValue := int64(0)

	if len(value) != 0 {
		valueInt, err := strconv.ParseInt(value, 10, 64)

		if err != nil {
			return convertedValue, fmt.Errorf("error converting value to int64: %w", err)
		}

		if valueInt < UnixTimestampStart {
			// Typically for dates on year 1601
			convertedValue = 0
		} else {
			// Reduce the 100-nanosecond interval count to seconds directly.
			// Scaling to nanoseconds first (* 100) overflows int64 for large
			// values such as the 0x7FFFFFFFFFFFFFFF "never" sentinel.
			convertedValue = (valueInt - UnixTimestampStart) / int64(1e7)
		}
	}

	return convertedValue, nil
}

// ConvertLDAPDurationToSeconds converts an LDAP duration to seconds.
// LDAP durations are represented as the number of 100-nanosecond intervals.
//
// Parameters:
//
//	value (string): The LDAP duration to be converted.
//
// Returns:
//
//	int64: The converted duration in seconds.
//	error: An error object if the conversion fails, otherwise nil.
//
// Example:
//
//	ldapDuration := "864000000000"
//	seconds, err := ConvertLDAPDurationToSeconds(ldapDuration)
//	if err != nil {
//		logger.Warn(fmt.Sprintf("Error converting LDAP duration to seconds: %s", err))
//		return
//	}
//	// seconds will be 86400 (corresponding to 1 day)
//
// Note:
//
//	If the LDAP duration is invalid or cannot be converted, the function returns 0 and an error.
func ConvertLDAPDurationToSeconds(value string) (int64, error) {
	convertedValue := int64(0)

	if len(value) != 0 {
		valueInt, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return convertedValue, fmt.Errorf("error converting LDAP duration to seconds: %w", err)
		}

		if valueInt < 0 {
			valueInt = valueInt * int64(-1)
		}

		// Convert intervals of 100-nanoseconds to Seconds
		convertedValue = valueInt / int64(1e7)
	}

	return convertedValue, nil
}

// ConvertSecondsToLDAPDuration converts a duration in seconds to an LDAP duration.
// LDAP durations are represented as the number of 100-nanosecond intervals.
//
// Parameters:
//
//	value (int64): The duration in seconds to be converted.
//
// Returns:
//
//	string: The converted duration as an LDAP duration string.
//
// Example:
//
//	seconds := int64(86400) // 1 day
//	ldapDuration := ConvertSecondsToLDAPDuration(seconds)
//	if err != nil {
//		logger.Warn(fmt.Sprintf("Error converting seconds to LDAP duration: %s", err))
//		return
//	}
//	// ldapDuration will be "864000000000" (corresponding to 1 day)
//
// Note:
//
//	If the input value is negative, the function will still convert it to an LDAP duration.
func ConvertSecondsToLDAPDuration(value int64) string {
	convertedValue := fmt.Sprintf("%d", value*int64(1e7))
	return convertedValue
}

// ConvertUnixTimeStampToLDAPTimeStamp converts a Unix timestamp to an LDAP timestamp.
// LDAP timestamps are represented as the number of 100-nanosecond intervals since January 1, 1601 (UTC).
//
// Parameters:
//
//	value (time.Time): The Unix timestamp to be converted.
//
// Returns:
//
//	int64: The converted timestamp as an LDAP timestamp.
//
// Example:
//
//	unixTime := time.Now()
//	ldapTime := ConvertUnixTimeStampToLDAPTimeStamp(unixTime)
//	// ldapTime will be the corresponding LDAP timestamp for the current time.
//
// Note:
//
//	The function multiplies the Unix timestamp by 10,000,000 to convert seconds to 100-nanosecond intervals
//	and then adds the number of 100-nanosecond intervals between January 1, 1601, and January 1, 1970.
func ConvertUnixTimeStampToLDAPTimeStamp(value time.Time) int64 {
	ldapvalue := value.Unix() * (1e7)
	ldapvalue = ldapvalue + UnixTimestampStart
	return int64(ldapvalue)
}
