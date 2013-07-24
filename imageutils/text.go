package imageutils

import (
	"image"
	"image/color"
	"image/draw"
)

func DrawText(img draw.Image, r image.Rectangle, c color.Color, text string) {
	r = r.Canon()
	colorImg := &image.Uniform{c}
	charRect := image.Rectangle{
		Min: r.Min,
		Max: r.Min.Add(image.Point{4, 6}),
	}
	for _, char := range []byte(text) {
		if char == '\n' {
			nextLine(&charRect, r.Min.X)
			continue
		}
		mask := MaskForCharacter(char)
		draw.DrawMask(img, charRect, colorImg, image.Point{0, 0}, mask, image.Point{0, 0}, draw.Over)
		charRect = charRect.Add(image.Point{4, 0})
		if !charRect.In(r) {
			nextLine(&charRect, r.Min.X)
		}
	}
}

func nextLine(r *image.Rectangle, xstart int) {
	r.Min.X = xstart
	r.Min.Y += 6
	r.Max = r.Min.Add(image.Point{4, 6})
}

func MaskForCharacter(c byte) image.Image {
	mask := image.NewAlpha(image.Rect(0, 0, 4, 6))
	rows := font[int(c)]
	for y, row := range rows {
		for x := uint(0); x < 4; x++ {
			if (row>>x)&1 > 0 {
				mask.Set(4-(int(x)+1), y, color.Black)
			}
		}
	}
	return mask
}
