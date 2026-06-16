![](./.github/banner.png)

<p align="center">
      Manticore is a framework for working with Windows network protocols (SMB, LDAP, DCE/RPC, and more), cryptography, and authentication, designed for building cross-platform security tooling.
      <br>
      <a href="https://github.com/TheManticoreProject/Manticore/actions/workflows/unit_tests.yaml" title="Build"><img alt="Build and Release" src="https://github.com/TheManticoreProject/Manticore/actions/workflows/unit_tests.yaml/badge.svg"></a>
      <img alt="GitHub release (latest by date)" src="https://img.shields.io/github/v/release/TheManticoreProject/Manticore">
      <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/TheManticoreProject/Manticore">
      <a href="https://twitter.com/intent/follow?screen_name=podalirius_" title="Follow"><img src="https://img.shields.io/twitter/follow/podalirius_?label=Podalirius&style=social"></a>
      <a href="https://www.youtube.com/c/Podalirius_?sub_confirmation=1" title="Subscribe"><img alt="YouTube Channel Subscribers" src="https://img.shields.io/youtube/channel/subscribers/UCF_x5O7CSfr82AfNVTKOv_A?style=social"></a>
      <br>
</p>

## Features

- [x] **Cross-Platform Support**: Works on Windows, Linux, and macOS.
- [x] **Multiple Authentication Protocols**: Supports NTLM, Kerberos (soon), and LDAP authentication.
- [x] **Cryptography**: [cmac](crypto/cmac/), [dcc](crypto/dcc/), [dcc2](crypto/dcc2/), [gppp](crypto/gppp/), [lm](crypto/lm/), [md4](crypto/md4/), [nt](crypto/nt/), [ntlmv1](crypto/ntlmv1/), [ntlmv2](crypto/ntlmv2/), [pkcs7](crypto/pkcs7/), [rc4](crypto/rc4/), [uuid](crypto/uuid/)
- [x] **Network Protocol Implementations**: Includes SMB, LDAP, and other common Windows protocols.
- [x] **Extensible Architecture**: Easily add new modules and functionality.

## Installation

To use this framework you can either download the latest release from the [GitHub release page](https://github.com/TheManticoreProject/Manticore/releases) or add it to your project with the following `go` command:

```bash
go get github.com/TheManticoreProject/Manticore@latest
```

## Contributing

Pull requests are welcome. Feel free to open an issue if you want to add other features.

## Credits
  - [Remi GASCOU (Podalirius)](https://github.com/p0dalirius) for the creation of the [Manticore](https://github.com/TheManticoreProject/Manticore) project.
