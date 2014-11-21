package main

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/surma/httptools"

	"github.com/voxelbrain/pixelpixel/pixelutils"
)

type PixelApi struct {
	pixels   map[string]*Pixel
	cc       ContainerCreator
	Messages chan *Message
	http.Handler
	*sync.RWMutex
}

func NewPixelApi(cc ContainerCreator) *PixelApi {
	pa := &PixelApi{
		RWMutex:  &sync.RWMutex{},
		Messages: make(chan *Message),
		pixels:   make(map[string]*Pixel),
		cc:       cc,
	}
	h := httptools.NewRegexpSwitch(map[string]http.Handler{
		"/": httptools.MethodSwitch{
			"GET":  http.HandlerFunc(pa.ListPixels),
			"POST": http.HandlerFunc(pa.CreatePixel),
		},
		"/([a-z0-9]+)(/.+)?": httptools.L{
			httptools.DiscardPathElements(1),
			httptools.SilentHandler(http.HandlerFunc(pa.ValidatePixelId)),
			httptools.NewRegexpSwitch(map[string]http.Handler{
				"/content": httptools.MethodSwitch{"GET": http.HandlerFunc(pa.GetPixelContent)},
				"/logs":    httptools.MethodSwitch{"GET": http.HandlerFunc(pa.GetPixelLogs)},
				"/fs":      httptools.MethodSwitch{"GET": http.HandlerFunc(pa.GetPixelFs)},
				"/": httptools.MethodSwitch{
					"GET":    http.HandlerFunc(pa.ShowPixel),
					"PUT":    http.HandlerFunc(pa.UpdatePixel),
					"DELETE": http.HandlerFunc(pa.DeletePixel),
				},
			}),
		},
	})
	pa.Handler = h
	return pa
}

func (pa *PixelApi) ListPixels(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	enc.Encode(pa.pixels)
}

func (pa *PixelApi) CreatePixel(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id := GenerateAlnumString(3)

	buf := &bytes.Buffer{}
	io.Copy(buf, r.Body)
	if buf.Len() <= 0 {
		http.Error(w, "Empty fs", http.StatusBadRequest)
		return
	}
	ctr, err := pa.cc.CreateContainer(tar.NewReader(bytes.NewReader(buf.Bytes())), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pixel := &Pixel{
		Id:         id,
		Container:  ctr,
		Filesystem: fsObject(tar.NewReader(bytes.NewReader(buf.Bytes()))),
		LastImage:  &bytes.Buffer{},
		Clicks:     make(chan *pixelutils.Click),
	}
	io.Copy(pixel.LastImage, bytes.NewReader(blackPixel.Bytes()))

	pa.Lock()
	pa.pixels[id] = pixel
	pa.Unlock()

	pa.Messages <- &Message{
		Pixel: id,
		Type:  TypeCreate,
	}

	go pa.pixelListener(pixel)

	http.Error(w, id, http.StatusCreated)
}

func (pa *PixelApi) UpdatePixel(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	pixel := w.(httptools.VarsResponseWriter).Vars()["pixel"].(*Pixel)

	buf := &bytes.Buffer{}
	io.Copy(buf, r.Body)
	if buf.Len() <= 0 {
		http.Error(w, "Empty fs", http.StatusBadRequest)
		return
	}

	ctr, err := pa.cc.CreateContainer(tar.NewReader(bytes.NewReader(buf.Bytes())), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	StopContainer(pixel.Container)
	pixel.Container = ctr
	pixel.Filesystem = fsObject(tar.NewReader(bytes.NewReader(buf.Bytes())))
	pixel.LastImage = &bytes.Buffer{}
	pixel.Broken = false
	io.Copy(pixel.LastImage, bytes.NewReader(blackPixel.Bytes()))

	pa.Messages <- &Message{
		Pixel: pixel.Id,
		Type:  TypeChange,
	}

	go pa.pixelListener(pixel)

	http.Error(w, pixel.Id, http.StatusCreated)
}

func (pa *PixelApi) DeletePixel(w http.ResponseWriter, r *http.Request) {
	pixel := w.(httptools.VarsResponseWriter).Vars()["pixel"].(*Pixel)
	token, err := r.Cookie("deleteToken")
	if options.DeleteToken == "" || err != nil || token.Value != options.DeleteToken {
		http.Error(w, pixel.Id, http.StatusForbidden)
		return
	}

	StopContainer(pixel.Container)
	pa.Messages <- &Message{
		Pixel: pixel.Id,
		Type:  TypeDelete,
	}
	http.Error(w, pixel.Id, http.StatusOK)
}

func (pa *PixelApi) pixelListener(pixel *Pixel) {
	id := pixel.Id
	// FIXME: Arbitrary wait to accomodate startup
	time.Sleep(1 * time.Second)
	addr := pixel.Address()
	c, err := net.Dial("tcp", addr.String())
	if err != nil {
		pa.Messages <- pixel.Fail("Could not connect to pixel %s: %s", id, err)
		return
	}
	defer c.Close()

	go func() {
		pixel.Wait()
		pa.Messages <- pixel.Fail("Pixel %s terminated", id)
	}()

	go func() {
		enc := json.NewEncoder(c)
		for click := range pixel.Clicks {
			err := enc.Encode(click)
			if err != nil {
				return
			}
		}
	}()

	tr := tar.NewReader(c)
	for {
		_, err := tr.Next()
		if err != nil {
			log.Printf("Pixel %s closed its reader: %s", id, err)
			pa.Messages <- pixel.Fail(err.Error())
			return
		}
		pixel.LastImage.Reset()
		io.Copy(pixel.LastImage, tr)
		pa.Messages <- &Message{
			Pixel: id,
			Type:  TypeChange,
		}
	}
}

func (pa *PixelApi) ValidatePixelId(w http.ResponseWriter, r *http.Request) {
	vars := w.(httptools.VarsResponseWriter).Vars()
	id, ok := vars["1"].(string)
	if !ok {
		http.Error(w, "Pixel ID missing", http.StatusBadRequest)
		return
	}

	pa.RLock()
	pixel, ok := pa.pixels[id]
	pa.RUnlock()
	if !ok {
		http.Error(w, "No such pixel", http.StatusBadRequest)
		return
	}
	vars["pixel"] = pixel
}

func (pa *PixelApi) ShowPixel(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	enc.Encode(w.(httptools.VarsResponseWriter).Vars()["pixel"].(*Pixel))
}

func (pa *PixelApi) GetPixelLogs(w http.ResponseWriter, r *http.Request) {
	pixel := w.(httptools.VarsResponseWriter).Vars()["pixel"].(*Pixel)

	io.WriteString(w, pixel.Logs())
}

func (pa *PixelApi) GetPixelFs(w http.ResponseWriter, r *http.Request) {
	pixel := w.(httptools.VarsResponseWriter).Vars()["pixel"].(*Pixel)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pixel.Filesystem)
}

func (pa *PixelApi) GetPixelContent(w http.ResponseWriter, r *http.Request) {
	pixel := w.(httptools.VarsResponseWriter).Vars()["pixel"].(*Pixel)

	w.Header().Set("cache-control", "private, max-age=0, no-cache")
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", pixel.LastImage.Len()))
	io.Copy(w, bytes.NewReader(pixel.LastImage.Bytes()))
}

func (pa *PixelApi) ReportClick(c *pixelutils.Click) {
	pixel, ok := pa.pixels[c.PixelId]
	if !ok {
		log.Printf("Received unknown Pixel ID %s", c.PixelId)
		return
	}
	pixel.Clicks <- c
}

func fsObject(r *tar.Reader) map[string]string {
	fs := map[string]string{}
	for {
		hdr, err := r.Next()
		if err != nil {
			break
		}
		if !strings.HasSuffix(hdr.Name, ".go") {
			continue
		}
		buf := &bytes.Buffer{}
		io.Copy(buf, r)
		fs[hdr.Name] = buf.String()
	}
	return fs
}
