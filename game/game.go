package game

import (
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	ScreenWidth    = 320
	ScreenHeight   = 480
	PlatformWidth  = 60
	PlatformHeight = 10
	PlayerWidth    = 40
	PlayerHeight   = 40
	BirdWidth      = 40
	BirdHeight     = 30
	CloudWidth     = 80
	CloudHeight    = 40
	Gravity        = 0.2
	JumpVelocity   = -8
	PlatformCount  = 10
	BirdCount      = 3
	CloudCount     = 5
	SnowflakeCount = 40
	RaindropCount  = 50
	BirdSpeedMin   = 1
	BirdSpeedMax   = 3
	CloudSpeedMin  = 0.2
	CloudSpeedMax  = 1.0
)

// Weather types
const (
	WeatherClear = iota
	WeatherRain
	WeatherSnow
)

// We don't actually need the embed since the files are loaded directly
// This is a workaround as embed paths can't be relative to parent directories

// Platform represents a platform in the game
type Platform struct {
	X, Y float64
}

// Bird represents a bird obstacle
type Bird struct {
	X, Y      float64
	SpeedX    float64
	Direction int // 1 for right, -1 for left
}

// Cloud represents a background cloud
type Cloud struct {
	X, Y   float64
	SpeedX float64
	Width  float64
	Height float64
	Alpha  float64 // transparency
}

// Weather particle (rain or snow)
type Particle struct {
	X, Y   float64
	SpeedX float64
	SpeedY float64
	Size   float64
	Alpha  float64
}

// Player represents the player character
type Player struct {
	X, Y        float64
	VelocityY   float64
	FacingRight bool
}

// Game implements ebiten.Game interface
type Game struct {
	player       Player
	platforms    []Platform
	birds        []Bird
	clouds       []Cloud
	particles    []Particle
	camera       float64
	score        int
	playerImg    *ebiten.Image
	platformImg  *ebiten.Image
	birdLeftImg  *ebiten.Image
	birdRightImg *ebiten.Image
	cloudImg     *ebiten.Image
	gameOver     bool
	nightMode    bool
	weather      int
	startTime    time.Time
	cycleTime    time.Duration
	weatherTimer float64 // counter for weather changes
}

// loadImage loads an image from file path
func loadImage(path string) *ebiten.Image {
	imgFile, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open image: %v", err)
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		log.Fatalf("Failed to decode image: %v", err)
	}

	return ebiten.NewImageFromImage(img)
}

// NewGame creates a new game instance
func NewGame() *Game {
	// We don't need to seed in newer Go versions

	g := &Game{
		player: Player{
			X:           ScreenWidth / 2,
			Y:           ScreenHeight - 100,
			FacingRight: true,
		},
		platforms:    make([]Platform, PlatformCount),
		birds:        make([]Bird, BirdCount),
		clouds:       make([]Cloud, CloudCount),
		particles:    make([]Particle, 0, RaindropCount),
		gameOver:     false,
		startTime:    time.Now(),
		cycleTime:    time.Minute * 2,     // Day/night cycle every 2 minutes
		weatherTimer: rand.Float64() * 15, // Random time until weather changes
		weather:      WeatherClear,
	}

	// Load images
	g.playerImg = loadImage("./assets/player.png")
	g.platformImg = loadImage("./assets/platform.png")
	g.birdLeftImg = loadImage("./assets/bird_left.png")
	g.birdRightImg = loadImage("./assets/bird_right.png")
	g.cloudImg = loadImage("./assets/cloud.png")

	// Set night mode initially based on system time
	hour := time.Now().Hour()
	g.nightMode = hour < 6 || hour > 18

	// Initial platform directly under the player
	g.platforms[0] = Platform{
		X: g.player.X - PlatformWidth/2,
		Y: ScreenHeight - 30,
	}

	// Generate random platforms
	for i := 1; i < PlatformCount; i++ {
		g.platforms[i] = Platform{
			X: rand.Float64() * (ScreenWidth - PlatformWidth),
			Y: float64(i) * (ScreenHeight / PlatformCount),
		}
	}

	// Initialize birds
	for i := 0; i < BirdCount; i++ {
		direction := 1
		if rand.Float64() < 0.5 {
			direction = -1
		}

		g.birds[i] = Bird{
			X:         rand.Float64() * ScreenWidth,
			Y:         rand.Float64() * ScreenHeight / 2, // Birds in upper half
			SpeedX:    BirdSpeedMin + rand.Float64()*(BirdSpeedMax-BirdSpeedMin),
			Direction: direction,
		}
	}

	// Initialize clouds
	for i := 0; i < CloudCount; i++ {
		g.clouds[i] = Cloud{
			X:      rand.Float64() * ScreenWidth,
			Y:      rand.Float64() * ScreenHeight * 0.7, // Clouds in top 70% of screen
			SpeedX: CloudSpeedMin + rand.Float64()*(CloudSpeedMax-CloudSpeedMin),
			Width:  CloudWidth * (0.7 + rand.Float64()*0.6), // Random size variation
			Height: CloudHeight * (0.7 + rand.Float64()*0.6),
			Alpha:  0.5 + rand.Float64()*0.5, // Random transparency
		}
	}

	return g
}

// generateParticle creates a new rain or snow particle
func (g *Game) generateParticle() Particle {
	var particle Particle

	if g.weather == WeatherRain {
		// Raindrop
		particle = Particle{
			X:      rand.Float64() * ScreenWidth,
			Y:      -5,
			SpeedX: 1 + rand.Float64()*2, // slight horizontal movement
			SpeedY: 8 + rand.Float64()*4, // fast fall
			Size:   2 + rand.Float64()*3,
			Alpha:  0.6 + rand.Float64()*0.4,
		}
	} else if g.weather == WeatherSnow {
		// Snowflake
		particle = Particle{
			X:      rand.Float64() * ScreenWidth,
			Y:      -5,
			SpeedX: -1 + rand.Float64()*2, // random drift
			SpeedY: 1 + rand.Float64()*2,  // slow fall
			Size:   2 + rand.Float64()*4,
			Alpha:  0.7 + rand.Float64()*0.3,
		}
	}

	return particle
}

// Update updates the game state
func (g *Game) Update() error {
	if g.gameOver {
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			*g = *NewGame()
		}
		return nil
	}

	// Toggle night mode with 'N' key
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		g.nightMode = !g.nightMode
	}

	// Toggle weather with 'W' key
	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.weather = (g.weather + 1) % 3 // Cycle through weather types
		g.particles = g.particles[:0]   // Clear particles
	}

	// Update day/night cycle based on time
	elapsed := time.Since(g.startTime).Seconds()
	cycleSeconds := g.cycleTime.Seconds()

	// Auto-toggle night mode based on cycle time
	if int(elapsed/cycleSeconds)%2 == 1 {
		g.nightMode = true
	} else {
		g.nightMode = false
	}

	// Weather timer and changes
	g.weatherTimer -= 0.016 // Assume ~60 FPS
	if g.weatherTimer <= 0 {
		// Change weather randomly
		g.weather = rand.Intn(3)
		g.weatherTimer = 15 + rand.Float64()*20 // 15-35 seconds until next change
		g.particles = g.particles[:0]           // Clear particles when weather changes
	}

	// Generate particles based on weather
	if g.weather == WeatherRain {
		// Generate raindrops
		if len(g.particles) < RaindropCount && rand.Float64() < 0.3 {
			g.particles = append(g.particles, g.generateParticle())
		}
	} else if g.weather == WeatherSnow {
		// Generate snowflakes
		if len(g.particles) < SnowflakeCount && rand.Float64() < 0.2 {
			g.particles = append(g.particles, g.generateParticle())
		}
	}

	// Update particles
	for i := 0; i < len(g.particles); i++ {
		g.particles[i].X += g.particles[i].SpeedX
		g.particles[i].Y += g.particles[i].SpeedY

		// Remove particles that go off screen
		if g.particles[i].Y > ScreenHeight {
			g.particles[i] = g.particles[len(g.particles)-1]
			g.particles = g.particles[:len(g.particles)-1]
			i--
		}
	}

	// Handle input
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.player.X -= 3
		g.player.FacingRight = false
		if g.player.X < 0 {
			g.player.X = ScreenWidth
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.player.X += 3
		g.player.FacingRight = true
		if g.player.X > ScreenWidth {
			g.player.X = 0
		}
	}

	// Apply gravity
	g.player.VelocityY += Gravity
	g.player.Y += g.player.VelocityY

	// Update cloud positions
	for i := range g.clouds {
		g.clouds[i].X += g.clouds[i].SpeedX

		// Wrap around screen
		if g.clouds[i].X > ScreenWidth {
			g.clouds[i].X = -g.clouds[i].Width
		}
	}

	// Update bird positions
	for i := range g.birds {
		b := &g.birds[i]
		b.X += b.SpeedX * float64(b.Direction)

		// Wrap around screen
		if b.X < -BirdWidth && b.Direction < 0 {
			b.X = ScreenWidth
		} else if b.X > ScreenWidth && b.Direction > 0 {
			b.X = -BirdWidth
		}

		// Check for collision with player
		if g.player.X+PlayerWidth/4 >= b.X &&
			g.player.X-PlayerWidth/4 <= b.X+BirdWidth &&
			g.player.Y+PlayerHeight/4 >= b.Y &&
			g.player.Y-PlayerHeight/4 <= b.Y+BirdHeight {
			g.gameOver = true
		}
	}

	// Check for platform collisions
	if g.player.VelocityY > 0 {
		for i := range g.platforms {
			p := &g.platforms[i]
			if g.player.X+PlayerWidth/3 >= p.X &&
				g.player.X-PlayerWidth/3 <= p.X+PlatformWidth &&
				g.player.Y+PlayerHeight/2 >= p.Y &&
				g.player.Y+PlayerHeight/2 <= p.Y+PlatformHeight &&
				g.player.VelocityY > 0 {
				g.player.VelocityY = JumpVelocity
			}
		}
	}

	// Camera follows player when jumping high
	highPoint := ScreenHeight * 0.4
	if g.player.Y < highPoint {
		diff := highPoint - g.player.Y
		g.camera += diff
		g.player.Y += diff

		// Move platforms down
		for i := range g.platforms {
			g.platforms[i].Y += diff

			// If platform goes off screen, create new one at the top
			if g.platforms[i].Y > ScreenHeight {
				g.platforms[i].Y = 0
				g.platforms[i].X = rand.Float64() * (ScreenWidth - PlatformWidth)
				g.score++
			}
		}

		// Move birds down
		for i := range g.birds {
			g.birds[i].Y += diff

			// If bird goes off screen, create new one at the top
			if g.birds[i].Y > ScreenHeight {
				g.birds[i].Y = -BirdHeight
				g.birds[i].X = rand.Float64() * ScreenWidth
				g.birds[i].Direction = 1
				if rand.Float64() < 0.5 {
					g.birds[i].Direction = -1
				}
				g.birds[i].SpeedX = BirdSpeedMin + rand.Float64()*(BirdSpeedMax-BirdSpeedMin)
			}
		}

		// Move clouds down
		for i := range g.clouds {
			g.clouds[i].Y += diff

			// If cloud goes off screen, create new one at the top
			if g.clouds[i].Y > ScreenHeight {
				g.clouds[i].Y = -CloudHeight
				g.clouds[i].X = rand.Float64() * ScreenWidth
				g.clouds[i].SpeedX = CloudSpeedMin + rand.Float64()*(CloudSpeedMax-CloudSpeedMin)
				g.clouds[i].Alpha = 0.5 + rand.Float64()*0.5
			}
		}
	}

	// Game over if player falls below screen
	if g.player.Y > ScreenHeight {
		g.gameOver = true
	}

	return nil
}

// Draw draws the game screen
func (g *Game) Draw(screen *ebiten.Image) {
	var bgColor color.RGBA

	// Set background color based on day/night mode
	if g.nightMode {
		bgColor = color.RGBA{0x20, 0x30, 0x50, 0xff} // Dark blue night sky
	} else {
		bgColor = color.RGBA{0x80, 0xa0, 0xc0, 0xff} // Light blue day sky
	}

	// Clear the screen
	screen.Fill(bgColor)

	// Draw clouds
	for _, c := range g.clouds {
		op := &ebiten.DrawImageOptions{}

		// Scale cloud based on its size
		sx := c.Width / CloudWidth
		sy := c.Height / CloudHeight
		op.GeoM.Scale(sx, sy)

		// Position cloud
		op.GeoM.Translate(c.X, c.Y)

		// Adjust cloud color and transparency based on night mode
		if g.nightMode {
			op.ColorM.Scale(0.5, 0.5, 0.7, c.Alpha*0.7) // Darker, bluer clouds at night
		} else {
			op.ColorM.Scale(1, 1, 1, c.Alpha) // Normal white clouds during day
		}

		screen.DrawImage(g.cloudImg, op)
	}

	// Draw platforms
	for _, p := range g.platforms {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X, p.Y)

		// Apply night mode color adjustment
		if g.nightMode {
			op.ColorM.Scale(0.7, 0.7, 0.9, 1) // Slightly darker, bluer at night
		}

		screen.DrawImage(g.platformImg, op)
	}

	// Draw birds
	for _, b := range g.birds {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(b.X, b.Y)

		// Apply night mode color adjustment
		if g.nightMode {
			op.ColorM.Scale(0.7, 0.7, 0.8, 1) // Darker at night
		}

		if b.Direction > 0 {
			screen.DrawImage(g.birdRightImg, op)
		} else {
			screen.DrawImage(g.birdLeftImg, op)
		}
	}

	// Draw weather particles (rain or snow)
	for _, p := range g.particles {
		if g.weather == WeatherRain {
			// Draw raindrops as blue lines
			x1 := p.X
			y1 := p.Y
			x2 := p.X - p.SpeedX*0.5
			y2 := p.Y - p.SpeedY*0.5

			if g.nightMode {
				ebitenutil.DrawLine(screen, x1, y1, x2, y2, color.RGBA{100, 150, 255, uint8(p.Alpha * 255)})
			} else {
				ebitenutil.DrawLine(screen, x1, y1, x2, y2, color.RGBA{70, 130, 230, uint8(p.Alpha * 255)})
			}
		} else if g.weather == WeatherSnow {
			// Draw snowflakes as small white dots
			size := p.Size
			if g.nightMode {
				ebitenutil.DrawRect(screen, p.X, p.Y, size, size, color.RGBA{200, 200, 255, uint8(p.Alpha * 255)})
			} else {
				ebitenutil.DrawRect(screen, p.X, p.Y, size, size, color.RGBA{255, 255, 255, uint8(p.Alpha * 255)})
			}
		}
	}

	// Draw player
	op := &ebiten.DrawImageOptions{}
	if !g.player.FacingRight {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(PlayerWidth, 0)
	}
	op.GeoM.Translate(g.player.X-PlayerWidth/2, g.player.Y-PlayerHeight/2)

	// Apply night mode color adjustment
	if g.nightMode {
		op.ColorM.Scale(0.7, 0.7, 0.9, 1) // Darker at night
	}

	screen.DrawImage(g.playerImg, op)

	// Draw score and info
	ebitenutil.DebugPrintAt(screen, "Score: "+strconv.Itoa(g.score), 5, 5)

	// Display current weather
	var weatherText string
	switch g.weather {
	case WeatherClear:
		weatherText = "Clear"
	case WeatherRain:
		weatherText = "Rainy"
	case WeatherSnow:
		weatherText = "Snowy"
	}

	// Display time mode
	var timeText string
	if g.nightMode {
		timeText = "Night"
	} else {
		timeText = "Day"
	}

	modeText := timeText + " / " + weatherText
	ebitenutil.DebugPrintAt(screen, modeText, 5, 20)
	ebitenutil.DebugPrintAt(screen, "N: Toggle Night, W: Toggle Weather", 5, ScreenHeight-20)

	// Draw game over message
	if g.gameOver {
		msg := "Game Over! Press SPACE to restart"
		ebitenutil.DebugPrintAt(
			screen,
			msg,
			ScreenWidth/2-len(msg)*3,
			ScreenHeight/2,
		)
	}
}

// Layout implements ebiten.Game interface
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
