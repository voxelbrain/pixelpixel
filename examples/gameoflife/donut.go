package main

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

func DonutCoords(x, y int, r image.Rectangle) (int, int) {
	r = r.Canon()
	x = (int(math.Mod(float64(x), float64(r.Dx())))+r.Dx())%r.Dx() + r.Min.X
	y = (int(math.Mod(float64(y), float64(r.Dy())))+r.Dy())%r.Dy() + r.Min.Y
	return x, y
}

type DonutImage struct {
	draw.Image
}

func (di *DonutImage) Set(x, y int, c color.Color) {
	x, y = DonutCoords(x, y, di.Bounds())
	di.Image.Set(x, y, c)
}

func (di *DonutImage) At(x, y int) color.Color {
	x, y = DonutCoords(x, y, di.Bounds())
	return di.Image.At(x, y)
}

type GameBoard interface {
	Dimensions() (int, int)
	Get(x, y int) bool
	Set(x, y int, alive int)
}

type ImageBoard struct {
	img draw.Image
}

func (ib *ImageBoard) Dimensions() (int, int) {
	return ib.img.Bounds().Dx(), ib.img.Bounds().Dy()
}

func (ib *ImageBoard) Get(x, y int) bool {
	return ib.img.At(x, y) == color.Black
}

func (ib *ImageBoard) Set(x, y int, alive bool) {
	c := color.White
	if alive {
		c = color.Black
	}
	ib.img.Set(x, y, c)
}
