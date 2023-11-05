package main

import (
	"flag"
	"fmt"
	"strconv"

	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/the-migus-group/go-dhcp/pkg/dhcp"
	"github.com/the-migus-group/go-dhcp/pkg/dhcp/microsoft"
)

func main() {
	var inputFilePath string

	flag.StringVar(&inputFilePath, "f", "", "The path to an input file containing the output of 'netsh.exe dhcp server dump'; defaults to stdin")
	flag.Parse()

	scopes := map[string]microsoft.Scope{}
	if inputFilePath != "" {
		if inputFile, err := os.Open(inputFilePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input '%s': %s", inputFilePath, err)
			os.Exit(1)
		} else {
			defer inputFile.Close()
			scopes = microsoft.ParseScopes(inputFile)
		}
	} else {
		scopes = microsoft.ParseScopes(os.Stdin)
	}

	scopesTable := tablewriter.NewWriter(os.Stdout)
	scopesTable.SetHeader([]string{
		"Scope",
		"Name",
		"Description",
		"Reservations",
		"State (enabled)"})

	for _, s := range scopes {
		scopesTable.Append([]string{
			s.Scope,
			s.Name,
			s.Description,
			strconv.Itoa(len(*s.Reservations)),
			strconv.FormatBool(s.State),
		})
	}
	scopesTable.Render()

	reservationsTable := tablewriter.NewWriter(os.Stdout)
	reservationsTable.SetHeader([]string{
		"Scope", "IP Address", "Client ID", "Name", "Description"})
	for _, s := range scopes {
		for _, reservation := range *s.Reservations {
			reservationsTable.Append([]string{
				reservation.Scope,
				reservation.IPAddress,
				reservation.ClientID,
				reservation.Name,
				reservation.Description,
			})
		}
	}
	reservationsTable.Render()

	optionsTable := tablewriter.NewWriter(os.Stdout)
	optionsTable.SetHeader([]string{"Scope", "Option ID", "Option Type", "Option"})
	for _, s := range scopes {
		for _, o := range *s.Options {
			optionsTable.Append([]string{
				o.Scope,
				dhcp.OptionName(o.OptionID),
				microsoft.DHCPOptionType(o.OptionType),
				o.Value,
			})
		}
	}
	optionsTable.Render()

	exclusionsTable := tablewriter.NewWriter(os.Stdout)
	exclusionsTable.SetHeader([]string{"Scope", "IP Start", "IP End"})
	for _, s := range scopes {
		for _, ipr := range *s.ExcludeRanges {
			exclusionsTable.Append([]string{
				ipr.Scope,
				ipr.Start,
				ipr.End,
			})
		}
	}
	exclusionsTable.Render()

	ipRangesTable := tablewriter.NewWriter(os.Stdout)
	ipRangesTable.SetHeader([]string{"Scope", "IP Start", "IP End"})
	for _, s := range scopes {
		for _, ipr := range *s.IPRanges {
			ipRangesTable.Append([]string{
				ipr.Scope,
				ipr.Start,
				ipr.End,
			})
		}
	}
	ipRangesTable.Render()
}
