package main

import (
	"time"

	"github.com/voxelbrain/pixelpixel/protocol"
)

func main() {
	time.AfterFunc(4*time.Second, func() {
		panic("CRASH")
	})
	protocol.ServePixel(func(p *protocol.Pixel) {
		p.Commit()
		// Block indefinitely
		select {}
	})
}
