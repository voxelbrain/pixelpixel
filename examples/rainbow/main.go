package main

import (
	"archive/tar"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"net"
	"os"
)

var (
	red = color.NRGBA{255, 255, 255, 0}
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
	draw.Draw(img, image.Rect(0, 0, 256, 256), &image.Uniform{red}, image.Point{0, 0}, draw.Over)
	tw := tar.NewWriter(c)
	tw.WriteHeader(&tar.Header{})
	png.Encode(tw, img)
	tw.Flush()
	tw.WriteHeader(&tar.Header{})
	select {}
}
