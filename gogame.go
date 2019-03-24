package main

import (
	"fmt"
	"image"
	_ "image/png"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"strconv"

	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

const (
	WINDOW_HEIGHT = 800
	WINDOW_WIDTH  = 800
	PACMAN_SPEED  = 2
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
func getRectInGrid(width float64, height float64, totalx int, totaly int, x int, y int) pixel.Rect {
	gridWidth := width / float64(totalx)
	gridHeight := height / float64(totaly)
	return pixel.R(float64(x)*gridWidth, float64(y)*gridHeight, float64((x+1))*gridWidth, float64((y+1))*gridHeight)
}

type Direction int

const (
	up Direction = iota
	down
	left
	right
)

type block struct {
	frame pixel.Rect
	pos   pixel.Rect
	sheet pixel.Picture
}
func (blk block) draw(t pixel.Target) {
	sprite := pixel.NewSprite(nil, pixel.Rect{})
	sprite.Set(blk.sheet, blk.frame)
	sprite.Draw(t, pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(
			blk.pos.W()/sprite.Frame().W(),
			blk.pos.H()/sprite.Frame().H(),
		)).
		Moved(blk.pos.Center()),
	)
}

type coin struct {
	frame pixel.Rect
	pos   pixel.Rect
	sheet pixel.Picture
}
func (cn coin) draw(t pixel.Target) {
	sprite := pixel.NewSprite(nil, pixel.Rect{})
	sprite.Set(cn.sheet, cn.frame)
	sprite.Draw(t, pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(
			cn.pos.W()/sprite.Frame().W(),
			cn.pos.H()/sprite.Frame().H(),
		)).
		Moved(cn.pos.Center()),
	)
}

type board struct {
	sheet pixel.Picture
}
func (brd *board) load(worldMap [][]uint8, sheet pixel.Picture) error {
	var err error
	brd.sheet = sheet
	if err != nil {
		panic(err)
	}
	return nil
}
func (brd *board) draw(t pixel.Target) error {
	blkFrame := getFrame(24, 24, 0, 2)
	coinFrame := getFrame(12, 12, 16, 19)
	worldMap := World.worldMap
	for i := 0; i < len(worldMap); i++ {
		for j := 0; j < len(worldMap[0]); j++ {
			if worldMap[i][j] == 0 {
				block{frame: blkFrame, pos: getRectInGrid(WINDOW_WIDTH, WINDOW_HEIGHT, len(worldMap[0]), len(worldMap), j, i), sheet: brd.sheet}.draw(t)
			} else if worldMap[i][j] == 1 {
				coin{frame: coinFrame, pos: getRectInGrid(WINDOW_WIDTH, WINDOW_HEIGHT, len(worldMap[0]), len(worldMap), j, i), sheet: brd.sheet}.draw(t)
			}
		}
	}
	return nil
}

type ghost struct {
	direction Direction
	anims     map[Direction][]pixel.Rect
	rate      float64
	counter   float64
	frame     pixel.Rect    //stores current frame. updates in update function
	sheet     pixel.Picture //stores spritesheel in pixel picture format
	pos       pixel.Rect
	gridX     int
	gridY     int
}
func (gh *ghost) load(sheet pixel.Picture) error {
	var err error
	gh.sheet = sheet
	if err != nil {
		panic(err)
	}
	gh.rate = 1 / 5.0
	gh.gridX = 2
	gh.gridY = 1
	gh.pos = getRectInGrid(WINDOW_WIDTH, WINDOW_HEIGHT, len(World.worldMap[0]), len(World.worldMap), gh.gridX, gh.gridY)
	gh.direction = right
	gh.anims = make(map[Direction][]pixel.Rect)
	gh.frame = getFrame(24, 24, 1, 6)
	gh.anims[up] = append(gh.anims[up], getFrame(24, 24, 6, 0))
	gh.anims[up] = append(gh.anims[up], getFrame(24, 24, 7, 0))
	gh.anims[down] = append(gh.anims[down], getFrame(24, 24, 2, 0))
	gh.anims[down] = append(gh.anims[down], getFrame(24, 24, 3, 0))
	gh.anims[left] = append(gh.anims[left], getFrame(24, 24, 4, 0))
	gh.anims[left] = append(gh.anims[left], getFrame(24, 24, 5, 0))
	gh.anims[right] = append(gh.anims[right], getFrame(24, 24, 0, 0))
	gh.anims[right] = append(gh.anims[right], getFrame(24, 24, 1, 0))
	return nil
}
func (gh *ghost) draw(t pixel.Target) {
	sprite := pixel.NewSprite(nil, pixel.Rect{})
	sprite.Set(gh.sheet, gh.frame)
	sprite.Draw(t, pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(
			gh.pos.W()/sprite.Frame().W(),
			gh.pos.H()/sprite.Frame().H(),
		)).
		Moved(gh.pos.Center()),
	)
}
func (gh *ghost) update(dt float64) {
	gh.counter = dt //why dt is based on time
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
	if gh.gridX < 0 || gh.gridX >= len(World.worldMap[0]) || gh.gridY < 0 || gh.gridY > len(World.worldMap) || World.worldMap[gh.gridY][gh.gridX] == 0 {
		gh.gridX = old_gridx
		gh.gridY = old_gridy
		possible := make([]Direction, 0)
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
		gh.direction = possible[rand.Intn(len(possible))]

	} else {
		gh.pos = getRectInGrid(WINDOW_WIDTH, WINDOW_HEIGHT, len(World.worldMap[0]), len(World.worldMap), gh.gridX, gh.gridY)
	}
	i := int(math.Floor(gh.counter / gh.rate))
	gh.frame = gh.anims[gh.direction][i%len(gh.anims[gh.direction])]
}

type pacman struct {
	direction Direction
	anims     map[Direction][]pixel.Rect
	rate      float64
	counter   float64
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
	pm.rate = 1 / 5.0
	pm.gridX = 1
	pm.gridY = 1
	pm.pos = getRectInGrid(WINDOW_WIDTH, WINDOW_HEIGHT, len(World.worldMap[0]), len(World.worldMap), pm.gridX, pm.gridY)

	// pm.pos = pixel.R(560, 680, 600, 720)
	// pm.pos = getRectInGrid()
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
func (pm *pacman) getNewGridPos(direction Direction)(int, int){
	if direction == right {
		return pm.gridX+1, pm.gridY
	}
	if direction == left {
		return pm.gridX-1, pm.gridY
	}
	if direction == up {
		return pm.gridX, pm.gridY+1
	}
	if direction == down {
		return pm.gridX, pm.gridY-1
	}
	return pm.gridX, pm.gridY
}
func (pm *pacman) update(dt float64, direction Direction) {
	pm.counter = dt //why dt is based on time
	// pm.direction = direction
	newGridX, newGridY := pm.getNewGridPos(direction)
	if newGridX < 0 || newGridX >= len(World.worldMap[0]) || newGridY < 0 || newGridY > len(World.worldMap) || World.worldMap[newGridY][newGridX] == 0 {
		newGridX, newGridY = pm.getNewGridPos(pm.direction)
		if newGridX < 0 || newGridX >= len(World.worldMap[0]) || newGridY < 0 || newGridY > len(World.worldMap) || World.worldMap[newGridY][newGridX] == 0 {
			// newGridX, newGridY = pm.getNewGridPos(pm.direction)
		}else{
			pm.gridX, pm.gridY = newGridX, newGridY
			pm.pos = getRectInGrid(WINDOW_WIDTH, WINDOW_HEIGHT, len(World.worldMap[0]), len(World.worldMap), pm.gridX, pm.gridY)
		}
	} else {
		pm.direction = direction
		pm.gridX, pm.gridY = newGridX, newGridY
		pm.pos = getRectInGrid(WINDOW_WIDTH, WINDOW_HEIGHT, len(World.worldMap[0]), len(World.worldMap), pm.gridX, pm.gridY)
	}
	for _, ghost := range World.ghosts {
		if pm.gridX == ghost.gridX && pm.gridY == ghost.gridY {
			World.gameOver = true
		}
	}
	if World.worldMap[pm.gridY][pm.gridX] == 1 {
		World.worldMap[pm.gridY][pm.gridX] = 2
		fmt.Println(World.score)
		World.score++
	}
	i := int(math.Floor(pm.counter / pm.rate))
	pm.frame = pm.anims[pm.direction][i%len(pm.anims[pm.direction])]
}

type world struct {
	pm       *pacman
	brd      *board
	ghosts   []*ghost
	worldMap [][]uint8
	score    int
	gameOver bool
}
var World = &world{}


func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Packman",
		Bounds: pixel.R(0, 0, WINDOW_WIDTH, WINDOW_HEIGHT),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	// canvas := pixelgl.NewCanvas(pixel.R(-160/2, -120/2, 160/2, 120/2))

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
	if err != nil {
		panic(err)
	}
	imd := imdraw.New(sheet)
	brd := &board{}
	err = brd.load(worldMap, sheet)
	if err != nil {
		panic(err)
	}
	pm := &pacman{}
	err = pm.load(sheet)
	if err != nil {
		panic(err)
	}
	gh1 := &ghost{}
	err = gh1.load(sheet)
	if err != nil {
		panic(err)
	}
	gh2 := &ghost{}
	err = gh2.load(sheet)
	if err != nil {
		panic(err)
	}
	gh3 := &ghost{}
	err = gh3.load(sheet)
	if err != nil {
		panic(err)
	}
	gh4 := &ghost{}
	err = gh4.load(sheet)
	if err != nil {
		panic(err)
	}
	World.pm = pm
	World.brd = brd
	World.worldMap = worldMap
	World.ghosts = []*ghost{gh1, gh2, gh3, gh4}
	last := time.Now()

	var direction Direction
	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	for !win.Closed() {
		if World.gameOver == true {
			basicTxt := text.New(pixel.V(100, 500), basicAtlas)
			fmt.Fprintln(basicTxt, "Game Over!!")
			fmt.Fprintln(basicTxt, "Score:"+strconv.Itoa(World.score))
			win.Clear(colornames.Black)
			basicTxt.Draw(win, pixel.IM.Scaled(basicTxt.Orig, 4))
			win.Update()
			time.Sleep(1000 * time.Millisecond)
			break
		}
		time.Sleep(100 * time.Millisecond)
		dt := time.Since(last).Seconds()

		win.Clear(colornames.Black)
		imd.Clear()
		
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
		pm.update(dt, direction)
		pm.draw(imd)
		gh1.update(dt)
		gh1.draw(imd)
		gh2.update(dt)
		gh2.draw(imd)
		gh3.update(dt)
		gh3.draw(imd)
		gh4.update(dt)
		gh4.draw(imd)

		imd.Draw(win)

		basicTxt := text.New(pixel.V(500, 750), basicAtlas)
		fmt.Fprintln(basicTxt, "Score:"+strconv.Itoa(World.score))
		basicTxt.Draw(win, pixel.IM.Scaled(basicTxt.Orig, 3))

		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
