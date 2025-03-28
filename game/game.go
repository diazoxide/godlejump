package game

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
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
	Gravity        = 0.15   // Reduced gravity for easier control
	JumpVelocity   = -7     // Slightly less powerful jump for better control
	PlatformCount  = 10
	InitialBirdCount = 1    // Start with just 1 bird
	MaxBirdCount   = 8      // Maximum number of birds at highest difficulty
	MaxBirdsPerLine = 2     // Maximum birds allowed at the same height
	CloudCount     = 5
	SnowflakeCount = 40
	RaindropCount  = 50
	InitialBirdSpeedMin = 0.7  // Start with slower birds
	InitialBirdSpeedMax = 1.5
	MaxBirdSpeedMin = 2.5      // Maximum bird speed at highest difficulty
	MaxBirdSpeedMax = 4.0
	CloudSpeedMin  = 0.2
	CloudSpeedMax  = 1.0
	BoostSpawnChance = 0.15   // Increased boost chance (15%)
	BulletSpeed    = 5
	FlyDuration    = 4.0     // Increased flying time
	ShootCooldown  = 0.4     // Shorter cooldown for shooting
	BoostDuration  = 12.0    // Longer boost duration
	ScorePerDifficulty = 20  // Score increment when difficulty increases
)

// Weather types
const (
	WeatherClear = iota
	WeatherRain
	WeatherSnow
)

// Boost types
const (
	BoostNone = iota
	BoostSpeed
	BoostJump
	BoostShield
)

// Bullet represents a projectile fired by the player
type Bullet struct {
	X, Y      float64
	Direction int
	Speed     float64
	Active    bool
}

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
	CanFly      bool
	FlyTimer    float64
	ShootTimer  float64
	Bullets     []Bullet
	BoostType   int
	BoostTimer  float64
}

// Boost represents a powerup that the player can collect
type Boost struct {
	X, Y     float64
	Type     int
	Active   bool
}

// Game implements ebiten.Game interface
type Game struct {
	player       Player
	platforms    []Platform
	birds        []Bird
	clouds       []Cloud
	particles    []Particle
	boosts       []Boost
	bullets      []Bullet
	camera       float64
	score        int
	difficulty   int        // Current difficulty level
	birdCount    int        // Current number of birds (increases with difficulty)
	birdSpeedMin float64    // Current min bird speed (increases with difficulty)
	birdSpeedMax float64    // Current max bird speed (increases with difficulty)
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
	gameTime     float64 // time elapsed since game start (in seconds)
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
			CanFly:      false,
			FlyTimer:    0,
			ShootTimer:  0,
			Bullets:     make([]Bullet, 0),
			BoostType:   BoostNone,
			BoostTimer:  0,
		},
		platforms:    make([]Platform, PlatformCount),
		birds:        make([]Bird, InitialBirdCount),  // Start with fewer birds
		clouds:       make([]Cloud, CloudCount),
		particles:    make([]Particle, 0, RaindropCount),
		boosts:       make([]Boost, 0, 3),
		bullets:      make([]Bullet, 0, 10),
		score:        0,
		difficulty:   0,                      // Start at difficulty 0
		birdCount:    InitialBirdCount,       // Start with initial bird count
		birdSpeedMin: InitialBirdSpeedMin,    // Start with slower birds
		birdSpeedMax: InitialBirdSpeedMax,
		gameOver:     false,
		startTime:    time.Now(),
		cycleTime:    time.Minute * 2,        // Day/night cycle every 2 minutes
		weatherTimer: rand.Float64() * 15,    // Random time until weather changes
		weather:      WeatherClear,
		gameTime:     0,
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
	for i := 0; i < InitialBirdCount; i++ {
		direction := 1
		if rand.Float64() < 0.5 {
			direction = -1
		}

		g.birds[i] = Bird{
			X:         rand.Float64() * ScreenWidth,
			Y:         rand.Float64() * ScreenHeight / 2, // Birds in upper half
			SpeedX:    g.birdSpeedMin + rand.Float64()*(g.birdSpeedMax-g.birdSpeedMin),
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

	// Update game time
	g.gameTime += 1.0 / 60.0 // Assume 60 FPS

	// Toggle night mode with 'N' key
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		g.nightMode = !g.nightMode
	}

	// Toggle weather with 'W' key
	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.weather = (g.weather + 1) % 3 // Cycle through weather types
		g.particles = g.particles[:0]   // Clear particles
	}

	// Update day/night cycle based on game time (every 2 minutes)
	minutesPassed := int(g.gameTime / 60)
	if minutesPassed % 2 == 0 {
		g.nightMode = false
	} else {
		g.nightMode = true
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
	
	// Update boost timers
	if g.player.BoostTimer > 0 {
		g.player.BoostTimer -= 1.0 / 60.0
		if g.player.BoostTimer <= 0 {
			g.player.BoostType = BoostNone
		}
	}

	// Update fly timer
	if g.player.CanFly {
		g.player.FlyTimer -= 1.0 / 60.0
		if g.player.FlyTimer <= 0 {
			g.player.CanFly = false
		}
	}

	// Update shoot timer
	if g.player.ShootTimer > 0 {
		g.player.ShootTimer -= 1.0 / 60.0
	}
	
	// Update boosts
	for i := 0; i < len(g.boosts); i++ {
		// Check for collision with player
		if g.boosts[i].Active &&
			g.player.X+PlayerWidth/3 >= g.boosts[i].X &&
			g.player.X-PlayerWidth/3 <= g.boosts[i].X+PlatformWidth/2 &&
			g.player.Y+PlayerHeight/2 >= g.boosts[i].Y &&
			g.player.Y-PlayerHeight/2 <= g.boosts[i].Y+PlatformHeight*2 {
			
			// Apply boost effect
			g.player.BoostType = g.boosts[i].Type
			g.player.BoostTimer = BoostDuration
			
			// Deactivate boost
			g.boosts[i].Active = false
			
			// If it's the fly boost, enable flying
			if g.boosts[i].Type == BoostJump {
				g.player.CanFly = true
				g.player.FlyTimer = FlyDuration
			}
		}
		
		// Remove inactive boosts
		if !g.boosts[i].Active {
			g.boosts[i] = g.boosts[len(g.boosts)-1]
			g.boosts = g.boosts[:len(g.boosts)-1]
			i--
		}
	}

	// Handle input
	playerSpeed := 3.0
	if g.player.BoostType == BoostSpeed {
		playerSpeed = 5.0 // Speed boost makes player move faster
	}

	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.player.X -= playerSpeed
		g.player.FacingRight = false
		if g.player.X < 0 {
			g.player.X = ScreenWidth
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.player.X += playerSpeed
		g.player.FacingRight = true
		if g.player.X > ScreenWidth {
			g.player.X = 0
		}
	}

	// Fly with Up key (if can fly)
	if (ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW)) && g.player.CanFly {
		g.player.VelocityY = -4 // Fly upward
	}

	// Toggle flying with F key
	if inpututil.IsKeyJustPressed(ebiten.KeyF) && g.player.FlyTimer <= 0 {
		g.player.CanFly = true
		g.player.FlyTimer = FlyDuration
	}

	// Shooting with Space key
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) && g.player.ShootTimer <= 0 {
		// Create a new bullet
		direction := 1
		if !g.player.FacingRight {
			direction = -1
		}
		
		bullet := Bullet{
			X:         g.player.X + float64(direction*PlayerWidth/2),
			Y:         g.player.Y,
			Direction: direction,
			Speed:     BulletSpeed,
			Active:    true,
		}
		
		g.bullets = append(g.bullets, bullet)
		g.player.ShootTimer = ShootCooldown
	}

	// Apply gravity (unless flying)
	g.player.VelocityY += Gravity
	g.player.Y += g.player.VelocityY

	// Update bullets
	for i := 0; i < len(g.bullets); i++ {
		g.bullets[i].X += g.bullets[i].Speed * float64(g.bullets[i].Direction)
		
		// Check if bullet is off screen
		if g.bullets[i].X < 0 || g.bullets[i].X > ScreenWidth {
			g.bullets[i] = g.bullets[len(g.bullets)-1]
			g.bullets = g.bullets[:len(g.bullets)-1]
			i--
			continue
		}
		
		// Check for collision with birds
		for j := range g.birds {
			b := &g.birds[j]
			if g.bullets[i].X >= b.X && 
				g.bullets[i].X <= b.X+BirdWidth &&
				g.bullets[i].Y >= b.Y &&
				g.bullets[i].Y <= b.Y+BirdHeight {
				
				// Remove bird and regenerate it above
				b.Y = -BirdHeight * 2  // Move bird off screen to be regenerated
				
				// Remove bullet
				g.bullets[i] = g.bullets[len(g.bullets)-1]
				g.bullets = g.bullets[:len(g.bullets)-1]
				i--
				break
			}
		}
	}

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
			
			// Shield boost protects against birds
			if g.player.BoostType != BoostShield {
				g.gameOver = true
			} else {
				// Remove bird and regenerate it above instead of game over
				b.Y = -BirdHeight * 2
			}
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
				
				// Check if difficulty should increase
				newDifficulty := g.score / ScorePerDifficulty
				if newDifficulty > g.difficulty {
					g.difficulty = newDifficulty
					
					// Calculate how many birds based on difficulty (cap at MaxBirdCount)
					newBirdCount := InitialBirdCount + g.difficulty
					if newBirdCount > MaxBirdCount {
						newBirdCount = MaxBirdCount
					}
					
					// If we need more birds than we currently have
					if newBirdCount > g.birdCount {
						// Add more birds
						for j := g.birdCount; j < newBirdCount; j++ {
							direction := 1
							if rand.Float64() < 0.5 {
								direction = -1
							}
							
							// Place new bird above the screen
							newBird := Bird{
								X:         rand.Float64() * ScreenWidth,
								Y:         -BirdHeight * float64(1+j%MaxBirdsPerLine), // Stagger birds vertically
								SpeedX:    g.birdSpeedMin + rand.Float64()*(g.birdSpeedMax-g.birdSpeedMin),
								Direction: direction,
							}
							g.birds = append(g.birds, newBird)
						}
						g.birdCount = newBirdCount
					}
					
					// Increase bird speed gradually up to max values
					progressFactor := float64(g.difficulty) / 10 // Full speed increase over ~10 difficulty levels
					if progressFactor > 1 {
						progressFactor = 1
					}
					
					// Linear interpolation between initial and max speeds
					g.birdSpeedMin = InitialBirdSpeedMin + progressFactor*(MaxBirdSpeedMin-InitialBirdSpeedMin)
					g.birdSpeedMax = InitialBirdSpeedMax + progressFactor*(MaxBirdSpeedMax-InitialBirdSpeedMax)
				}
				
				// Potentially spawn a boost on this platform
				if rand.Float64() < BoostSpawnChance {
					boostType := rand.Intn(3) + 1 // Random boost type 1-3
					
					boost := Boost{
						X:      g.platforms[i].X + PlatformWidth/4,
						Y:      g.platforms[i].Y - PlatformHeight*2,
						Type:   boostType,
						Active: true,
					}
					
					g.boosts = append(g.boosts, boost)
				}
			}
		}

		// Move birds down
		for i := range g.birds {
			g.birds[i].Y += diff

			// If bird goes off screen, create new one at the top
			if g.birds[i].Y > ScreenHeight {
				// Check for existing birds at similar heights (enforce max birds per line)
				validPosition := false
				maxAttempts := 10
				attempts := 0
				
				// Keep trying new positions until we find a valid one
				for !validPosition && attempts < maxAttempts {
					// Start with a random Y position above the screen
					newY := -BirdHeight - float64(rand.Intn(3))*BirdHeight
					
					// Check if this position would cause more than MaxBirdsPerLine at same height
					birdsAtSameHeight := 0
					for j := range g.birds {
						if j != i && math.Abs(g.birds[j].Y-newY) < BirdHeight {
							birdsAtSameHeight++
						}
					}
					
					// If we have fewer than max birds per line at this height, it's valid
					if birdsAtSameHeight < MaxBirdsPerLine {
						g.birds[i].Y = newY
						validPosition = true
					}
					
					attempts++
				}
				
				// If we couldn't find a valid position after max attempts, place bird higher
				if !validPosition {
					g.birds[i].Y = -BirdHeight * (5 + rand.Float64()*5)
				}
				
				g.birds[i].X = rand.Float64() * ScreenWidth
				g.birds[i].Direction = 1
				if rand.Float64() < 0.5 {
					g.birds[i].Direction = -1
				}
				
				// Use current dynamic speed range
				g.birds[i].SpeedX = g.birdSpeedMin + rand.Float64()*(g.birdSpeedMax-g.birdSpeedMin)
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
	
	// Draw boosts
	for _, b := range g.boosts {
		if b.Active {
			var boostColor color.RGBA
			
			// Different colors for different boost types
			switch b.Type {
			case BoostSpeed:
				boostColor = color.RGBA{255, 50, 50, 255} // Red for speed
			case BoostJump:
				boostColor = color.RGBA{50, 255, 50, 255} // Green for jump/fly
			case BoostShield:
				boostColor = color.RGBA{50, 50, 255, 255} // Blue for shield
			}
			
			// Adjust color for night mode
			if g.nightMode {
				boostColor.R = uint8(float64(boostColor.R) * 0.7)
				boostColor.G = uint8(float64(boostColor.G) * 0.7)
				boostColor.B = uint8(float64(boostColor.B) * 0.8)
			}
			
			// Draw boost as a colored circle
			ebitenutil.DrawCircle(screen, b.X, b.Y, 10, boostColor)
		}
	}
	
	// Draw bullets
	for _, b := range g.bullets {
		if b.Active {
			bulletColor := color.RGBA{255, 255, 0, 255} // Yellow bullets
			if g.nightMode {
				bulletColor = color.RGBA{200, 200, 50, 255} // Darker yellow at night
			}
			
			ebitenutil.DrawCircle(screen, b.X, b.Y, 3, bulletColor)
		}
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
	
	// Display active boost
	var boostText string
	switch g.player.BoostType {
	case BoostNone:
		boostText = "No Boost"
	case BoostSpeed:
		boostText = "Speed Boost: " + fmt.Sprintf("%.1f", g.player.BoostTimer)
	case BoostJump:
		boostText = "Jump Boost: " + fmt.Sprintf("%.1f", g.player.BoostTimer)
	case BoostShield:
		boostText = "Shield Boost: " + fmt.Sprintf("%.1f", g.player.BoostTimer)
	}
	ebitenutil.DebugPrintAt(screen, boostText, 5, 35)
	
	// Display if flying is active
	if g.player.CanFly {
		flyText := "Flying: " + fmt.Sprintf("%.1f", g.player.FlyTimer)
		ebitenutil.DebugPrintAt(screen, flyText, 5, 50)
	}
	
	// Display difficulty level
	difficultyText := fmt.Sprintf("Difficulty: %d (Birds: %d)", g.difficulty, len(g.birds))
	ebitenutil.DebugPrintAt(screen, difficultyText, 5, 65)
	
	// Controls info at bottom
	ebitenutil.DebugPrintAt(screen, "Left/Right: Move, F: Fly, Space: Shoot", 5, ScreenHeight-35)
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
