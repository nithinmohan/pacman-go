package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"image"
	"os"
	"fmt"
)

func getImage(filePath string) (image.Image, error) {
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
	return img, nil
}

func getFrame(img image.Image, frameWidth float64, frameHeight float64, xGrid int, yGrid int) (pixel.Picture, pixel.Rect){
	sheet := pixel.PictureDataFromImage(img)
	return sheet, pixel.R(
		float64(xGrid)*frameWidth,
		float64(xGrid+1)*frameWidth,
		float64(yGrid)*frameHeight,
		float64(yGrid+1)*frameHeight,
	)
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
	for !win.Closed() {
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
