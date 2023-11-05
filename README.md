# DHCP Data Munging

A DHCP Data Munging library and tools written in Go.

## Usage

The `scopes` command renders ASCII tables of Scopes from the input:

```PowerShell
$DHCPServer = New-PSSession DC1
Invoke-Command $DHCPServer { netsh.exe dhcp server dump } | .\scopes.exe -a
```
