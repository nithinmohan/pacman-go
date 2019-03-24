package main

import (
	_ "image/png"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"

	"golang.org/x/image/colornames"
)

const (
	WINDOW_HEIGHT = 800
	WINDOW_WIDTH  = 800
)


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
}

func main() {
	pixelgl.Run(run)
}
