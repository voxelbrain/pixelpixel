package main

import (
	"math/rand"
	"time"

	"github.com/voxelbrain/pixelpixel/imageutils"
	"github.com/voxelbrain/pixelpixel/protocol"
)

const (
	Size = 64
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	protocol.ServePixel(func(p *protocol.Pixel) {
		board := &ImageBoard{&DonutImage{imageutils.DimensionChanger(p, Size, Size)}}
		initBoard(board)
		for {
			p.Commit()
			NextGen(board)
			time.Sleep(300 * time.Millisecond)
		}
	})
}

func initBoard(b GameBoard) {
	for y := 0; y < Size; y++ {
		for x := 0; x < Size; x++ {
			alive := false
			if rand.Float64() < 0.5 {
				alive = true
			}
			b.Set(x, y, alive)
		}
	}
}
