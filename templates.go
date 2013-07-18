package main

import (
	"mime"
	"net/http"
	"path/filepath"
	"text/template"
)

const (
	border    = 30
	spacing   = 10
	pixelSize = 256
)

func TemplateData() interface{} {
	return map[string]interface{}{
		"NumPixelsPerRow": options.NumPixelsPerRow,
		"Spacing":         spacing + border,
		"PixelSize":       pixelSize,
		"TotalWidth":      (pixelSize + spacing) * options.NumPixelsPerRow,
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
