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
	wall, _ := pixelutils.PixelPusher()
	pixel := pixelutils.NewPixel()

	board := &ImageBoard{
		Image: &DonutImage{pixelutils.DimensionChanger(pixel, Size, Size)},
		Alive: pixelutils.Green,
		Dead:  pixelutils.Black,
	}
	initBoard(board)
	for {
		wall <- pixel
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
