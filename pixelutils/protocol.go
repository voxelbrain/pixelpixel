package pixelutils

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

type Pixel interface {
	draw.Image
	SubImage(r image.Rectangle) image.Image
}

func NewPixel() Pixel {
	return image.NewRGBA(image.Rect(0, 0, 256, 256))
}

func PixelPusher() chan<- Pixel {
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

	return commitLoop(c)
}

func commitLoop(w io.WriteCloser) chan<- Pixel {
	c := make(chan Pixel)
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
