package main

import (
	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

const (
	WINDOW_HEIGHT = 800
	WINDOW_WIDTH  = 800
)

type world struct {
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

	//load game objects

	for !win.Closed() {
		//update game objects
		//draw game objects
	}
}

func main() {
	pixelgl.Run(run)
}
