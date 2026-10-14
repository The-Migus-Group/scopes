// Package render turns a parsed `netsh dhcp server dump` into output
// documents: the default ASCII tables, CSV, JSON, and YAML.
package render

import (
	"net/netip"
	"sort"
	"strings"

	"github.com/the-migus-group/scopes/pkg/dhcp"
	"github.com/the-migus-group/scopes/pkg/dhcp/netsh"
)

// Document is the complete, ordered representation of a parsed dump. It
// mirrors the sections the table view renders and feeds every output
// format.
type Document struct {
	// Server is the DHCP server the dump came from, when known.
	Server string `json:"server,omitempty"`
	// Scopes holds one entry per scope, sorted by network address.
	Scopes []Scope `json:"scopes"`
}

// Scope is a single DHCP scope with its ranges, exclusions, options,
// and reservations.
type Scope struct {
	Scope        string        `json:"scope"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Mask         string        `json:"mask"`
	State        string        `json:"state"` // "Active" or "Inactive"
	Reserved     int           `json:"reserved"`
	IPRanges     []IPRange     `json:"ip_ranges"`
	Exclusions   []IPRange     `json:"exclusions"`
	Options      []Option      `json:"options"`
	Reservations []Reservation `json:"reservations"`
}

// IPRange is a range of IP addresses belonging to a scope.
type IPRange struct {
	Scope string `json:"scope"`
	Start string `json:"start"`
	End   string `json:"end"`
}

// Option is a DHCP option configured on a scope, with its IANA name
// resolved.
type Option struct {
	Scope string `json:"scope"`
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Reservation is a client reservation defined on a scope.
type Reservation struct {
	Scope       string `json:"scope"`
	IPAddress   string `json:"ip"`
	ClientID    string `json:"client_id"`
	Type        string `json:"type"` // "bootp", "dhcp", or "both"
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Options controls which sections the table and CSV views include.
// JSON and YAML always contain the complete document.
type Options struct {
	// ExcludeRanges omits the IP ranges and exclusions sections.
	ExcludeRanges bool
	// IncludeReservations adds the client reservations section to the
	// table view and switches CSV to the reservations table.
	IncludeReservations bool
}

// New builds a Document from parsed scopes. The result is
// deterministic: scopes are sorted by network address and every section
// is sorted within each scope, so identical dumps render identically.
func New(scopes netsh.Scopes) *Document {
	doc := &Document{Scopes: make([]Scope, 0, len(scopes))}
	networks := make([]string, 0, len(scopes))
	for network := range scopes {
		networks = append(networks, network)
	}
	sort.SliceStable(networks, func(i, j int) bool {
		return compareIPs(networks[i], networks[j]) < 0
	})
	for _, network := range networks {
		s := scopes[network]
		if doc.Server == "" {
			doc.Server = s.Server
		}
		scope := Scope{
			Scope:        s.Scope,
			Name:         s.Name,
			Description:  s.Description,
			Mask:         s.Mask,
			State:        "Inactive",
			IPRanges:     []IPRange{},
			Exclusions:   []IPRange{},
			Options:      []Option{},
			Reservations: []Reservation{},
		}
		if s.State {
			scope.State = "Active"
		}
		if s.Reservations != nil {
			scope.Reserved = len(*s.Reservations)
		}
		if s.IPRanges != nil {
			for _, r := range *s.IPRanges {
				scope.IPRanges = append(scope.IPRanges, IPRange{
					Scope: r.Scope,
					Start: r.Start,
					End:   r.End,
				})
			}
			sortRanges(scope.IPRanges)
		}
		if s.ExcludeRanges != nil {
			for _, r := range *s.ExcludeRanges {
				scope.Exclusions = append(scope.Exclusions, IPRange{
					Scope: r.Scope,
					Start: r.Start,
					End:   r.End,
				})
			}
			sortRanges(scope.Exclusions)
		}
		if s.Options != nil {
			for _, o := range *s.Options {
				scope.Options = append(scope.Options, Option{
					Scope: o.Scope,
					ID:    o.OptionID,
					Name:  dhcp.OptionName(o.OptionID),
					Type:  netsh.DHCPOptionType(o.OptionType),
					Value: o.Value,
				})
			}
			sort.SliceStable(scope.Options, func(i, j int) bool {
				if scope.Options[i].Scope != scope.Options[j].Scope {
					return scope.Options[i].Scope < scope.Options[j].Scope
				}
				if scope.Options[i].ID != scope.Options[j].ID {
					return scope.Options[i].ID < scope.Options[j].ID
				}
				return scope.Options[i].Value < scope.Options[j].Value
			})
		}
		if s.Reservations != nil {
			for _, r := range *s.Reservations {
				scope.Reservations = append(scope.Reservations, Reservation{
					Scope:       r.Scope,
					IPAddress:   r.IPAddress,
					ClientID:    r.ClientID,
					Type:        reservationType(r.ReservationType),
					Name:        r.Name,
					Description: r.Description,
				})
			}
			sort.SliceStable(scope.Reservations, func(i, j int) bool {
				if c := compareIPs(scope.Reservations[i].IPAddress, scope.Reservations[j].IPAddress); c != 0 {
					return c < 0
				}
				return scope.Reservations[i].ClientID < scope.Reservations[j].ClientID
			})
		}
		doc.Scopes = append(doc.Scopes, scope)
	}
	return doc
}

// reservationType names a netsh reservation type constant.
func reservationType(t uint8) string {
	switch t {
	case netsh.ReservationTypeBootp:
		return "bootp"
	case netsh.ReservationTypeDhcp:
		return "dhcp"
	case netsh.ReservationTypeBoth:
		return "both"
	default:
		return "unknown"
	}
}

func sortRanges(ranges []IPRange) {
	sort.SliceStable(ranges, func(i, j int) bool {
		if c := compareIPs(ranges[i].Start, ranges[j].Start); c != 0 {
			return c < 0
		}
		return compareIPs(ranges[i].End, ranges[j].End) < 0
	})
}

// compareIPs orders dotted-quad addresses numerically; values that do
// not parse as IP addresses fall back to lexicographic order.
func compareIPs(a, b string) int {
	addrA, errA := netip.ParseAddr(a)
	addrB, errB := netip.ParseAddr(b)
	switch {
	case errA == nil && errB == nil:
		return addrA.Compare(addrB)
	case errA == nil:
		return -1
	case errB == nil:
		return 1
	default:
		return strings.Compare(a, b)
	}
}
