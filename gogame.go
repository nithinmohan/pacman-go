package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	// "github.com/faiface/pixel/imdraw"
	"fmt"
	"image"
	_ "image/png"
	"math"
	"os"
	"time"

	"golang.org/x/image/colornames"
)

func getSheet(filePath string) (pixel.Picture, error) {
	imgFile, err := os.Open(filePath)
	defer imgFile.Close()
	if err != nil {
		fmt.Println("Cannot read file:", err)
		return nil, err
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		fmt.Println("Cannot decode file:", err)
		return nil, err
	}
	sheet := pixel.PictureDataFromImage(img)
	return sheet, nil
}

func getFrame(frameWidth float64, frameHeight float64, xGrid int, yGrid int) pixel.Rect {
	//https://github.com/faiface/pixel/wiki/Drawing-a-Sprite#picture

	return pixel.R(
		float64(xGrid)*frameWidth,
		float64(yGrid)*frameHeight,
		float64(xGrid+1)*frameWidth,
		float64(yGrid+1)*frameHeight,
	)
}

type direction int

const (
	up direction = iota
	down
	left
	right
)

type pacman struct {
	direction direction
	anims     map[direction][]pixel.Rect
	rate      float64
	counter   float64
	frame     pixel.Rect    //stores current frame. updates in update function
	sheet     pixel.Picture //stores spritesheel in pixel picture format
	pos       pixel.Rect
}

func (pm *pacman) load() error {
	var err error
	pm.sheet, err = getSheet("spritemap-384.png")
	if err != nil {
		panic(err)
	}
	pm.rate = 1/10.0
	pm.pos = pixel.R(10, 10, 100, 100)
	pm.anims = make(map[direction][]pixel.Rect)
	pm.frame = getFrame(24, 24, 1, 6)
	pm.anims[up] = append(pm.anims[up], getFrame(24, 24, 1, 6))
	pm.anims[up] = append(pm.anims[up], getFrame(24, 24, 3, 6))
	pm.anims[down] = append(pm.anims[down], getFrame(24, 24, 5, 6))
	pm.anims[down] = append(pm.anims[down], getFrame(24, 24, 7, 6))
	pm.anims[left] = append(pm.anims[left], getFrame(24, 24, 0, 6))
	pm.anims[left] = append(pm.anims[left], getFrame(24, 24, 2, 6))
	pm.anims[right] = append(pm.anims[right], getFrame(24, 24, 4, 6))
	pm.anims[right] = append(pm.anims[right], getFrame(24, 24, 6, 6))
	return nil

}
func (pm *pacman) draw(t pixel.Target) {
	sprite := pixel.NewSprite(nil, pixel.Rect{})
	sprite.Set(pm.sheet, pm.frame)
	sprite.Draw(t, pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(
			pm.pos.W()/sprite.Frame().W(),
			pm.pos.H()/sprite.Frame().H(),
		)).
		Moved(pm.pos.Center()),
	)
	// sprite.Draw(t, pixel.IM)
	// sprite := pixel.NewSprite(pm.sheet, pm.sheet.Bounds())
	// sprite.Draw(t, pixel.IM)
}
func (pm *pacman) update(dt float64, directionValue direction) {
	pm.counter = dt //why dt is based on time
	pm.direction = directionValue
	directionVecMap := make(map[direction]pixel.Vec)
	directionVecMap[right] = pixel.V(1,0)
	directionVecMap[left] = pixel.V(-1,0)
	directionVecMap[up] = pixel.V(0,1)
	directionVecMap[down] = pixel.V(0,-1) 
	pm.pos = pm.pos.Moved(directionVecMap[directionValue])
	i := int(math.Floor(pm.counter / pm.rate))
	fmt.Println(i)
	pm.frame = pm.anims[pm.direction][i%len(pm.anims[pm.direction])]
}
func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Packman",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	// canvas := pixelgl.NewCanvas(pixel.R(-160/2, -120/2, 160/2, 120/2))
	// imd := imdraw.New(nil)

	pm := &pacman{}
	err = pm.load()
	if err != nil {
		panic(err)
	}
	last := time.Now()
	for !win.Closed() {
		// pm.draw(win)
		// imd.Draw(canvas)
		dt := time.Since(last).Seconds()
		win.Clear(colornames.Black)
		pm.update(dt, right)
		pm.draw(win)
		// win.SetMatrix(pixel.IM.Scaled(pixel.ZV,
		// 	math.Min(
		// 		win.Bounds().W()/canvas.Bounds().W(),
		// 		win.Bounds().H()/canvas.Bounds().H(),
		// 	),
		// ).Moved(win.Bounds().Center()))
		// canvas.Draw(win, pixel.IM.Moved(canvas.Bounds().Center()))
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
