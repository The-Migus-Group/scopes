package render

import (
	"io"
	"strconv"
	"strings"
)

// YAML renders the complete document as YAML. The emitter is
// deliberately small and conservative — mappings, sequences, and
// plain-or-double-quoted scalars only — so the tool keeps working with
// no dependencies beyond the standard library and tablewriter.
func YAML(w io.Writer, doc *Document) error {
	var b strings.Builder
	if doc.Server != "" {
		b.WriteString("server: " + yamlScalar(doc.Server) + "\n")
	}
	if len(doc.Scopes) == 0 {
		b.WriteString("scopes: []\n")
	} else {
		b.WriteString("scopes:\n")
		for i := range doc.Scopes {
			writeYAMLScope(&b, &doc.Scopes[i])
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func writeYAMLScope(b *strings.Builder, scope *Scope) {
	b.WriteString("  - scope: " + yamlScalar(scope.Scope) + "\n")
	b.WriteString("    name: " + yamlScalar(scope.Name) + "\n")
	b.WriteString("    description: " + yamlScalar(scope.Description) + "\n")
	b.WriteString("    mask: " + yamlScalar(scope.Mask) + "\n")
	b.WriteString("    state: " + yamlScalar(scope.State) + "\n")
	b.WriteString("    reserved: " + strconv.Itoa(scope.Reserved) + "\n")
	writeYAMLRanges(b, "ip_ranges", scope.IPRanges)
	writeYAMLRanges(b, "exclusions", scope.Exclusions)

	if len(scope.Options) == 0 {
		b.WriteString("    options: []\n")
	} else {
		b.WriteString("    options:\n")
		for i := range scope.Options {
			o := &scope.Options[i]
			b.WriteString("      - id: " + strconv.Itoa(o.ID) + "\n")
			b.WriteString("        name: " + yamlScalar(o.Name) + "\n")
			b.WriteString("        type: " + yamlScalar(o.Type) + "\n")
			b.WriteString("        value: " + yamlScalar(o.Value) + "\n")
		}
	}

	if len(scope.Reservations) == 0 {
		b.WriteString("    reservations: []\n")
	} else {
		b.WriteString("    reservations:\n")
		for i := range scope.Reservations {
			r := &scope.Reservations[i]
			b.WriteString("      - scope: " + yamlScalar(r.Scope) + "\n")
			b.WriteString("        ip: " + yamlScalar(r.IPAddress) + "\n")
			b.WriteString("        client_id: " + yamlScalar(r.ClientID) + "\n")
			b.WriteString("        type: " + yamlScalar(r.Type) + "\n")
			b.WriteString("        name: " + yamlScalar(r.Name) + "\n")
			b.WriteString("        description: " + yamlScalar(r.Description) + "\n")
		}
	}
}

func writeYAMLRanges(b *strings.Builder, key string, ranges []IPRange) {
	b.WriteString("    " + key + ":")
	if len(ranges) == 0 {
		b.WriteString(" []\n")
		return
	}
	b.WriteString("\n")
	for _, r := range ranges {
		b.WriteString("      - scope: " + yamlScalar(r.Scope) + "\n")
		b.WriteString("        start: " + yamlScalar(r.Start) + "\n")
		b.WriteString("        end: " + yamlScalar(r.End) + "\n")
	}
}

// yamlScalar renders s as a plain YAML scalar when that is unambiguous
// and a double-quoted scalar otherwise. Quoting is conservative:
// anything that could be misread as a bool, null, number, indicator, or
// comment gets quoted, so parsing the output always yields the original
// strings.
func yamlScalar(s string) string {
	if needsYAMLQuote(s) {
		return strconv.Quote(s)
	}
	return s
}

func needsYAMLQuote(s string) bool {
	if s == "" {
		return true
	}
	// Strings that resolve to booleans or null must be quoted.
	switch strings.ToLower(s) {
	case "true", "false", "yes", "no", "on", "off", "y", "n", "null", "nil", "none", "~":
		return true
	}
	// Strings that resolve to numbers must be quoted.
	if _, err := strconv.ParseInt(s, 0, 64); err == nil {
		return true
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}
	// Indicator characters, comments, and flow syntax need quoting.
	if strings.ContainsAny(s, ":#{}[],&*!|>'\"%@`\\") {
		return true
	}
	// Leading or trailing whitespace needs quoting.
	if first, last := s[0], s[len(s)-1]; first == ' ' || first == '\t' || last == ' ' || last == '\t' {
		return true
	}
	// A leading indicator character needs quoting.
	if strings.ContainsAny(s[:1], "-?:.!&*|>%@`{}[],#\"'") {
		return true
	}
	// Control characters need quoting.
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return true
		}
	}
	return false
}
