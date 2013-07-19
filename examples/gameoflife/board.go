package main

import (
	"image/color"
	"image/draw"
)

type GameBoard interface {
	Dimensions() (int, int)
	Get(x, y int) bool
	Set(x, y int, alive bool)
}

type ImageBoard struct {
	img draw.Image
}

func (ib *ImageBoard) Dimensions() (int, int) {
	return ib.img.Bounds().Dx(), ib.img.Bounds().Dy()
}

func (ib *ImageBoard) Get(x, y int) bool {
	r, _, _, _ := ib.img.At(x, y).RGBA()
	if r == 0 {
		return true
	}
	return false
}

func (ib *ImageBoard) Set(x, y int, alive bool) {
	c := color.White
	if alive {
		c = color.Black
	}
	ib.img.Set(x, y, c)
}

func NextGen(board GameBoard) {
	w, h := board.Dimensions()
	newBoard := make([]bool, w*h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			aliveNeighbors := countNeighbors(board, x, y)
			newBoard[y*w+x] = false
			if board.Get(x, y) && (aliveNeighbors == 2 || aliveNeighbors == 3) {
				newBoard[y*w+x] = true
			}
			if !board.Get(x, y) && aliveNeighbors == 3 {
				newBoard[y*w+x] = true
			}
		}
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			board.Set(x, y, newBoard[y*w+x])
		}
	}
}

func countNeighbors(board GameBoard, x, y int) int {
	return toInt(board.Get(x-1, y-1)) +
		toInt(board.Get(x, y-1)) +
		toInt(board.Get(x+1, y-1)) +
		toInt(board.Get(x-1, y)) +
		toInt(board.Get(x+1, y)) +
		toInt(board.Get(x-1, y+1)) +
		toInt(board.Get(x, y+1)) +
		toInt(board.Get(x+1, y+1))

}

func toInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
