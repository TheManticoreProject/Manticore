// Package filesystem holds Windows file-system data structures used over SMB and
// by the local file-system APIs.
//
// Today it implements the [MS-FSCC] information classes exchanged via SMB2
// QUERY_INFO / SET_INFO (and locally by NtQueryInformationFile /
// NtSetInformationFile) — the FILE_*_INFORMATION and FILE_FS_*_INFORMATION
// structures. The class numbers that select them live in the infoclass
// subpackage. More file-system structures (reparse points, EAs, control codes,
// …) can be added here as they are needed.
package filesystem
