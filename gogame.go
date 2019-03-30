package main

import (
	"fmt"
	"image"
	_ "image/png"
	"math"
	"os"
	"golang.org/x/image/colornames"
	"time"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

const (
	WINDOW_HEIGHT = 800
	WINDOW_WIDTH  = 800
)

type Direction int

const (
	up Direction = iota
	down
	left
	right
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

//Get a rect coordinates after dividing the picture
func getFrame(frameWidth float64, frameHeight float64, xGrid int, yGrid int) pixel.Rect {
	return pixel.R(
		float64(xGrid)*frameWidth,
		float64(yGrid)*frameHeight,
		float64(xGrid+1)*frameWidth,
		float64(yGrid+1)*frameHeight,
	)
}

//Get a rect coordinates by index after dividing the picture to girds
func getRectInGrid(width float64, height float64, totalx int, totaly int, x int, y int) pixel.Rect {
	gridWidth := width / float64(totalx)
	gridHeight := height / float64(totaly)
	return pixel.R(float64(x)*gridWidth, float64(y)*gridHeight, float64((x+1))*gridWidth, float64((y+1))*gridHeight)
}

type pacman struct {
	direction Direction
	anims     map[Direction][]pixel.Rect
	rate      float64
	frame     pixel.Rect    //stores current frame. updates in update function
	sheet     pixel.Picture //stores spritesheel in pixel picture format
	pos       pixel.Rect
	gridX     int
	gridY     int
}

func (pm *pacman) load(sheet pixel.Picture) error {
	var err error
	pm.sheet = sheet
	if err != nil {
		panic(err)
	}
	pm.pos = getRectInGrid(WINDOW_WIDTH, WINDOW_HEIGHT, 20, 20, pm.gridX, pm.gridY)
	pm.anims = make(map[Direction][]pixel.Rect)
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
}
func (pm *pacman) getNewGridPos(direction Direction) (int, int) {
	if direction == right {
		return pm.gridX + 1, pm.gridY
	}
	if direction == left {
		return pm.gridX - 1, pm.gridY
	}
	if direction == up {
		return pm.gridX, pm.gridY + 1
	}
	if direction == down {
		return pm.gridX, pm.gridY - 1
	}
	return pm.gridX, pm.gridY
}
func (pm *pacman) update(dt float64, direction Direction) {
	pm.direction = direction
	pm.gridX, pm.gridY = pm.getNewGridPos(direction)
	i := int(math.Floor(dt / pm.rate))
	pm.pos = getRectInGrid(WINDOW_WIDTH, WINDOW_HEIGHT, 20,20, pm.gridX, pm.gridY)
	pm.frame = pm.anims[pm.direction][i%len(pm.anims[pm.direction])]
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Pacman",
		Bounds: pixel.R(0, 0, WINDOW_WIDTH, WINDOW_HEIGHT),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	sheet, err := getSheet("spritemap-384.png")
	pm := &pacman{gridX:1,gridY:1,rate:1/5.0}
	//load game objects
	err = pm.load(sheet)
	if err != nil {
		panic(err)
	}
	last := time.Now()
	for !win.Closed() {
		win.Clear(colornames.Black)
		//update game objects
		dt := time.Since(last).Seconds()
		pm.update(dt, right)
		pm.draw(win)
		//draw game objects
		win.Update()
		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	pixelgl.Run(run)
}
