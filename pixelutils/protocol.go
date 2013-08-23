package pixelutils

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"net"
	"os"
)

type Click struct {
	PixelId  string `json:"key"`
	Position struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"position"`
}

func (c *Click) Point() image.Point {
	return image.Point{
		X: c.Position.X,
		Y: c.Position.Y,
	}
}

func NewPixel() draw.Image {
	return image.NewRGBA(image.Rect(0, 0, 256, 256))
}

func PixelPusher() (chan<- draw.Image, <-chan *Click) {
	addr := fmt.Sprintf("localhost:%s", os.Getenv("PORT"))
	log.Printf("Starting pixel on %s", addr)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Could not open socket on %s: %s", addr, err)
	}

	c, err := l.Accept()
	if err != nil {
		log.Fatalf("Could not accept connection: %s", err)
	}

	return commitLoop(c), clickDecoder(c)
}

func commitLoop(w io.WriteCloser) chan<- draw.Image {
	c := make(chan draw.Image)
	go func() {
		buf := &bytes.Buffer{}
		tw := tar.NewWriter(w)
		for img := range c {
			buf.Reset()
			png.Encode(buf, img)

			tw.WriteHeader(&tar.Header{
				Size: int64(buf.Len()),
			})
			io.Copy(tw, buf)
			tw.Flush()
		}
	}()
	return c
}

func clickDecoder(r io.ReadCloser) <-chan *Click {
	c := make(chan *Click)
	go func() {
		dec := json.NewDecoder(r)
		for {
			click := &Click{}
			err := dec.Decode(&click)
			if err != nil {
				log.Printf("Received invalid click object: %s", err)
				continue
			}
			// Discard click if there's no one to read it.
			// Don't build up backpressure.
			select {
			case c <- click:
			default:
			}
		}
	}()
	return c
}
