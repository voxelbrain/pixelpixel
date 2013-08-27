package pixelutils

import (
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	"image/draw"
)

// StretchCopy transforms src to fill dst and draws the result over
// dst.
func StretchCopy(dst draw.Image, src image.Image) {
	resizeImg := imaging.Resize(src, dst.Bounds().Dx(), dst.Bounds().Dy(), imaging.Lanczos)
	draw.Draw(dst, dst.Bounds(), resizeImg, resizeImg.Bounds().Min, draw.Over)
}

// SubImage extracts a rectangular subset of img as a separate image.
// If present, the `SubImage()` method will be used. Otherwise, the
// behaviour will be emulated.
func SubImage(img draw.Image, r image.Rectangle) draw.Image {
	if di, ok := img.(subimager); ok {
		si := di.SubImage(r)
		dsi, ok := si.(draw.Image)
		if !ok {
			panic("Image is drawable, subimage is not.")
		}
		return dsi
	}
	return &subimage{
		Image:  img,
		bounds: img.Bounds().Intersect(r),
	}
}

type subimager interface {
	SubImage(r image.Rectangle) image.Image
}

type subimage struct {
	draw.Image
	bounds image.Rectangle
}

func (si *subimage) Bounds() image.Rectangle {
	return si.bounds
}

func (si *subimage) At(x, y int) color.Color {
	p := image.Point{x, y}
	if !p.In(si.Bounds()) {
		return color.Black
	}
	return si.Image.At(x, y)
}

func (si *subimage) Set(x, y int, c color.Color) {
	p := image.Point{x, y}
	if !p.In(si.Bounds()) {
		return
	}
	si.Image.Set(x, y, c)
}

func (si *subimage) SubImage(r image.Rectangle) image.Image {
	nsi := &subimage{}
	*nsi = *si
	nsi.bounds = si.Bounds().Intersect(r)
	return nsi
}
