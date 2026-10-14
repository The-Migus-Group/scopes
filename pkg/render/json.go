package render

import (
	"encoding/json"
	"io"
)

// JSON renders the complete document as indented JSON. Unlike the table
// and CSV views it always includes every section, and empty sections are
// emitted as [] so the schema stays stable for diffing.
func JSON(w io.Writer, doc *Document) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(doc)
}
