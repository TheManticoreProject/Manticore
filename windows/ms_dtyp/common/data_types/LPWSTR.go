package data_types

// The LPWSTR type is a 32-bit pointer to a string of 16-bit Unicode characters, which MAY be null-terminated.
// The LPWSTR type specifies a pointer to a sequence of Unicode characters, which MAY be terminated by a null character
// (usually referred to as "null-terminated Unicode").
// In some protocols, an acceptable option is to not terminate a sequence of Unicode characters with a null character.
// Where this option applies, it is indicated in the protocol specification. In this situation, the LPWSTR or PWSTR type
// MUST either be tagged with the IDL modifier [string], which indicates string semantics, or MUST be accompanied by an
// explicit length specifier, as specified in the RPC_UNICODE_STRING (section 2.3.10) structure.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/50e9ef83-d6fd-4e22-a34a-2c6b4e3c24f3
type LPWSTR = *WCHAR
