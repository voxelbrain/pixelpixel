package pixelutils

import (
	"image"
	"image/color"
	"image/draw"
)

type dimensionChanger struct {
	draw.Image
	pixel              image.Rectangle
	paddingX, paddingY int
	bounds             image.Rectangle
}

// PixelSizeChanger enlarges pixels to the given size, effectively
// making the image smaller. If either of the pixel dimensions is not
// a multiple of the respective image dimension, a padding of unusable
// pixels will remain to the right and bottom.
func PixelSizeChanger(img draw.Image, w, h int) draw.Image {
	b := img.Bounds().Canon()
	return &dimensionChanger{
		Image:    img,
		paddingX: b.Dx() % w,
		paddingY: b.Dy() % h,
		pixel:    image.Rect(0, 0, w, h),
		bounds:   image.Rect(0, 0, b.Dx()/w, b.Dy()/h),
	}
}

// Like PixelSizeChanger, but instead of the pixel size the new image
// is given.
func DimensionChanger(img draw.Image, w, h int) draw.Image {
	b := img.Bounds().Canon()
	return &dimensionChanger{
		Image:    img,
		paddingX: b.Dx() % w,
		paddingY: b.Dy() % h,
		pixel:    image.Rect(0, 0, b.Dx()/w, b.Dy()/h),
		bounds:   image.Rect(0, 0, w, h),
	}
}

func (d *dimensionChanger) Set(x, y int, c color.Color) {
	p := image.Point{x, y}
	if !p.In(d.bounds) {
		return
	}
	draw.Draw(d.Image, d.pixel.Add(image.Point{x * d.pixel.Dx(), y * d.pixel.Dy()}).Add(d.Image.Bounds().Canon().Min), &image.Uniform{c}, image.Point{0, 0}, draw.Over)
}

func (d *dimensionChanger) At(x, y int) color.Color {
	p := image.Point{x, y}
	if !p.In(d.bounds) {
		return nil
	}
	return d.Image.At(x*d.pixel.Dx()+d.Image.Bounds().Canon().Min.X, y*d.pixel.Dy()+d.Image.Bounds().Canon().Min.Y)
}

func (d *dimensionChanger) Bounds() image.Rectangle {
	return d.bounds
}

func (d *dimensionChanger) SubImage(r image.Rectangle) image.Image {
	var newD = &dimensionChanger{}
	*newD = *d
	newD.bounds = r
	return newD
}
