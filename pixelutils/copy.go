package pixelutils

import (
	"github.com/Zwobot/go-resample/resample"
	"image"
	"image/draw"
)

func Resize(dst draw.Image, src image.Image) {
	resizeImg, _ := resample.Resize(image.Point{dst.Bounds().Dx(), dst.Bounds().Dy()}, src)
	draw.Draw(dst, dst.Bounds(), resizeImg, src.Bounds().Canon().Min, draw.Over)
}

func SubPixel(pixel Pixel, r image.Rectangle) Pixel {
	return pixel.SubImage(r).(Pixel)
}
