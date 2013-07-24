package main

import (
	"log"
	"math/rand"
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
		Lxc             bool          `goptions:"-x, --lxc, description='Use lxc containers for pixels'"`
		Help            goptions.Help `goptions:"-h, --help, description='Show this help'"`
	}{
		Listen:          "localhost:8080",
		StaticDir:       "./static",
		TemplateDir:     "./templates",
		NumPixelsPerRow: 4,
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	goptions.ParseAndFail(&options)

	if options.Lxc {
		log.Fatalf("LXC support not implemented yet")
	}

	pa := NewPixelApi(NewLocalContainerCreator())

	r := mux.NewRouter()
	r.PathPrefix("/ws").Handler(NewStreamingHandler(pa))
	r.PathPrefix("/templates").Methods("GET").Handler(http.StripPrefix("/templates", templateRenderer{
		Dir:  options.TemplateDir,
		Data: TemplateData(),
	}))

	r.PathPrefix("/pixels").Handler(http.StripPrefix("/pixels", pa))

	r.PathPrefix("/").Methods("GET").Handler(http.FileServer(http.Dir(options.StaticDir)))

	log.Printf("Starting webserver on %s...", options.Listen)
	err := http.ListenAndServe(options.Listen, r)
	if err != nil {
		log.Fatalf("Could not start webserver: %s", err)
	}
}

func NewStreamingHandler(pa *PixelApi) websocket.Handler {
	f := NewFanout(pa.Messages)
	return websocket.Handler(func(c *websocket.Conn) {
		messages := f.Output()
		defer f.Close(messages)

		go func() {
			buf := make([]byte, 16)
			for {
				_, err := c.Read(buf)
				if err != nil {
					f.Close(messages)
					return
				}
			}
		}()

		func() {
			pa.RLock()
			defer pa.RUnlock()
			for _, pixel := range pa.pixels {
				websocket.JSON.Send(c, &Message{
					Pixel: pixel.Id,
					Type:  TypeCreate,
				})
				if !pixel.IsRunning() {
					websocket.JSON.Send(c, &Message{
						Pixel: pixel.Id,
						Type:  TypeFailure,
					})
				}
			}
		}()
		for {
			select {
			case message, ok := <-messages:
				if !ok {
					return
				}
				websocket.JSON.Send(c, message)
			}
		}
	})
}
