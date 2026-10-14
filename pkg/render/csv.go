package render

import (
	"encoding/csv"
	"io"
	"strconv"
)

// CSV renders the document as a single flat CSV table: the scopes
// overview by default, or the client reservations table when
// opts.IncludeReservations is set.
func CSV(w io.Writer, doc *Document, opts Options) error {
	out := csv.NewWriter(w)
	if opts.IncludeReservations {
		out.Write([]string{"Scope", "IP Address", "Client ID", "Name", "Description"})
		for _, scope := range doc.Scopes {
			for _, reservation := range scope.Reservations {
				out.Write([]string{
					reservation.Scope,
					reservation.IPAddress,
					reservation.ClientID,
					reservation.Name,
					reservation.Description,
				})
			}
		}
	} else {
		out.Write([]string{"Scope", "Name", "Description", "Reserved", "State"})
		for _, scope := range doc.Scopes {
			out.Write([]string{
				scope.Scope,
				scope.Name,
				scope.Description,
				strconv.Itoa(scope.Reserved),
				scope.State,
			})
		}
	}
	out.Flush()
	return out.Error()
}
