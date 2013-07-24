package main

import (
	"math/rand"
	"time"

	"github.com/voxelbrain/pixelpixel/pixelutils"
)

const (
	Size = 64
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	c := pixelutils.PixelPusher()
	img := pixelutils.NewPixel()

	board := &ImageBoard{&DonutImage{pixelutils.DimensionChanger(img, Size, Size)}}
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
