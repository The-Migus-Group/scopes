package main

import (
	"flag"
	"fmt"
	"strconv"

	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/the-migus-group/scopes/pkg/dhcp"
	"github.com/the-migus-group/scopes/pkg/dhcp/netsh"
)

// Version is stamped at build time via -ldflags "-X main.Version=...".
var Version = "dev"

func main() {
	var inputFilePath string
	var excludeRanges, includeReservations bool
	var showVersion bool

	flag.StringVar(&inputFilePath, "f", "", "A file containing the output of 'netsh.exe dhcp server dump'")
	flag.BoolVar(&excludeRanges, "q", false, "Exclude IP ranges (and exclusions) tables in the output")
	flag.BoolVar(&includeReservations, "R", false, "Include DHCP Client Reservations tables in the output")
	flag.BoolVar(&showVersion, "v", false, "Print version and exit")
	flag.Parse()
	if showVersion {
		fmt.Println(Version)
		return
	}

	scopes := map[string]netsh.Scope{}
	// Parse the scopes from either the input file or stdin
	var err error
	if inputFilePath != "" {
		var inputFile *os.File
		inputFile, err = os.Open(inputFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening '%s': %s\n", inputFilePath, err)
			os.Exit(1)
		}
		defer inputFile.Close()
		scopes, err = netsh.ParseScopes(inputFile)
	} else {
		scopes, err = netsh.ParseScopes(os.Stdin)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing input: %s\n", err)
		os.Exit(1)
	}
	// It's all table building and rendering from here
	scopesTable := tablewriter.NewWriter(os.Stdout)
	scopesTable.SetHeader([]string{
		"Scope",
		"Name",
		"Description",
		"Reserved",
		"State"})

	for _, scope := range scopes {
		state := "Active"
		if !scope.State {
			state = "Inactive"
		}
		scopesTable.Append([]string{
			scope.Scope,
			scope.Name,
			scope.Description,
			strconv.Itoa(len(*scope.Reservations)),
			state,
		})
	}
	fmt.Print("\n🗂️ Scopes\n")
	scopesTable.Render()

	if !excludeRanges {
		ipRangesTable := tablewriter.NewWriter(os.Stdout)
		ipRangesTable.SetHeader([]string{"Scope", "IP Start", "IP End"})
		for _, scope := range scopes {
			for _, ipr := range *scope.IPRanges {
				ipRangesTable.Append([]string{
					ipr.Scope,
					ipr.Start,
					ipr.End,
				})
			}
		}
		fmt.Print("\n📊 IP Ranges\n")
		ipRangesTable.Render()

		exclusionsTable := tablewriter.NewWriter(os.Stdout)
		exclusionsTable.SetHeader([]string{"Scope", "IP Start", "IP End"})
		for _, scope := range scopes {
			for _, ipr := range *scope.ExcludeRanges {
				exclusionsTable.Append([]string{
					ipr.Scope,
					ipr.Start,
					ipr.End,
				})
			}
		}
		fmt.Print("\n🚫 Exclusions\n")
		exclusionsTable.Render()
	}

	optionsTable := tablewriter.NewWriter(os.Stdout)
	optionsTable.SetHeader([]string{"Scope", "Option ID", "Option Type", "Option"})
	for _, scope := range scopes {
		for _, o := range *scope.Options {
			optionsTable.Append([]string{
				o.Scope,
				dhcp.OptionName(o.OptionID),
				netsh.DHCPOptionType(o.OptionType),
				o.Value,
			})
		}
	}
	fmt.Print("\n⚙️ DHCP Options\n")
	optionsTable.Render()

	if !includeReservations {
		return
	}

	reservationsTable := tablewriter.NewWriter(os.Stdout)
	reservationsTable.SetHeader([]string{
		"Scope", "IP Address", "Client ID", "Name", "Description"})
	for _, scope := range scopes {
		for _, reservation := range *scope.Reservations {
			reservationsTable.Append([]string{
				reservation.Scope,
				reservation.IPAddress,
				reservation.ClientID,
				reservation.Name,
				reservation.Description,
			})
		}
	}
	fmt.Print("\n🎟️ Client Reservations\n")
	reservationsTable.Render()
}
