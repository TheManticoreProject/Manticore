package data_types

// The LPSTR type and its alias PSTR specify a pointer to an array of 8-bit characters, which MAY be terminated by a null character.
// In some protocols, it is acceptable to not terminate with a null character, and this option will be indicated in the specification. In this case, the LPSTR or PSTR type MUST either be tagged with the IDL modifier [string], that indicates string semantics, or be accompanied by an explicit length specifier, for example [size_is()].
// The format of the characters MUST be specified by the protocol that uses them. Two common 8-bit formats are ANSI and UTF-8.
// A 32-bit pointer to a string of 8-bit characters, which MAY be null-terminated.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/3f6cc0e2-1303-4088-a26b-fb9582f29197
type LPSTR = *CHAR
