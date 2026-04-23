// Package report renders prospect reports as JSON and HTML files.
package report

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"time"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

// Report is the complete output document containing scored prospects.
type Report struct {
	GeneratedAt time.Time                 `json:"generated_at"`
	Entries     []explorer.ProspectReport `json:"entries"`
}

// New creates a Report from use-case output.
func New(entries []explorer.ProspectReport) *Report {
	return &Report{GeneratedAt: time.Now().UTC(), Entries: entries}
}

// WriteJSON writes the report as indented JSON to w.
func (r *Report) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(r); err != nil {
		return fmt.Errorf("encode json report: %w", err)
	}
	return nil
}

// WriteHTML writes the report as a self-contained HTML page to w.
func (r *Report) WriteHTML(w io.Writer) error {
	if err := htmlTpl.Execute(w, r); err != nil {
		return fmt.Errorf("execute html template: %w", err)
	}
	return nil
}

var htmlTpl = template.Must(template.New("report").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Prospect Report — {{.GeneratedAt.Format "2006-01-02 15:04 UTC"}}</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: system-ui, sans-serif; max-width: 900px; margin: 2rem auto; padding: 0 1rem; color: #1a1a1a; }
  h1 { margin-bottom: 0.5rem; }
  .meta { color: #666; margin-bottom: 2rem; }
  .card { border: 1px solid #ddd; border-radius: 8px; padding: 1.25rem; margin-bottom: 1rem; }
  .card h2 { font-size: 1.1rem; margin-bottom: 0.25rem; }
  .card .headline { color: #555; font-size: 0.9rem; }
  .scores { display: flex; gap: 1rem; margin: 0.75rem 0; font-size: 0.85rem; }
  .scores span { background: #f0f0f0; padding: 0.25rem 0.5rem; border-radius: 4px; }
  .reasoning { font-size: 0.85rem; color: #444; margin-bottom: 0.75rem; }
  .message { background: #f8f8f0; border-left: 3px solid #4a90d9; padding: 0.75rem; font-size: 0.95rem; position: relative; }
  .copy-btn { position: absolute; top: 0.5rem; right: 0.5rem; cursor: pointer; background: #4a90d9; color: #fff; border: none; border-radius: 4px; padding: 0.25rem 0.5rem; font-size: 0.75rem; }
  .copy-btn:hover { background: #357abd; }
  .empty { text-align: center; color: #888; padding: 3rem; }
</style>
</head>
<body>
<h1>Prospect Report</h1>
<p class="meta">Generated {{.GeneratedAt.Format "2006-01-02 15:04 UTC"}} · {{len .Entries}} prospect{{if ne (len .Entries) 1}}s{{end}}</p>
{{if not .Entries}}<p class="empty">No drafted prospects to display.</p>{{end}}
{{range $i, $e := .Entries}}
<div class="card">
  <h2><a href="{{$e.ProfileURL}}" target="_blank" rel="noopener">{{$e.Name}}</a></h2>
  <p class="headline">{{$e.Headline}} · {{$e.Location}}</p>
  <div class="scores">
    <span>Score: {{$e.WorthinessScore}}/10</span>
    <span>Critique: {{$e.CritiqueScore}}/15</span>
  </div>
  {{if $e.ScoreReasoning}}<p class="reasoning">{{$e.ScoreReasoning}}</p>{{end}}
  <div class="message" id="msg-{{$i}}">
    {{$e.DraftedMessage}}
    <button class="copy-btn" onclick="navigator.clipboard.writeText(document.getElementById('msg-{{$i}}').firstChild.textContent.trim())">Copy</button>
  </div>
</div>
{{end}}
</body>
</html>
`))
