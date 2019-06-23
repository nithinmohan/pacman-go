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
	"math/rand"
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

type ghost struct {
	direction Direction //current direction of object
	anims     map[Direction][]pixel.Rect //stores direction to frames list map
	rate      float64 //animation rate
	frame     pixel.Rect    //stores current frame. updates in update function
	sheet     pixel.Picture //stores spritesheel in pixel picture format
	gridX     int //position in grid
	gridY     int //position in grid
	spriteRow int
	spriteCol int
}
func (gh *ghost) load(sheet pixel.Picture) error {
	var err error
	gh.sheet = sheet
	if err != nil {
		panic(err)
	}
	gh.direction = right
	gh.anims = make(map[Direction][]pixel.Rect)
	gh.frame = getFrame(24, 24, 1, 6)
	gh.anims[up] = append(gh.anims[up], getFrame(24, 24, gh.spriteCol+6, gh.spriteRow))
	gh.anims[up] = append(gh.anims[up], getFrame(24, 24, gh.spriteCol+7, gh.spriteRow))
	gh.anims[down] = append(gh.anims[down], getFrame(24, 24, gh.spriteCol+2, gh.spriteRow))
	gh.anims[down] = append(gh.anims[down], getFrame(24, 24, gh.spriteCol+3, gh.spriteRow))
	gh.anims[left] = append(gh.anims[left], getFrame(24, 24, gh.spriteCol+4, gh.spriteRow))
	gh.anims[left] = append(gh.anims[left], getFrame(24, 24, gh.spriteCol+5, gh.spriteRow))
	gh.anims[right] = append(gh.anims[right], getFrame(24, 24, gh.spriteCol+0, gh.spriteRow))
	gh.anims[right] = append(gh.anims[right], getFrame(24, 24, gh.spriteCol+1, gh.spriteRow))
	return nil
}
func (gh *ghost) draw(t pixel.Target) {
	sprite := pixel.NewSprite(nil, pixel.Rect{})
	sprite.Set(gh.sheet, gh.frame)
	pos := getRectInGrid(WINDOW_WIDTH, WINDOW_HEIGHT, len(World.worldMap[0]), len(World.worldMap), gh.gridX, gh.gridY)
	sprite.Draw(t, pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(
			pos.W()/sprite.Frame().W(),
			pos.H()/sprite.Frame().H(),
		)).
		Moved(pos.Center()),
	)
}
func (gh *ghost) update(dt float64) {
	directionValue := gh.direction
	old_gridx := gh.gridX
	old_gridy := gh.gridY
	if directionValue == right {
		gh.gridX += 1
	}
	if directionValue == left {
		gh.gridX -= 1
	}
	if directionValue == up {
		gh.gridY += 1
	}
	if directionValue == down {
		gh.gridY -= 1
	}
	//cbeck for collision with block
	if gh.gridX < 0 || gh.gridX >= len(World.worldMap[0]) || gh.gridY < 0 || gh.gridY > len(World.worldMap) || World.worldMap[gh.gridY][gh.gridX] == 0 {
		gh.gridX = old_gridx
		gh.gridY = old_gridy
		possible := make([]Direction, 0)
		//find list of possible direction where ghost can move
		if World.worldMap[gh.gridY+1][gh.gridX] != 0 {
			possible = append(possible, up)
		}
		if World.worldMap[gh.gridY-1][gh.gridX] != 0 {
			possible = append(possible, down)
		}
		if World.worldMap[gh.gridY][gh.gridX+1] != 0 {
			possible = append(possible, right)
		}
		if World.worldMap[gh.gridY][gh.gridX-1] != 0 {
			possible = append(possible, left)
		}
		//select one direction out of it
		gh.direction = possible[rand.Intn(len(possible))]

	}
	i := int(math.Floor(dt / gh.rate))
	gh.frame = gh.anims[gh.direction][i%len(gh.anims[gh.direction])]
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
	//Map of all frames to be showed per direction of pacman
  	//we will use this for the animation
	pm.anims = make(map[Direction][]pixel.Rect)
	//here 1,6 is the cordinates of image to be shown, after dividing the sprite sheet to 24x24
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
func (pm *pacman) isCollidingWithWall(gridX int, gridY int) bool{
	return gridX < 0 || gridX >= len(World.worldMap[0]) || gridY < 0 || gridY > len(World.worldMap) || World.worldMap[gridY][gridX] == 0
}

func (pm *pacman) update(dt float64, direction Direction) {
	//get the new position according to the direction passed
	newGridX, newGridY := pm.getNewGridPos(direction)
    if !pm.isCollidingWithWall(newGridX, newGridY) {
		//If newly calculated position is not colliding with block
		//update the direction and position
		pm.direction = direction
		pm.gridX, pm.gridY = newGridX, newGridY
        
	} else {
		//if the position is colliding, try with existing direction
		newGridX, newGridY = pm.getNewGridPos(pm.direction)
        if !pm.isCollidingWithWall(newGridX, newGridY) {
            pm.gridX, pm.gridY = newGridX, newGridY
		}
	}
	i := int(math.Floor(dt / pm.rate))
	pm.frame = pm.anims[pm.direction][i%len(pm.anims[pm.direction])]
}

type world struct {
	pm       *pacman
	brd      *board
	ghosts   []*ghost
	worldMap [][]uint8
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
	gh1 := &ghost{gridX:5,gridY:10,rate:1/5.0,spriteRow:0, spriteCol:0}
	gh2 := &ghost{gridX:15,gridY:14,rate:1/5.0,spriteRow:1, spriteCol:0}
	gh3 := &ghost{gridX:8,gridY:3,rate:1/5.0,spriteRow:3, spriteCol:0}
	gh4 := &ghost{gridX:2,gridY:9,rate:1/5.0,spriteRow:1, spriteCol:8}
	//load game objects
	err = pm.load(sheet)
	if err != nil {
		panic(err)
	}
	imd := imdraw.New(sheet)
	brd := &board{}

	objectsToLoad := []loadable{brd, pm, gh1, gh2, gh3, gh4}

	for _, object:=range(objectsToLoad){
		err = object.load(sheet)
		if err != nil {
			panic(err)
		}
	}
	World.pm = pm
	World.brd = brd
	World.ghosts = []*ghost{gh1, gh2, gh3, gh4}
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
		for _,gh :=range(World.ghosts){
			gh.update(dt)
			gh.draw(imd)
		}
		imd.Draw(win)
		//draw game objects
		win.Update()
		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	pixelgl.Run(run) 
}
