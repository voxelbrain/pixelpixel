package main

import (
	"mime"
	"net/http"
	"path/filepath"
	"text/template"
)

const (
	borderSize = 30
	spacing    = 10
	pixelSize  = 256
)

func TemplateData() interface{} {
	return map[string]interface{}{
		"NumPixelsPerRow": options.NumPixelsPerRow,
		"Spacing":         spacing,
		"PixelSize":       pixelSize,
		"BorderSize":      borderSize,
		"TotalWidth":      (pixelSize + spacing + 2*borderSize) * options.NumPixelsPerRow,
	}
}

type templateRenderer struct {
	Dir  string
	Data interface{}
}

func (tr templateRenderer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	contentType := mime.TypeByExtension(filepath.Ext(r.URL.Path))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	path := filepath.Join(tr.Dir, r.URL.Path) + ".tpl"
	tpl, err := template.ParseFiles(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	tpl.Execute(w, tr.Data)
}
