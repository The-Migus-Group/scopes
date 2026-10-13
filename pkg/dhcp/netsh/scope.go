package netsh

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
)

// Scoped contextualizes an object with the server and scope it belongs to.
type Scoped struct {
	Text   string
	Server string
	Scope  string
}

// IPRange is a range of IP addresses, e.g., 1.2.3.24-1.2.3.240
type IPRange struct {
	Scoped
	Start string
	End   string
}

const (
	// ReservationTypeBootp is a BOOTP reservation
	ReservationTypeBootp = iota
	// ReservationTypeDhcp is a DHCP reservation
	ReservationTypeDhcp
	// ReservationTypeBoth is a BOOTP and DHCP reservation, i.e., "both"
	ReservationTypeBoth
)

// Reservation represents a DHCP reservation which is defined at the Scope level
type Reservation struct {
	Scoped
	IPAddress       string
	ClientID        string
	ReservationType uint8
	Name            string
	Description     string
}

// Scope contains a set of IP ranges, excluded IP ranges, options, reservations,
// and other metadata.  It is defined at the Server level.
// A Scope also has a "State" flag that reflects whether it is enabled or not.
type Scope struct {
	Scoped
	Mask          string
	ExcludeRanges *[]IPRange
	IPRanges      *[]IPRange
	Options       *[]DHCPOption
	Reservations  *[]Reservation
	Name          string
	Description   string
	State         bool
}

// Scopes is a map of Scope objects, keyed by "network," e.g., "192.168.1.0"
type Scopes map[string]Scope

// ParseScopes parses a `netsh dhcp server dump` and returns a Scopes object.
func ParseScopes(inputFile io.Reader) (Scopes, error) {
	// Parse the file line-by-line, attempting to match the regular expression
	// and calling the corresponding parsing function on success.
	// Each parsing functions accepts the current scopes object and return a new one.
	scanner := bufio.NewScanner(inputFile)
	scopes := make(Scopes)
	for scanner.Scan() {
		for regexp, parser := range map[*regexp.Regexp](func([]string, Scopes) Scopes){
			regexp.MustCompile(`^Dhcp Server ([\w\.\\]+) Scope (\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b) Add reservedip (\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b) ([A-Fa-f0-9\.\-]+) "(.*)" "(.*)" "(BOOTP|BOTH|DHCP)?"$`): func(matches []string, scopes Scopes) Scopes {
				r := Reservation{
					Scoped: Scoped{
						Text:   matches[0],
						Server: matches[1],
						Scope:  matches[2],
					},
					IPAddress:   matches[3],
					ClientID:    matches[4],
					Name:        matches[5],
					Description: matches[6],
				}
				switch matches[7] {
				case "BOOTP":
					r.ReservationType = ReservationTypeBootp
				case "DHCP":
					r.ReservationType = ReservationTypeDhcp
				case "BOTH":
					r.ReservationType = ReservationTypeBoth
				}
				if scope, ok := scopes[matches[2]]; ok && scope.Reservations != nil {
					*scope.Reservations = append(*scope.Reservations, r)
					scopes[matches[2]] = scope
				}
				return scopes
			},
			regexp.MustCompile(`^Dhcp Server ([\w\.\\]+) add scope (\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b) (\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b) "(.*)" "(.*)"$`): func(matches []string, scopes Scopes) Scopes {
				scopes[matches[2]] = Scope{
					Scoped: Scoped{
						Text:   matches[0],
						Server: matches[1],
						Scope:  matches[2],
					},
					Mask:          matches[3],
					Name:          matches[4],
					Description:   matches[5],
					ExcludeRanges: &[]IPRange{},
					IPRanges:      &[]IPRange{},
					Options:       &[]DHCPOption{},
					Reservations:  &[]Reservation{},
				}
				return scopes
			},
			regexp.MustCompile(`^Dhcp Server ([\w\.\\]+) Scope (\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b) set optionvalue (\d+) (DWORD|STRING|IPADDRESS) "(\S+)"\s*$`): func(matches []string, scopes Scopes) Scopes {
				o := DHCPOption{
					Scoped: Scoped{
						Text:   matches[0],
						Server: matches[1],
						Scope:  matches[2],
					},
					Value: matches[5],
				}
				switch matches[4] {
				case "DWORD":
					o.OptionType = DHCPOptionTypeDWORD
				case "STRING":
					o.OptionType = DHCPOptionTypeSTRING
				case "IPADDRESS":
					o.OptionType = DHCPOptionTypeIPADDRESS
				}
				o.OptionID, _ = strconv.Atoi(matches[3])
				if scope, ok := scopes[matches[2]]; ok && scope.Options != nil {
					*scope.Options = append(*scope.Options, o)
					scopes[matches[2]] = scope
				}
				return scopes
			},
			regexp.MustCompile(`^Dhcp Server ([\w\.\\]+) Scope (\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b) set state ([01])$`): func(matches []string, scopes Scopes) Scopes {
				if scope, ok := scopes[matches[2]]; ok {
					scope.State = matches[3] == "1"
					scopes[matches[2]] = scope
				}
				return scopes
			},
			regexp.MustCompile(`^Dhcp Server ([\w\.\\]+) Scope (\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b) Add iprange (\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b) (\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b)\s*$`): func(matches []string, scopes Scopes) Scopes {
				i := IPRange{
					Scoped: Scoped{
						Text:   matches[0],
						Server: matches[1],
						Scope:  matches[2],
					},
					Start: matches[3],
					End:   matches[4],
				}
				if scope, ok := scopes[matches[2]]; ok && scope.IPRanges != nil {
					*scope.IPRanges = append(*scope.IPRanges, i)
					scopes[matches[2]] = scope
				}
				return scopes
			},
			regexp.MustCompile(`^Dhcp Server ([\w\.\\]+) Scope (\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b) add excluderange (\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b) (\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b)$`): func(matches []string, scopes Scopes) Scopes {
				e := IPRange{
					Scoped: Scoped{
						Text:   matches[0],
						Server: matches[1],
						Scope:  matches[2],
					},
					Start: matches[3],
					End:   matches[4],
				}
				if scope, ok := scopes[matches[2]]; ok && scope.ExcludeRanges != nil {
					*scope.ExcludeRanges = append(*scope.ExcludeRanges, e)
					scopes[matches[2]] = scope
				}
				return scopes
			},
		} {
			// Match the line and call the corresponding parsing function.
			if matches := regexp.FindStringSubmatch(scanner.Text()); matches != nil {
				scopes = parser(matches, scopes)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return scopes, nil
}
