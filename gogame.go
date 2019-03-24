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
	COIN_COUNT = 160
)

//Get Picture object from a img file
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
				block{frame: blkFrame, pos: getRectInGrid(WINDOW_WIDTH, WINDOW_HEIGHT, len(worldMap[0]), len(worldMap), j, i), sheet: brd.sheet}.draw(t)
			} else if worldMap[i][j] == 1 {
				coin{frame: coinFrame, pos: getRectInGrid(WINDOW_WIDTH, WINDOW_HEIGHT, len(worldMap[0]), len(worldMap), j, i), sheet: brd.sheet}.draw(t)
			}
		}
	}
	return nil
}

type ghost struct {
	direction Direction //current direction of object
	anims     map[Direction][]pixel.Rect //stores direction to frames list map
	rate      float64 //animation rate
	frame     pixel.Rect    //stores current frame. updates in update function
	sheet     pixel.Picture //stores spritesheel in pixel picture format
	pos       pixel.Rect //stores position to draw the sprite
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
	gh.pos = getRectInGrid(WINDOW_WIDTH, WINDOW_HEIGHT, len(World.worldMap[0]), len(World.worldMap), gh.gridX, gh.gridY)
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
	sprite.Draw(t, pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(
			gh.pos.W()/sprite.Frame().W(),
			gh.pos.H()/sprite.Frame().H(),
		)).
		Moved(gh.pos.Center()),
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
	if World.pm.gridX == gh.gridX && World.pm.gridY == gh.gridY {
		World.gameOver = true
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
	i := int(math.Floor(dt / pm.rate))
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

type loadable interface{
	load(pixel.Picture) error
}

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
	pm := &pacman{gridX:1,gridY:1,rate:1/5.0}
	gh1 := &ghost{gridX:5,gridY:10,rate:1/5.0,spriteRow:0, spriteCol:0}
	gh2 := &ghost{gridX:15,gridY:14,rate:1/5.0,spriteRow:1, spriteCol:0}
	gh3 := &ghost{gridX:8,gridY:3,rate:1/5.0,spriteRow:3, spriteCol:0}
	gh4 := &ghost{gridX:2,gridY:9,rate:1/5.0,spriteRow:1, spriteCol:8}
	objectsToLoad := []loadable{brd, pm, gh1, gh2, gh3, gh4}

	for _, object:=range(objectsToLoad){
		err = object.load(sheet)
		if err != nil {
			panic(err)
		}
	}
	World.pm = pm
	World.brd = brd
	World.worldMap = worldMap
	World.ghosts = []*ghost{gh1, gh2, gh3, gh4}
	last := time.Now()

	direction:=right
	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	for !win.Closed() {
		if World.gameOver == true || World.score == COIN_COUNT{
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
		for _,gh :=range(World.ghosts){
			gh.update(dt)
			gh.draw(imd)
		}

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
