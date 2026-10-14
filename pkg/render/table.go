package render

import (
	"fmt"
	"io"
	"strconv"

	"github.com/olekukonko/tablewriter"
)

// Section headers, spelled out in code points so the bytes match the
// original output exactly regardless of editor emoji handling.
const (
	scopesHeader       = "\n\U0001F5C2\uFE0F Scopes\n"
	ipRangesHeader     = "\n\U0001F4CA IP Ranges\n"
	exclusionsHeader   = "\n\U0001F6AB Exclusions\n"
	optionsHeader      = "\n\u2699\uFE0F DHCP Options\n"
	reservationsHeader = "\n\U0001F3AC Client Reservations\n"
)

// Table renders the document as the human-readable ASCII tables.
func Table(w io.Writer, doc *Document, opts Options) {
	scopesTable := tablewriter.NewWriter(w)
	scopesTable.SetHeader([]string{"Scope", "Name", "Description", "Reserved", "State"})
	for _, scope := range doc.Scopes {
		scopesTable.Append([]string{
			scope.Scope,
			scope.Name,
			scope.Description,
			strconv.Itoa(scope.Reserved),
			scope.State,
		})
	}
	fmt.Fprint(w, scopesHeader)
	scopesTable.Render()

	if !opts.ExcludeRanges {
		ipRangesTable := tablewriter.NewWriter(w)
		ipRangesTable.SetHeader([]string{"Scope", "IP Start", "IP End"})
		for _, scope := range doc.Scopes {
			for _, ipr := range scope.IPRanges {
				ipRangesTable.Append([]string{ipr.Scope, ipr.Start, ipr.End})
			}
		}
		fmt.Fprint(w, ipRangesHeader)
		ipRangesTable.Render()

		exclusionsTable := tablewriter.NewWriter(w)
		exclusionsTable.SetHeader([]string{"Scope", "IP Start", "IP End"})
		for _, scope := range doc.Scopes {
			for _, ipr := range scope.Exclusions {
				exclusionsTable.Append([]string{ipr.Scope, ipr.Start, ipr.End})
			}
		}
		fmt.Fprint(w, exclusionsHeader)
		exclusionsTable.Render()
	}

	optionsTable := tablewriter.NewWriter(w)
	optionsTable.SetHeader([]string{"Scope", "Option ID", "Option Type", "Option"})
	for _, scope := range doc.Scopes {
		for _, o := range scope.Options {
			optionsTable.Append([]string{o.Scope, o.Name, o.Type, o.Value})
		}
	}
	fmt.Fprint(w, optionsHeader)
	optionsTable.Render()

	if !opts.IncludeReservations {
		return
	}

	reservationsTable := tablewriter.NewWriter(w)
	reservationsTable.SetHeader([]string{"Scope", "IP Address", "Client ID", "Name", "Description"})
	for _, scope := range doc.Scopes {
		for _, reservation := range scope.Reservations {
			reservationsTable.Append([]string{
				reservation.Scope,
				reservation.IPAddress,
				reservation.ClientID,
				reservation.Name,
				reservation.Description,
			})
		}
	}
	fmt.Fprint(w, reservationsHeader)
	reservationsTable.Render()
}
