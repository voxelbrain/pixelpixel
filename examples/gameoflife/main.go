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
	c := protocol.PixelPusher()
	img := protocol.NewPixel()

	board := &ImageBoard{&DonutImage{imageutils.DimensionChanger(img, Size, Size)}}
	initBoard(board)
	for {
		c <- img
		NextGen(board)
		time.Sleep(300 * time.Millisecond)
	}
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
