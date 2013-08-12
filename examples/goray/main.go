package main

import (
	"github.com/voxelbrain/pixelpixel/pixelutils"
	"log"
	"os"

	"bitbucket.org/zombiezen/goray"
	"bitbucket.org/zombiezen/goray/intersect"
	raylog "bitbucket.org/zombiezen/goray/log"
	"bitbucket.org/zombiezen/goray/yamlscene"

	// Anonymous imports to register with yaml parser
	_ "bitbucket.org/zombiezen/goray/cameras"
	_ "bitbucket.org/zombiezen/goray/integrators"
	_ "bitbucket.org/zombiezen/goray/lights"
	_ "bitbucket.org/zombiezen/goray/materials"
	_ "bitbucket.org/zombiezen/goray/shaders/texmap"
	_ "bitbucket.org/zombiezen/goray/textures"
)

const (
	sceneFile = `scene.yml`
)

func main() {
	c := pixelutils.PixelPusher()
	pixel := pixelutils.NewPixel()

	sc, integ := buildScene(sceneFile)
	img := goray.Render(sc, integ, raylog.Default)

	pixelutils.StretchCopy(pixel, img)
	c <- pixel
	select {}
}

func buildScene(file string) (*goray.Scene, goray.Integrator) {
	sc := goray.NewScene(goray.IntersecterBuilder(intersect.NewKD), raylog.Default)

	f, err := os.Open("scene.yml")
	if err != nil {
		log.Fatalf("Could not read scene: %s", err)
	}
	defer f.Close()
	integ, err := yamlscene.Load(f, sc, yamlscene.Params{})
	if err != nil {
		log.Fatalf("Could not parse scene: %s", err)
	}
	return sc, integ
}
