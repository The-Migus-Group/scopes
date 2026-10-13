package netsh

import (
	"strings"
	"testing"
)

func TestParseScopes(t *testing.T) {
	reader := strings.NewReader(`
Dhcp Server \\ULTRA.engops.io add scope 192.168.65.0 255.255.255.0 "x" "x, a.b.c., 'XyZ'"`)
	scope := "192.168.65.0"
	scopes, _ := ParseScopes(reader)
	if len(scopes) != 1 {
		t.Errorf("Expected 1 scope, got %d", len(scopes))
	}
	if scopes[scope].Name != "x" {
		t.Errorf("Expected scope name 'x', got '%s'", scopes[scope].Name)
	}
	if scopes[scope].Description != "x, a.b.c., 'XyZ'" {
		t.Errorf(
			"Expected scope description 'x, a.b.c., `'XyZ`', got '%s'",
			scopes[scope].Description,
		)
	}
}

func TestParseScopesOptions(t *testing.T) {
	reader := strings.NewReader(`
Dhcp Server \\ULTRA.engops.io add scope 192.168.65.0 255.255.255.0 "" ""
Dhcp Server \\ULTRA.engops.io Scope 192.168.65.0 set optionvalue 51 DWORD "7200"
Dhcp Server \\ULTRA.engops.io Scope 192.168.65.0 set optionvalue 3 IPADDRESS "192.168.65.1"
Dhcp Server \\ULTRA.engops.io Scope 192.168.65.0 set optionvalue 1 IPADDRESS "255.255.255.0"
Dhcp Server \\ULTRA.engops.io Scope 192.168.65.0 set optionvalue 6 IPADDRESS "192.168.65.2" `)

	scopes, _ := ParseScopes(reader)
	scope := scopes["192.168.65.0"]
	if len(*scope.Options) != 4 {
		t.Errorf("Expected 4 options, got %d", len(*scope.Options))
	}
	if (*scope.Options)[0].OptionID != 51 {
		t.Errorf("Expected option ID 51, got %d", (*scope.Options)[0].OptionID)
	}
	if (*scope.Options)[0].OptionType != DHCPOptionTypeDWORD {
		t.Errorf(
			"Expected option type 'DWORD', got '%s'",
			DHCPOptionType((*scope.Options)[0].OptionType),
		)
	}
	if (*scope.Options)[0].Value != "7200" {
		t.Errorf("Expected option value '7200', got '%s'", (*scope.Options)[0].Value)
	}
	if (*scope.Options)[1].OptionID != 3 {
		t.Errorf("Expected option ID 3, got %d", (*scope.Options)[1].OptionID)
	}
	if (*scope.Options)[1].OptionType != DHCPOptionTypeIPADDRESS {
		t.Errorf(
			"Expected option type 'IPADDRESS', got '%s'",
			DHCPOptionType((*scope.Options)[1].OptionType),
		)
	}
}

func TestParseScopesReservations(t *testing.T) {
	reader := strings.NewReader(`
Dhcp Server \\ULTRA.engops.io add scope 192.168.65.0 255.255.255.0 "" ""
Dhcp Server \\ULTRA.engops.io Scope 192.168.65.0 Add reservedip 192.168.65.5 00155d020150 "terrific.engops.io" "" "BOTH"
Dhcp Server \\ULTRA.engops.io Scope 192.168.65.0 Add reservedip 192.168.65.11 00155d020162 "reallynice.engops.io" "x 'yz' abc..." ""
Dhcp Server \\ULTRA.engops.io Scope 192.168.65.0 Add reservedip 192.168.65.6 00155d020160 "lovely.engops.io" "" "DHCP"`)
	scopes, _ := ParseScopes(reader)
	scope := scopes["192.168.65.0"]
	if len(*scope.Reservations) != 3 {
		t.Errorf("Expected 3 reservations, got %d", len(*scope.Reservations))
	}
	if (*scope.Reservations)[0].IPAddress != "192.168.65.5" {
		t.Errorf(
			"Expected reservation IP address '192.168.65.5', got '%s'",
			(*scope.Reservations)[0].IPAddress,
		)
	}
	if (*scope.Reservations)[1].ClientID != "00155d020162" {
		t.Errorf(
			"Expected reservation client ID '00155d020162', got '%s'",
			(*scope.Reservations)[0].ClientID,
		)
	}
	if (*scope.Reservations)[2].Name != "lovely.engops.io" {
		t.Errorf(
			"Expected reservation name 'lovely.engops.io', got '%s'",
			(*scope.Reservations)[0].Name,
		)
	}
	if (*scope.Reservations)[1].Description != "x 'yz' abc..." {
		t.Errorf(
			"Expected reservation description 'x `'yz`' abc...', got '%s'",
			(*scope.Reservations)[1].Description,
		)
	}
	if (*scope.Reservations)[2].ReservationType != ReservationTypeDhcp {
		t.Errorf(
			"Expected reservation type '%d', got '%d'",
			ReservationTypeDhcp,
			(*scope.Reservations)[0].ReservationType,
		)
	}
}

func TestParseScopesExclusions(t *testing.T) {
	reader := strings.NewReader(`
Dhcp Server \\ULTRA.engops.io add scope 192.168.66.0 255.255.254.0 "" ""
Dhcp Server \\ULTRA.engops.io Scope 192.168.66.0 add excluderange 192.168.66.1 192.168.66.25
Dhcp Server \\ULTRA.engops.io Scope 192.168.66.0 add excluderange 192.168.67.240 192.168.67.254`)
	scopes, _ := ParseScopes(reader)
	scope := scopes["192.168.66.0"]
	if len(*scope.ExcludeRanges) != 2 {
		t.Errorf("Expected 2 exclude ranges, got %d", len(*scope.ExcludeRanges))
	}
	if (*scope.ExcludeRanges)[0].Start != "192.168.66.1" {
		t.Errorf(
			"Expected exclude range start '192.168.66.1', got '%s'",
			(*scope.ExcludeRanges)[0].Start,
		)
	}
	if (*scope.ExcludeRanges)[1].End != "192.168.67.254" {
		t.Errorf(
			"Expected exclude range end '192.168.67.254', got '%s'",
			(*scope.ExcludeRanges)[1].End,
		)
	}
}

func TestParseScopesState(t *testing.T) {
	reader := strings.NewReader(`
Dhcp Server \\ULTRA.engops.io add scope 192.168.65.0 255.255.255.0 "" ""
Dhcp Server \\ULTRA.engops.io Scope 192.168.65.0 set state 0`)
	scopes, _ := ParseScopes(reader)
	if scope, ok := scopes["192.168.65.0"]; !ok {
		t.Error("Expected scope to be defined but it was not")
	} else if scope.State {
		t.Errorf("Expected scope state 'false', got '%t'", scope.State)
	}
}

func TestParseScopesStateActive(t *testing.T) {
	reader := strings.NewReader(`
Dhcp Server \\ULTRA.engops.io add scope 192.168.65.0 255.255.255.0 "" ""
Dhcp Server \\ULTRA.engops.io Scope 192.168.65.0 set state 1`)
	scopes, err := ParseScopes(reader)
	if err != nil {
		t.Fatalf("ParseScopes failed: %v", err)
	}
	scope, ok := scopes["192.168.65.0"]
	if !ok {
		t.Fatal("Expected scope to be defined but it was not")
	}
	if !scope.State {
		t.Error("Expected scope state 'true', got 'false'")
	}
}

func TestParseScopesOrphanDetailLines(t *testing.T) {
	// Detail lines referencing a scope that was never added must be skipped,
	// not dereferenced (previously a nil-pointer panic).
	reader := strings.NewReader(`
Dhcp Server \\ULTRA.engops.io Scope 192.168.65.0 Add reservedip 192.168.65.5 00155d020150 "terrific.engops.io" "" "BOTH"
Dhcp Server \\ULTRA.engops.io Scope 192.168.65.0 set optionvalue 51 DWORD "7200"
Dhcp Server \\ULTRA.engops.io Scope 192.168.65.0 Add iprange 192.168.65.1 192.168.65.254
Dhcp Server \\ULTRA.engops.io Scope 192.168.65.0 add excluderange 192.168.65.1 192.168.65.10
Dhcp Server \\ULTRA.engops.io Scope 192.168.65.0 set state 1`)
	scopes, err := ParseScopes(reader)
	if err != nil {
		t.Fatalf("ParseScopes failed: %v", err)
	}
	if len(scopes) != 0 {
		t.Errorf("Expected 0 scopes, got %d", len(scopes))
	}
}
