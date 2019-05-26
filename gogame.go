package main

import (
	"fmt"
	"image"
	_ "image/png"
	"math"
	"os"
	"golang.org/x/image/colornames"
	"github.com/faiface/pixel/imdraw"
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

type block struct {
	frame pixel.Rect
	sheet pixel.Picture
	gridX     int //position in grid
	gridY     int //position in grid
}

type board struct {
	sheet pixel.Picture
}
func (brd *board) load(sheet pixel.Picture) error {
	var err error
	brd.sheet = sheet
	if err != nil {
		panic(err)
	}
	return nil
}
func (brd *board) draw(t pixel.Target) error {
	blkFrame := getFrame(24, 24, 0, 5)
	coinFrame := getFrame(12, 12, 16, 19)
	worldMap := World.worldMap
	for i := 0; i < len(worldMap); i++ {
		for j := 0; j < len(worldMap[0]); j++ {
			if worldMap[i][j] == 0 {
				b:=block{frame: blkFrame, gridX:i, gridY:j, sheet:brd.sheet}
				b.draw(t)
			}else if worldMap[i][j] == 1 {
				coin{frame: coinFrame, gridX:i, gridY:j, sheet:brd.sheet}.draw(t)
			}
		}
	}
	return nil
}

func (blk block) draw(t pixel.Target) {
	sprite := pixel.NewSprite(nil, pixel.Rect{})
	sprite.Set(blk.sheet, blk.frame)
	pos := getRectInGrid(WINDOW_WIDTH, WINDOW_HEIGHT, len(World.worldMap[0]), len(World.worldMap), blk.gridY, blk.gridX)
	sprite.Draw(t, pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(
			pos.W()/sprite.Frame().W(),
			pos.H()/sprite.Frame().H(),
		)).
		Moved(pos.Center()),
	)
}
type coin struct {
		frame pixel.Rect
		gridX int
		gridY int
		sheet pixel.Picture
}
func (cn coin) draw(t pixel.Target) {
		sprite := pixel.NewSprite(nil, pixel.Rect{})
		sprite.Set(cn.sheet, cn.frame)
		pos := getRectInGrid(WINDOW_WIDTH, WINDOW_HEIGHT, len(World.worldMap[0]), len(World.worldMap), cn.gridY, cn.gridX)
		sprite.Draw(t, pixel.IM.
				ScaledXY(pixel.ZV, pixel.V(
						pos.W()/sprite.Frame().W(),
						pos.H()/sprite.Frame().H(),
				)).
				Moved(pos.Center()),
		)
}
type pacman struct {
	direction Direction
	anims     map[Direction][]pixel.Rect
	rate      float64
	frame     pixel.Rect    //stores current frame. updates in update function
	sheet     pixel.Picture //stores spritesheel in pixel picture format
	gridX     int
	gridY     int
}

func (pm *pacman) load(sheet pixel.Picture) error {
	var err error
	pm.sheet = sheet
	if err != nil {
		panic(err)
	}
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
	pos := getRectInGrid(WINDOW_WIDTH, WINDOW_HEIGHT, 20, 20, pm.gridX, pm.gridY)
	sprite.Draw(t, pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(
			pos.W()/sprite.Frame().W(),
			pos.H()/sprite.Frame().H(),
		)).
		Moved(pos.Center()),
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
	pm.frame = pm.anims[pm.direction][i%len(pm.anims[pm.direction])]
}

type world struct {
	pm       *pacman
	brd      *board
	worldMap [][]uint8
	score    int
	gameOver bool
}
type loadable interface{
	load(pixel.Picture) error
}

var World = &world{}

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

	worldMap := [][]uint8{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 1, 1, 1, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0},
		{0, 1, 0, 0, 1, 0, 0, 1, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 0, 0},
		{0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0},
		{0, 1, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 1, 0, 0, 0, 0},
		{0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 0, 1, 1, 1, 1, 0, 0, 0, 0},
		{0, 1, 0, 0, 1, 0, 0, 1, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0},
		{0, 1, 0, 0, 1, 0, 1, 1, 1, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0},
		{0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 1, 0, 1, 1, 1, 1, 1, 0, 1, 0, 0, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 1, 0, 0, 0, 0},
		{0, 1, 1, 1, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0},
		{0, 1, 0, 0, 1, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0},
		{0, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 1, 1, 0, 0, 0, 0},
		{0, 0, 1, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 0, 1, 0, 0, 0, 0, 0},
		{0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 0, 1, 1, 1, 1, 0, 0, 0, 0},
		{0, 1, 0, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0},
		{0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	World.worldMap = worldMap

	sheet, err := getSheet("spritemap-384.png")
	pm := &pacman{gridX:1,gridY:1,rate:1/5.0}
	//load game objects
	err = pm.load(sheet)
	if err != nil {
		panic(err)
	}
	imd := imdraw.New(sheet)
	brd := &board{}

	objectsToLoad := []loadable{brd, pm}

	for _, object:=range(objectsToLoad){
		err = object.load(sheet)
		if err != nil {
			panic(err)
		}
	}
	World.pm = pm
	World.brd = brd
	World.worldMap = worldMap

	direction:=right
	last := time.Now()
	for !win.Closed() {
		win.Clear(colornames.Black)
		imd.Clear()
		//update game objects
		if win.Pressed(pixelgl.KeyLeft) {
			direction = left
		}
		if win.Pressed(pixelgl.KeyRight) {
			direction = right
		}
		if win.Pressed(pixelgl.KeyUp) {
			direction = up
		}
		if win.Pressed(pixelgl.KeyDown) {
			direction = down
		}

		brd.draw(imd)

		dt := time.Since(last).Seconds()
		pm.update(dt, direction)
		pm.draw(win)
		imd.Draw(win)
		//draw game objects
		win.Update()
		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	pixelgl.Run(run)
}
