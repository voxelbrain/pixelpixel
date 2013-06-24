package main

import (
	"log"
	"net/http"
	"time"

	"code.google.com/p/go.net/websocket"
	"github.com/gorilla/mux"

	"github.com/voxelbrain/goptions"
)

var (
	options = struct {
		Listen          string        `goptions:"-l, --listen, description='Adress to bind webserver to'"`
		NumPixelsPerRow int           `goptions:"-r, --per-row, description='Number of pixels per row'"`
		TemplateDir     string        `goptions:"-t, --templates, description='Path to the templates'"`
		StaticDir       string        `goptions:"--static, description='Path to the static content'"`
		Help            goptions.Help `goptions:"-h, --help, description='Show this help'"`
	}{
		Listen:          "localhost:8080",
		StaticDir:       "./static",
		TemplateDir:     "./templates",
		NumPixelsPerRow: 4,
	}
)

func main() {
	goptions.ParseAndFail(&options)

	r := mux.NewRouter()
	r.PathPrefix("/ws").Handler(websocket.Handler(streamingHandler))
	r.PathPrefix("/templates").Methods("GET").Handler(http.StripPrefix("/templates", templateRenderer{
		Dir:  options.TemplateDir,
		Data: TemplateData(),
	}))
	r.PathPrefix("/").Methods("GET").Handler(http.FileServer(http.Dir(options.StaticDir)))

	log.Printf("Starting webserver on %s...", options.Listen)
	err := http.ListenAndServe(options.Listen, r)
	if err != nil {
		log.Fatalf("Could not start webserver: %s", err)
	}
}

func streamingHandler(c *websocket.Conn) {
}
