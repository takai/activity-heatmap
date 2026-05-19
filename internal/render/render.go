// Package render produces the static HTML heatmap report.
package render

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/takai/activity-heatmap/internal/store"
)

//go:embed template.html
var tmplSource string

var tmpl = template.Must(template.New("heatmap").Parse(tmplSource))

type viewModel struct {
	Title            string
	ChannelLabel     string
	Timezone         string
	LiveVideos       int
	TotalSavedVideos int
	GeneratedAt      string
	DataJSON         template.JS
}

// HTML renders the heatmap HTML for the given DataFile.
func HTML(df store.DataFile) ([]byte, error) {
	raw, err := json.Marshal(df)
	if err != nil {
		return nil, err
	}
	label := df.Channel.ID
	if df.Channel.Handle != nil && *df.Channel.Handle != "" {
		label = *df.Channel.Handle + " (" + df.Channel.ID + ")"
	}
	vm := viewModel{
		Title:            df.Channel.Title,
		ChannelLabel:     label,
		Timezone:         df.Timezone,
		LiveVideos:       df.Summary.LiveVideos,
		TotalSavedVideos: df.Summary.TotalSavedVideos,
		GeneratedAt:      df.GeneratedAt.Format("2006-01-02 15:04:05 MST"),
		DataJSON:         template.JS(raw),
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vm); err != nil {
		return nil, fmt.Errorf("render template: %w", err)
	}
	return buf.Bytes(), nil
}
