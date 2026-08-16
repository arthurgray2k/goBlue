package render

import (
	"encoding/json"
	"io"
)

// RenderJSON writes indented JSON representation of data to writer.
func RenderJSON(w io.Writer, data interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
