# Microsoft (netsh) DHCP server to

A tool to process the output of `netsh.exe dhcp server dump` on a Windows DHCP server.

## Usage

The `scopes` command renders ASCII tables of Scopes from the input:

```PowerShell
$DHCPServer = New-PSSession DC1
Invoke-Command $DHCPServer { netsh.exe dhcp server dump } | .\scopes.exe -a
```
