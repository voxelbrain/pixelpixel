package pixelutils

import (
	"github.com/Zwobot/go-resample/resample"
	"image"
	"image/draw"
)

func Copy(dst draw.Image, src image.Image, sr, dr image.Rectangle) {
	subRect := sr.Sub(sr.Min)
	subImg := image.NewRGBA(sr)
	draw.Draw(subImg, subRect, src, sr.Min, draw.Over)
	resizeImg, _ := resample.Resize(image.Point{dr.Dx(), dr.Dy()}, subImg)
	draw.Draw(dst, dr, resizeImg, image.Point{0, 0}, draw.Over)
}

type SubImager interface {
	draw.Image
	SubImage(r image.Rectangle) image.Image
}

func SubImage(img SubImager, r image.Rectangle) draw.Image {
	return img.SubImage(r).(draw.Image)
}
