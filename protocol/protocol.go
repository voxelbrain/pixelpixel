package protocol

import (
	"archive/tar"
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"net"
	"os"
)

type PixelHandler func(p *Pixel)

type Pixel struct {
	draw.Image
	commit chan bool
	done   chan bool
	rwc    io.ReadWriteCloser
}

func (p *Pixel) Commit() {
	p.commit <- true
	<-p.done
}

func ServePixel(h PixelHandler) {
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
	defer c.Close()
	p := &Pixel{
		Image:  image.NewRGBA(image.Rect(0, 0, 256, 256)),
		commit: make(chan bool),
		done:   make(chan bool),
		rwc:    c,
	}
	go commitLoop(p)
	h(p)
	log.Printf("Handler has returned")
}

func commitLoop(p *Pixel) {
	buf := &bytes.Buffer{}
	tw := tar.NewWriter(p.rwc)
	for _ = range p.commit {
		buf.Reset()
		png.Encode(buf, p)

		tw.WriteHeader(&tar.Header{
			Size: int64(buf.Len()),
		})
		io.Copy(tw, buf)
		tw.Flush()
		p.done <- true
	}
}
