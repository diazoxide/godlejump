// This is a temporary file to generate the initial assets
// It won't be used in the final game
package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

func main() {
	// Create player sprite
	playerImg := image.NewRGBA(image.Rect(0, 0, 40, 40))
	
	// Fill background with transparency
	for y := 0; y < 40; y++ {
		for x := 0; x < 40; x++ {
			playerImg.Set(x, y, color.RGBA{0, 0, 0, 0})
		}
	}
	
	// Draw character body (green circle)
	for y := 5; y < 35; y++ {
		for x := 5; x < 35; x++ {
			dx := float64(x - 20)
			dy := float64(y - 20)
			if dx*dx + dy*dy < 15*15 {
				playerImg.Set(x, y, color.RGBA{50, 200, 50, 255})
			}
		}
	}
	
	// Draw eyes
	for y := 12; y < 16; y++ {
		for x := 12; x < 16; x++ {
			playerImg.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	for y := 12; y < 16; y++ {
		for x := 24; x < 28; x++ {
			playerImg.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	
	// Draw pupils
	for y := 13; y < 15; y++ {
		for x := 13; x < 15; x++ {
			playerImg.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	for y := 13; y < 15; y++ {
		for x := 25; x < 27; x++ {
			playerImg.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	
	// Draw mouth
	for y := 25; y < 27; y++ {
		for x := 15; x < 25; x++ {
			playerImg.Set(x, y, color.RGBA{255, 100, 100, 255})
		}
	}
	
	// Save player image
	playerFile, _ := os.Create("player.png")
	defer playerFile.Close()
	png.Encode(playerFile, playerImg)
	
	// Create platform sprite
	platformImg := image.NewRGBA(image.Rect(0, 0, 60, 10))
	
	// Fill with light blue
	for y := 0; y < 10; y++ {
		for x := 0; x < 60; x++ {
			platformImg.Set(x, y, color.RGBA{100, 200, 255, 255})
		}
	}
	
	// Add some details
	for y := 2; y < 8; y++ {
		for x := 5; x < 55; x += 10 {
			platformImg.Set(x, y, color.RGBA{50, 150, 200, 255})
		}
	}
	
	// Save platform image
	platformFile, _ := os.Create("platform.png")
	defer platformFile.Close()
	png.Encode(platformFile, platformImg)
	
	// Create bird sprites (left and right facing)
	birdImg := image.NewRGBA(image.Rect(0, 0, 40, 30))
	
	// Fill background with transparency
	for y := 0; y < 30; y++ {
		for x := 0; x < 40; x++ {
			birdImg.Set(x, y, color.RGBA{0, 0, 0, 0})
		}
	}
	
	// Draw bird body
	for y := 10; y < 25; y++ {
		for x := 5; x < 35; x++ {
			birdImg.Set(x, y, color.RGBA{200, 100, 50, 255})
		}
	}
	
	// Draw wings
	for y := 5; y < 15; y++ {
		for x := 0; x < 15; x++ {
			birdImg.Set(x, y, color.RGBA{200, 150, 50, 255})
		}
	}
	for y := 5; y < 15; y++ {
		for x := 25; x < 40; x++ {
			birdImg.Set(x, y, color.RGBA{200, 150, 50, 255})
		}
	}
	
	// Draw eyes
	for y := 12; y < 16; y++ {
		for x := 8; x < 12; x++ {
			birdImg.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	for y := 13; y < 15; y++ {
		for x := 9; x < 11; x++ {
			birdImg.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	
	// Draw beak
	for y := 17; y < 20; y++ {
		for x := 0; x < 5; x++ {
			birdImg.Set(x, y, color.RGBA{255, 200, 0, 255})
		}
	}
	
	// Save bird left image
	birdLeftFile, _ := os.Create("bird_left.png")
	defer birdLeftFile.Close()
	png.Encode(birdLeftFile, birdImg)
	
	// Create right facing bird by flipping the left one
	birdRightImg := image.NewRGBA(image.Rect(0, 0, 40, 30))
	for y := 0; y < 30; y++ {
		for x := 0; x < 40; x++ {
			birdRightImg.Set(x, y, birdImg.At(39-x, y))
		}
	}
	
	// Save bird right image
	birdRightFile, _ := os.Create("bird_right.png")
	defer birdRightFile.Close()
	png.Encode(birdRightFile, birdRightImg)
	
	// Create cloud sprite
	cloudImg := image.NewRGBA(image.Rect(0, 0, 80, 40))
	
	// Fill background with transparency
	for y := 0; y < 40; y++ {
		for x := 0; x < 80; x++ {
			cloudImg.Set(x, y, color.RGBA{0, 0, 0, 0})
		}
	}
	
	// Draw cloud (multiple overlapping circles)
	centers := []struct{ x, y, r int }{
		{20, 20, 15},
		{35, 15, 12},
		{50, 18, 14},
		{60, 20, 10},
	}
	
	for y := 0; y < 40; y++ {
		for x := 0; x < 80; x++ {
			// Check if point is inside any of the circles
			for _, c := range centers {
				dx := float64(x - c.x)
				dy := float64(y - c.y)
				dist := math.Sqrt(dx*dx + dy*dy)
				
				if dist <= float64(c.r) {
					// White with slight transparency
					cloudImg.Set(x, y, color.RGBA{255, 255, 255, 230})
					break
				}
			}
		}
	}
	
	// Save cloud image
	cloudFile, _ := os.Create("cloud.png")
	defer cloudFile.Close()
	png.Encode(cloudFile, cloudImg)
}