package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/the-migus-group/scopes/pkg/dhcp/netsh"
)

func ptr[T any](v T) *T { return &v }

func testScopes() netsh.Scopes {
	return netsh.Scopes{
		"192.168.2.0": {
			Scoped: netsh.Scoped{Server: "DC2", Scope: "192.168.2.0"},
			Mask:   "255.255.255.0",
			Name:   "Second",
			IPRanges: ptr([]netsh.IPRange{
				{
					Scoped: netsh.Scoped{Scope: "192.168.2.0"},
					Start:  "192.168.2.100",
					End:    "192.168.2.200",
				},
				{
					Scoped: netsh.Scoped{Scope: "192.168.2.0"},
					Start:  "192.168.2.10",
					End:    "192.168.2.20",
				},
			}),
			ExcludeRanges: ptr([]netsh.IPRange{
				{
					Scoped: netsh.Scoped{Scope: "192.168.2.0"},
					Start:  "192.168.2.150",
					End:    "192.168.2.160",
				},
			}),
			Options: ptr([]netsh.DHCPOption{
				{
					Scoped:     netsh.Scoped{Scope: "192.168.2.0"},
					OptionID:   6,
					OptionType: netsh.DHCPOptionTypeIPADDRESS,
					Value:      "10.0.0.6",
				},
				{
					Scoped:     netsh.Scoped{Scope: "192.168.2.0"},
					OptionID:   3,
					OptionType: netsh.DHCPOptionTypeIPADDRESS,
					Value:      "192.168.2.1",
				},
				{
					Scoped:     netsh.Scoped{Scope: "192.168.2.0"},
					OptionID:   15,
					OptionType: netsh.DHCPOptionTypeSTRING,
					Value:      "corp.example",
				},
			}),
			Reservations: ptr([]netsh.Reservation{
				{
					Scoped:          netsh.Scoped{Scope: "192.168.2.0"},
					IPAddress:       "192.168.2.50",
					ClientID:        "00aa11bb22cc",
					ReservationType: netsh.ReservationTypeBoth,
					Name:            "printer",
					Description:     "HP LaserJet",
				},
			}),
		},
		"10.0.0.0": {
			Scoped:        netsh.Scoped{Server: "DC1", Scope: "10.0.0.0"},
			Mask:          "255.0.0.0",
			Name:          "Core",
			State:         true,
			IPRanges:      ptr([]netsh.IPRange{{Scoped: netsh.Scoped{Scope: "10.0.0.0"}, Start: "10.0.0.10", End: "10.0.0.240"}}),
			ExcludeRanges: ptr([]netsh.IPRange{}),
			Options:       ptr([]netsh.DHCPOption{}),
			Reservations:  ptr([]netsh.Reservation{}),
		},
	}
}

func TestNewSortsScopesAndSections(t *testing.T) {
	doc := New(testScopes())
	if doc.Server != "DC1" {
		t.Errorf("Server = %q, want %q", doc.Server, "DC1")
	}
	if len(doc.Scopes) != 2 {
		t.Fatalf("len(Scopes) = %d, want 2", len(doc.Scopes))
	}
	if doc.Scopes[0].Scope != "10.0.0.0" || doc.Scopes[1].Scope != "192.168.2.0" {
		t.Errorf("scopes not sorted numerically: %q, %q", doc.Scopes[0].Scope, doc.Scopes[1].Scope)
	}
	core := doc.Scopes[0]
	if core.State != "Active" || core.Reserved != 0 {
		t.Errorf("core state/reserved = %q/%d, want Active/0", core.State, core.Reserved)
	}
	if len(core.IPRanges) != 1 || core.IPRanges[0].Start != "10.0.0.10" {
		t.Errorf("core ip_ranges = %+v", core.IPRanges)
	}
	second := doc.Scopes[1]
	if second.State != "Inactive" || second.Reserved != 1 {
		t.Errorf("second state/reserved = %q/%d, want Inactive/1", second.State, second.Reserved)
	}
	if len(second.IPRanges) != 2 ||
		second.IPRanges[0].Start != "192.168.2.10" ||
		second.IPRanges[1].Start != "192.168.2.100" {
		t.Errorf("ip_ranges not sorted numerically: %+v", second.IPRanges)
	}
	wantOptions := []Option{
		{Scope: "192.168.2.0", ID: 3, Name: "Router", Type: "IPAddress", Value: "192.168.2.1"},
		{Scope: "192.168.2.0", ID: 6, Name: "Domain Server", Type: "IPAddress", Value: "10.0.0.6"},
		{Scope: "192.168.2.0", ID: 15, Name: "Domain Name", Type: "Text", Value: "corp.example"},
	}
	if len(second.Options) != len(wantOptions) {
		t.Fatalf("len(Options) = %d, want %d", len(second.Options), len(wantOptions))
	}
	for i, want := range wantOptions {
		if second.Options[i] != want {
			t.Errorf("Options[%d] = %+v, want %+v", i, second.Options[i], want)
		}
	}
	wantReservations := []Reservation{
		{Scope: "192.168.2.0", IPAddress: "192.168.2.50", ClientID: "00aa11bb22cc", Type: "both", Name: "printer", Description: "HP LaserJet"},
	}
	if len(second.Reservations) != len(wantReservations) {
		t.Fatalf("len(Reservations) = %d, want %d", len(second.Reservations), len(wantReservations))
	}
	for i, want := range wantReservations {
		if second.Reservations[i] != want {
			t.Errorf("Reservations[%d] = %+v, want %+v", i, second.Reservations[i], want)
		}
	}
}

func TestNewNilSlicesAreSafe(t *testing.T) {
	doc := New(netsh.Scopes{
		"10.1.2.0": {
			Scoped: netsh.Scoped{Server: "DC1", Scope: "10.1.2.0"},
			Name:   "Bare",
			State:  true,
		},
	})
	scope := doc.Scopes[0]
	if scope.Reserved != 0 {
		t.Errorf("Reserved = %d, want 0", scope.Reserved)
	}
	if scope.IPRanges == nil || len(scope.IPRanges) != 0 {
		t.Errorf("IPRanges = %#v, want empty non-nil slice", scope.IPRanges)
	}
	if scope.Exclusions == nil || len(scope.Exclusions) != 0 {
		t.Errorf("Exclusions = %#v, want empty non-nil slice", scope.Exclusions)
	}
	if scope.Options == nil || len(scope.Options) != 0 {
		t.Errorf("Options = %#v, want empty non-nil slice", scope.Options)
	}
	if scope.Reservations == nil || len(scope.Reservations) != 0 {
		t.Errorf("Reservations = %#v, want empty non-nil slice", scope.Reservations)
	}
}

func TestCSV(t *testing.T) {
	doc := New(testScopes())

	var overview bytes.Buffer
	if err := CSV(&overview, doc, Options{}); err != nil {
		t.Fatal(err)
	}
	wantOverview := "Scope,Name,Description,Reserved,State\n" +
		"10.0.0.0,Core,,0,Active\n" +
		"192.168.2.0,Second,,1,Inactive\n"
	if overview.String() != wantOverview {
		t.Errorf("CSV overview = %q, want %q", overview.String(), wantOverview)
	}

	var reservations bytes.Buffer
	if err := CSV(&reservations, doc, Options{IncludeReservations: true}); err != nil {
		t.Fatal(err)
	}
	wantReservations := "Scope,IP Address,Client ID,Name,Description\n" +
		"192.168.2.0,192.168.2.50,00aa11bb22cc,printer,HP LaserJet\n"
	if reservations.String() != wantReservations {
		t.Errorf("CSV reservations = %q, want %q", reservations.String(), wantReservations)
	}
}

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, New(testScopes())); err != nil {
		t.Fatal(err)
	}
	want := `{
  "server": "DC1",
  "scopes": [
    {
      "scope": "10.0.0.0",
      "name": "Core",
      "description": "",
      "mask": "255.0.0.0",
      "state": "Active",
      "reserved": 0,
      "ip_ranges": [
        {
          "scope": "10.0.0.0",
          "start": "10.0.0.10",
          "end": "10.0.0.240"
        }
      ],
      "exclusions": [],
      "options": [],
      "reservations": []
    },
    {
      "scope": "192.168.2.0",
      "name": "Second",
      "description": "",
      "mask": "255.255.255.0",
      "state": "Inactive",
      "reserved": 1,
      "ip_ranges": [
        {
          "scope": "192.168.2.0",
          "start": "192.168.2.10",
          "end": "192.168.2.20"
        },
        {
          "scope": "192.168.2.0",
          "start": "192.168.2.100",
          "end": "192.168.2.200"
        }
      ],
      "exclusions": [
        {
          "scope": "192.168.2.0",
          "start": "192.168.2.150",
          "end": "192.168.2.160"
        }
      ],
      "options": [
        {
          "scope": "192.168.2.0",
          "id": 3,
          "name": "Router",
          "type": "IPAddress",
          "value": "192.168.2.1"
        },
        {
          "scope": "192.168.2.0",
          "id": 6,
          "name": "Domain Server",
          "type": "IPAddress",
          "value": "10.0.0.6"
        },
        {
          "scope": "192.168.2.0",
          "id": 15,
          "name": "Domain Name",
          "type": "Text",
          "value": "corp.example"
        }
      ],
      "reservations": [
        {
          "scope": "192.168.2.0",
          "ip": "192.168.2.50",
          "client_id": "00aa11bb22cc",
          "type": "both",
          "name": "printer",
          "description": "HP LaserJet"
        }
      ]
    }
  ]
}
`
	if buf.String() != want {
		t.Errorf("JSON mismatch:\n got: %s\nwant: %s", buf.String(), want)
	}
}

func TestJSONEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, New(netsh.Scopes{})); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "{\n  \"scopes\": []\n}\n" {
		t.Errorf("empty JSON = %q", buf.String())
	}
}

func TestYAML(t *testing.T) {
	var buf bytes.Buffer
	if err := YAML(&buf, New(testScopes())); err != nil {
		t.Fatal(err)
	}
	want := `server: DC1
scopes:
  - scope: 10.0.0.0
    name: Core
    description: ""
    mask: 255.0.0.0
    state: Active
    reserved: 0
    ip_ranges:
      - scope: 10.0.0.0
        start: 10.0.0.10
        end: 10.0.0.240
    exclusions: []
    options: []
    reservations: []
  - scope: 192.168.2.0
    name: Second
    description: ""
    mask: 255.255.255.0
    state: Inactive
    reserved: 1
    ip_ranges:
      - scope: 192.168.2.0
        start: 192.168.2.10
        end: 192.168.2.20
      - scope: 192.168.2.0
        start: 192.168.2.100
        end: 192.168.2.200
    exclusions:
      - scope: 192.168.2.0
        start: 192.168.2.150
        end: 192.168.2.160
    options:
      - id: 3
        name: Router
        type: IPAddress
        value: 192.168.2.1
      - id: 6
        name: Domain Server
        type: IPAddress
        value: 10.0.0.6
      - id: 15
        name: Domain Name
        type: Text
        value: corp.example
    reservations:
      - scope: 192.168.2.0
        ip: 192.168.2.50
        client_id: 00aa11bb22cc
        type: both
        name: printer
        description: HP LaserJet
`
	if buf.String() != want {
		t.Errorf("YAML mismatch:\n got: %s\nwant: %s", buf.String(), want)
	}
}

func TestYAMLEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := YAML(&buf, New(netsh.Scopes{})); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "scopes: []\n" {
		t.Errorf("empty YAML = %q", buf.String())
	}
}

func TestYAMLQuoting(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", `""`},
		{"plain", "plain"},
		{"192.168.1.10", "192.168.1.10"},
		{"1.2.3", "1.2.3"},
		{"576", `"576"`},
		{"007", `"007"`},
		{"1.5", `"1.5"`},
		{"true", `"true"`},
		{"Yes", `"Yes"`},
		{"null", `"null"`},
		{"a: b", `"a: b"`},
		{"# comment", `"# comment"`},
		{"- item", `"- item"`},
		{" trailing", `" trailing"`},
		{"[flow]", `"[flow]"`},
		{"line\nbreak", `"line\nbreak"`},
		{"a\\b", `"a\\b"`},
	}
	for _, tc := range cases {
		if got := yamlScalar(tc.in); got != tc.want {
			t.Errorf("yamlScalar(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestCompareIPs(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"10.0.0.2", "10.0.0.10", -1},
		{"10.0.0.10", "10.0.0.2", 1},
		{"10.0.0.1", "10.0.0.1", 0},
		{"192.168.1.0", "10.0.0.0", 1},
		{"10.0.0.1", "not-an-ip", -1},
		{"abc", "abd", -1},
	}
	for _, tc := range cases {
		if got := sign(compareIPs(tc.a, tc.b)); got != tc.want {
			t.Errorf("compareIPs(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}

func TestTableSections(t *testing.T) {
	doc := New(testScopes())

	var full bytes.Buffer
	Table(&full, doc, Options{})
	out := full.String()
	for _, header := range []string{scopesHeader, ipRangesHeader, exclusionsHeader, optionsHeader} {
		if !strings.Contains(out, header) {
			t.Errorf("table output missing header %q", header)
		}
	}
	if strings.Contains(out, reservationsHeader) {
		t.Error("reservations section rendered without IncludeReservations")
	}

	var withReservations bytes.Buffer
	Table(&withReservations, doc, Options{IncludeReservations: true})
	if !strings.Contains(withReservations.String(), reservationsHeader) {
		t.Error("reservations section missing with IncludeReservations")
	}

	var withoutRanges bytes.Buffer
	Table(&withoutRanges, doc, Options{ExcludeRanges: true})
	out = withoutRanges.String()
	if strings.Contains(out, ipRangesHeader) || strings.Contains(out, exclusionsHeader) {
		t.Error("range sections rendered with ExcludeRanges")
	}
	if !strings.Contains(out, optionsHeader) {
		t.Error("options section missing with ExcludeRanges")
	}
}
