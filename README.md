# scopes

`scopes` parses the text output of `netsh.exe dhcp server dump` — the only
portable export a Windows DHCP server offers — and renders it as readable
ASCII tables: scopes, IP ranges, exclusions, DHCP options (with IANA
option-name lookup), and client reservations.

A `netsh dhcp server dump` can only be replayed by `netsh exec` on another
Windows DHCP server. This tool turns the same dump into something a human can
read from any platform. It is a single static binary with no dependencies
beyond the Go standard library and tablewriter.

## Install

With Go 1.21 or newer:

```sh
go install github.com/the-migus-group/scopes/cmd/scopes@latest
```

Prebuilt binaries for Linux, macOS, and Windows (amd64/arm64) are attached to
each [release](https://github.com/the-migus-group/scopes/releases), with
GPG-signed SHA-256 checksums.

## Usage

Pipe a dump straight from a Windows DHCP server:

```PowerShell
$DHCPServer = New-PSSession DC1
Invoke-Command $DHCPServer { netsh.exe dhcp server dump } | .\scopes.exe -R
```

Or parse a dump saved to a file (`netsh dhcp server dump > dump.txt`):

```sh
scopes -f dump.txt
```

| Flag | Effect |
| ---- | ------ |
| `-f <file>` | Read the dump from a file instead of stdin |
| `-q` | Omit the IP ranges and exclusions tables |
| `-R` | Include the client reservations table |
| `-v` | Print the version and exit |

## License

[MIT](LICENSE)
