package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/the-migus-group/scopes/pkg/dhcp/netsh"
	"github.com/the-migus-group/scopes/pkg/render"
)

// Version is stamped at build time via -ldflags "-X main.Version=...".
var Version = "dev"

func main() {
	var inputFilePath, outputFormat string
	var excludeRanges, includeReservations, showVersion bool

	flag.StringVar(&inputFilePath, "f", "", "A file containing the output of 'netsh.exe dhcp server dump'")
	flag.StringVar(&outputFormat, "o", "table", "Output format: table, csv, json, or yaml")
	flag.BoolVar(&excludeRanges, "q", false, "Exclude IP ranges (and exclusions) tables in the output")
	flag.BoolVar(&includeReservations, "R", false, "Include DHCP Client Reservations tables in the output")
	flag.BoolVar(&showVersion, "v", false, "Print version and exit")
	flag.Parse()
	if showVersion {
		fmt.Println(Version)
		return
	}
	switch outputFormat {
	case "table", "csv", "json", "yaml":
	default:
		fmt.Fprintf(os.Stderr, "Unknown output format %q (expected table, csv, json, or yaml)\n", outputFormat)
		os.Exit(1)
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

	doc := render.New(scopes)
	switch outputFormat {
	case "csv":
		err = render.CSV(os.Stdout, doc, render.Options{IncludeReservations: includeReservations})
	case "json":
		err = render.JSON(os.Stdout, doc)
	case "yaml":
		err = render.YAML(os.Stdout, doc)
	default:
		render.Table(os.Stdout, doc, render.Options{
			ExcludeRanges:       excludeRanges,
			IncludeReservations: includeReservations,
		})
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %s\n", err)
		os.Exit(1)
	}
}
