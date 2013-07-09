package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log"
	"net"
	"os"
	"time"
)

var (
	red  = color.NRGBA{255, 0, 0, 255}
	blue = color.NRGBA{0, 0, 255, 255}
)

func main() {
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

	img := image.NewRGBA(image.Rect(0, 0, 256, 256))
	fillColor := blue
	buf := &bytes.Buffer{}
	tw := tar.NewWriter(c)
	for {
		buf.Reset()
		if fillColor == blue {
			fillColor = red
		} else {
			fillColor = blue
		}
		draw.Draw(img, image.Rect(0, 0, 256, 256), &image.Uniform{fillColor}, image.Point{0, 0}, draw.Over)
		png.Encode(buf, img)

		tw.WriteHeader(&tar.Header{
			Size: int64(buf.Len()),
		})
		io.Copy(tw, buf)
		tw.Flush()
		time.Sleep(1000 * time.Millisecond)
	}
}
