package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/imdraw"
	"fmt"
	"image"
	_ "image/png"
	"math"
	"os"
	"time"
	"image/color"
	"golang.org/x/image/colornames"
)
const(
	WINDOW_HEIGHT = 800
	WINDOW_WIDTH = 800
	PACMAN_SPEED = .5
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

type Direction int

const (
	up Direction = iota
	down
	left
	right
)
type block struct{
	frame pixel.Rect
	pos pixel.Rect
	sheet pixel.Picture
}
func (blk *block) draw(t pixel.Target){
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
type board struct{
	sheet pixel.Picture
	blocks []*block
}
func (brd *board) load(worldMap [][]uint8, sheet pixel.Picture) error{
	var err error
	brd.sheet = sheet
	if err != nil {
		panic(err)
	}
	brd.blocks = make([]*block, 0)
	frame := getFrame(24, 24, 5, 0 )
	for i:=0;i<len(worldMap);i++{
		for j:=0;j<len(worldMap[0]);j++{
			if worldMap[i][j]==0{
				brd.blocks = append(brd.blocks, &block{frame:frame, pos:getRectInGrid(WINDOW_WIDTH,WINDOW_HEIGHT, len(worldMap[0]), len(worldMap),j, i),sheet:brd.sheet} )	
			}else{
				fmt.Println(getRectInGrid(WINDOW_WIDTH,WINDOW_HEIGHT, len(worldMap[0]), len(worldMap),j, i))
			}
		}
	}
	
	// brd.blocks = append(brd.blocks, &block{frame:frame, pos:getRectInGrid(WINDOW_WIDTH,WINDOW_HEIGHT,20,20,2,3),sheet:brd.sheet} )
	// brd.blocks = append(brd.blocks, &block{frame:frame, pos:getRectInGrid(WINDOW_WIDTH,WINDOW_HEIGHT,20,20,2,4),sheet:brd.sheet} )
	// brd.blocks = append(brd.blocks, &block{frame:frame, pos:getRectInGrid(WINDOW_WIDTH,WINDOW_HEIGHT,20,20,2,5),sheet:brd.sheet} )
	// brd.blocks = append(brd.blocks, &block{frame:frame, pos:getRectInGrid(WINDOW_WIDTH,WINDOW_HEIGHT,20,20,3,5),sheet:brd.sheet} )

	return nil
}
func (brd *board) check_collision(pos pixel.Rect) bool{
	for _, block := range brd.blocks{
		if(isColliding(block.pos, pos)){
			return true
		}
	}
	return false
}
func (brd *board) draw(t pixel.Target) error{
	for _, block := range brd.blocks{
		block.draw(t)
	}
	return nil
}
type platform struct {
	rect  pixel.Rect
	color color.Color
}
type pacman struct {
	direction Direction
	anims     map[Direction][]pixel.Rect
	rate      float64
	counter   float64
	frame     pixel.Rect    //stores current frame. updates in update function
	sheet     pixel.Picture //stores spritesheel in pixel picture format
	pos       pixel.Rect
}
type world struct{
	pm *pacman
	brd *board
} 
var World = &world{}
func (pm *pacman) load(sheet pixel.Picture) error {
	var err error
	pm.sheet = sheet
	if err != nil {
		panic(err)
	}
	pm.rate = 1 / 5.0
	pm.pos = pixel.R(560, 680, 600, 720)
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
	// sprite.Draw(t, pixel.IM)
	// sprite := pixel.NewSprite(pm.sheet, pm.sheet.Bounds())
	// sprite.Draw(t, pixel.IM)
}
func isColliding(rect1 pixel.Rect, rect2 pixel.Rect) bool{
	if rect1.Min.X < rect2.Max.X && rect1.Max.X > rect2.Min.X && rect1.Min.Y < rect2.Max.Y && rect1.Max.Y > rect2.Min.Y {
		return true
	}
	return false
}
func (pm *pacman) update(dt float64, directionValue Direction) {
	pm.counter = dt //why dt is based on time
	pm.direction = directionValue
	directionVecMap := make(map[Direction]pixel.Vec)
	directionVecMap[right] = pixel.V(10*PACMAN_SPEED, 0)
	directionVecMap[left] = pixel.V(-10*PACMAN_SPEED, 0)
	directionVecMap[up] = pixel.V(0, 10*PACMAN_SPEED)
	directionVecMap[down] = pixel.V(0, -10*PACMAN_SPEED)
	pos_new := pm.pos.Moved(directionVecMap[directionValue])
	if !World.brd.check_collision(pos_new){
		pm.pos = pos_new
	}

	i := int(math.Floor(pm.counter / pm.rate))
	pm.frame = pm.anims[pm.direction][i%len(pm.anims[pm.direction])]
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
	sheet, err:=getSheet("spritemap-384.png")
	if err!=nil{
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
	World.pm = pm
	World.brd = brd
	last := time.Now()

	var direction Direction
	for !win.Closed() {
		// pm.draw(win)
		// imd.Draw(canvas)
		// time.Sleep(100 * time.Millisecond)
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

		pm.update(dt, direction)
		pm.draw(imd)
		brd.draw(imd)
		imd.Draw(win)
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

func getRectInGrid(width float64, height float64, totalx int, totaly int, x int, y int) pixel.Rect{
	fmt.Println(x, y)
	gridWidth := width/float64(totalx)
	gridHeight := height/float64(totaly)
	return pixel.R(float64(x)*gridWidth,float64(y)*gridHeight, float64((x+1))*gridWidth, float64((y+1))*gridHeight)
}
func main() {
	pixelgl.Run(run)
}
