package netsh

import "fmt"

// DHCPOption represents a DHCP standard option defined at the Scope level
type DHCPOption struct {
	Scoped
	OptionID   int
	OptionType uint8
	Value      string
}

const (
	// DHCPOptionTypeDWORD represents a 32-bit unsigned integer
	DHCPOptionTypeDWORD = iota
	// DHCPOptionTypeSTRING is an option type that represents a string
	DHCPOptionTypeSTRING
	// DHCPOptionTypeIPADDRESS is a dotted-quad IP address, e.g., 1.2.3.4
	DHCPOptionTypeIPADDRESS
)

// DHCPOptionType returns a string representation of the Microsoft Option Type
func DHCPOptionType(optionType uint8) string {
	switch optionType {
	case DHCPOptionTypeDWORD:
		return "Integer"
	case DHCPOptionTypeSTRING:
		return "Text"
	case DHCPOptionTypeIPADDRESS:
		return "IPAddress"
	default:
		return fmt.Sprintf("Unknown (%d)", optionType)
	}
}
