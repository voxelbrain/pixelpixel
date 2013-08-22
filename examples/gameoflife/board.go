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
	Image       draw.Image
	Dead, Alive color.Color
}

func (ib *ImageBoard) Dimensions() (int, int) {
	return ib.Image.Bounds().Dx(), ib.Image.Bounds().Dy()
}

func (ib *ImageBoard) Get(x, y int) bool {
	if colorEqual(ib.Image.At(x, y), ib.Alive) {
		return true
	}
	return false
}

func (ib *ImageBoard) Set(x, y int, alive bool) {
	c := ib.Dead
	if alive {
		c = ib.Alive
	}
	ib.Image.Set(x, y, c)
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

func colorEqual(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}
