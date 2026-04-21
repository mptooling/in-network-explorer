// Package report renders prospect reports as HTML or JSON.
package report

import (
	"embed"
	"encoding/json"
	"html/template"
	"io"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

//go:embed report.gohtml
var tmplFS embed.FS

var htmlTmpl = template.Must(
	template.New("report.gohtml").Funcs(template.FuncMap{
		"scoreColor": scoreColor,
		"mul": func(a float64, b float64) float64 {
			return a * b
		},
	}).ParseFS(tmplFS, "report.gohtml"),
)

// scoreColor returns a CSS class name based on the worthiness score.
func scoreColor(score int) string {
	switch {
	case score >= 8:
		return "score-high"
	case score >= 5:
		return "score-mid"
	default:
		return "score-low"
	}
}

// RenderHTML writes the report as an HTML page to w.
func RenderHTML(w io.Writer, r *explorer.ProspectReport) error {
	return htmlTmpl.Execute(w, r)
}

// RenderJSON writes the report as indented JSON to w.
func RenderJSON(w io.Writer, r *explorer.ProspectReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}
