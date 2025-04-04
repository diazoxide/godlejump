// This is a file to generate the assets
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
)

// Create mountain layer images
func createMountainLayer(width, height int, baseColor color.RGBA, roughness float64) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with transparency initially
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 0})
		}
	}

	// Generate mountain silhouette using midpoint displacement
	points := make([]float64, width)
	
	// Start with a more natural mountain shape
	points[0] = float64(height) * 0.8  // Higher start point
	points[width-1] = float64(height) * 0.8

	// Recursive midpoint displacement
	var subdivide func(start, end int, displacement float64)
	subdivide = func(start, end int, displacement float64) {
		if end-start < 2 {
			return
		}

		mid := (start + end) / 2
		points[mid] = (points[start] + points[end]) / 2
		points[mid] += (rand.Float64()*2 - 1) * displacement

		// Ensure the mountain stays within bounds but allow for more height variation
		if points[mid] < float64(height)*0.3 {
			points[mid] = float64(height) * 0.3
		}
		if points[mid] > float64(height)*0.9 {
			points[mid] = float64(height) * 0.9
		}

		subdivide(start, mid, displacement*roughness)
		subdivide(mid, end, displacement*roughness)
	}

	subdivide(0, width-1, float64(height)*0.4)

	// Smooth out the points
	smoothedPoints := make([]float64, width)
	copy(smoothedPoints, points)
	for i := 1; i < width-1; i++ {
		smoothedPoints[i] = (points[i-1] + points[i]*2 + points[i+1]) / 4
	}
	points = smoothedPoints

	// Fill the mountain with gradient and proper alpha blending
	for x := 0; x < width; x++ {
		mountainHeight := int(points[x])
		for y := 0; y < height; y++ {
			if y >= mountainHeight {
				// Calculate gradient based on height
				progress := float64(y-mountainHeight) / float64(height-mountainHeight)
				
				// Add some noise to the color
				noise := (rand.Float64()*0.1 - 0.05)
				
				// Calculate alpha for smooth blending
				alpha := uint8(255 * (1.0 - math.Pow(progress, 0.5)))
				
				// Adjust color based on height
				r := uint8(float64(baseColor.R) * (1.0 - progress*0.3 + noise))
				g := uint8(float64(baseColor.G) * (1.0 - progress*0.3 + noise))
				b := uint8(float64(baseColor.B) * (1.0 - progress*0.3 + noise))
				
				img.Set(x, y, color.RGBA{r, g, b, alpha})
			}
		}
	}

	return img
}

func main() {
	// Create player sprite (flying character)
	playerImg := image.NewRGBA(image.Rect(0, 0, 40, 40))
	
	// Fill background with transparency
	for y := 0; y < 40; y++ {
		for x := 0; x < 40; x++ {
			playerImg.Set(x, y, color.RGBA{0, 0, 0, 0})
		}
	}
	
	// Draw bird-like body (blue)
	for y := 10; y < 30; y++ {
		for x := 10; x < 30; x++ {
			dx := float64(x - 20)
			dy := float64(y - 20)
			if dx*dx + dy*dy < 10*10 {
				playerImg.Set(x, y, color.RGBA{50, 100, 220, 255})
			}
		}
	}
	
	// Draw wings
	for y := 15; y < 25; y++ {
		for x := 2; x < 15; x++ {
			// Wing shape - curved
			dx := float64(x - 8)
			dy := float64(y - 20)
			if dx*dx/36 + dy*dy/25 < 1 {
				playerImg.Set(x, y, color.RGBA{100, 150, 240, 255})
			}
		}
	}
	
	// Right wing
	for y := 15; y < 25; y++ {
		for x := 25; x < 38; x++ {
			dx := float64(x - 32)
			dy := float64(y - 20)
			if dx*dx/36 + dy*dy/25 < 1 {
				playerImg.Set(x, y, color.RGBA{100, 150, 240, 255})
			}
		}
	}
	
	// Draw eyes
	for y := 14; y < 18; y++ {
		for x := 16; x < 19; x++ {
			playerImg.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	for y := 14; y < 18; y++ {
		for x := 22; x < 25; x++ {
			playerImg.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	
	// Draw pupils
	for y := 15; y < 17; y++ {
		for x := 17; x < 18; x++ {
			playerImg.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	for y := 15; y < 17; y++ {
		for x := 23; x < 24; x++ {
			playerImg.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	
	// Draw beak
	for y := 17; y < 22; y++ {
		for x := 30; x < 35; x++ {
			dx := float64(x - 32)
			dy := float64(y - 19)
			
			if dx*dx/25 + dy*dy/12 < 1 {
				playerImg.Set(x, y, color.RGBA{255, 200, 0, 255})
			}
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

	// Create mountain layers with different colors
	mountainColors := []color.RGBA{
		{160, 170, 180, 255}, // Back mountains (lighter gray)
		{130, 140, 160, 255}, // Middle mountains (medium gray-blue)
		{100, 110, 140, 255}, // Front mountains (darker blue-gray)
	}

	// Increased dimensions for better quality
	mountainWidth := 1200   // Wider for smoother tiling
	mountainHeight := 800   // Taller for better coverage

	for i, baseColor := range mountainColors {
		// Create mountain layer with adjusted parameters
		mountainImg := createMountainLayer(
			mountainWidth,
			mountainHeight,
			baseColor,
			0.8-float64(i)*0.2,  // Varying roughness for each layer
		)
		
		fileName := fmt.Sprintf("mountains_%d.png", i)
		mountainFile, _ := os.Create(fileName)
		png.Encode(mountainFile, mountainImg)
		mountainFile.Close()
	}
}