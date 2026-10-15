// Package auth derives IBM i host-server password substitutes and owns the
// redacted password credential.
//
// Host servers never receive the clear password. Client and server exchange
// eight-byte seeds, and the client sends a substitute derived from the
// profile, the password and both seeds. The server's password level selects
// the algorithm:
//
//   - levels 0/1: an 8-byte DES-based substitute (credential type 0x01);
//   - levels 2/3: a 20-byte SHA-1 substitute (credential type 0x03);
//   - level 4: a 64-byte SHA-512 substitute over a PBKDF2-HMAC-SHA512 token
//     (credential type 0x07).
//
// DES and SHA-1 are required for wire compatibility with V5R4-era systems,
// not endorsed as secure cryptography; prefer TLS. Level 4 is checked
// against known-answer vectors but has not been validated against a live
// level-4 system.
//
// Evidence: iSeriesNavigatorReverse docs/05-signon-and-authentication.md
// (algorithms, restricted sign-on character table) and the known-answer
// vectors in .research/family-bf/signon-crypto-vectors.txt, produced by the
// installed JTOpen 4.8 jar as a black-box oracle. Level 4 follows the
// algorithm outline in docs/05; its details (UTF-8 password for PBKDF2,
// UTF-16 code units for the salt's rightmost characters) were pinned by
// known-answer vectors produced by a JTOpen 21.0.7 jar used the same way.
// The code here is written from those descriptions, not translated from
// JTOpen.
package auth
