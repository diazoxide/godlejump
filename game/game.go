package game

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

//go:embed assets/*.png
var gameAssets embed.FS

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

	// Day cycle constants
	DayCycleLength = 1000.0  // Score points for a full day cycle
	SunriseStart   = 0.0     // Sunrise phase start (0.0 - 1.0)
	SunriseEnd     = 0.2     // Sunrise phase end
	DayStart       = 0.2     // Day phase start
	DayEnd         = 0.7     // Day phase end
	SunsetStart    = 0.7     // Sunset phase start
	SunsetEnd      = 0.9     // Sunset phase end
	NightStart     = 0.9     // Night phase start
	NightEnd       = 1.0     // Night phase end (wraps to 0.0)

	// Mountain parameters
	MountainCount  = 3       // Number of mountain layers
	MountainPoints = 8       // Control points for curves
	MountainDetail = 100     // Reduced detail but still smooth
	ParallaxFactor = 0.1     // Parallax factor
	MountainSliceHeight = 4  // Draw mountains in larger slices for better performance

	// Time phases in natural order
	TimeMidnight  = 0.0
	TimeNight     = 0.2
	TimeSunrise   = 0.4
	TimeMorning   = 0.6
	TimeDay       = 0.8
	TimeSunset    = 1.0
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

// Platform types
const (
	PlatformNormal = iota
	PlatformSticky
	PlatformDisappearing
)

// Platform animation states
const (
	PlatformIntact = iota
	PlatformBreaking
	PlatformBroken
)

// Bullet represents a projectile fired by the player
type Bullet struct {
	X, Y      float64
	Direction int
	Speed     float64
	Active    bool
}

// Platform represents a platform in the game
type Platform struct {
	X, Y        float64
	Type        int
	State       int
	BreakTimer  float64 // Timer for breaking animation
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

// Add this type and the color sets before the Game struct
type ColorSet struct {
	skyColors     [7]color.RGBA
	mountainTints [3]color.RGBA
}

// Add these types and functions before the Game struct
type HSV struct {
	H, S, V float64
}

type GradientParams struct {
	baseHue        float64  // Base hue for the gradient
	hueRange      float64  // How much the hue can vary
	satRange      [2]float64  // Min/max saturation
	valRange      [2]float64  // Min/max value/brightness
	mountainDepth float64  // How much darker/different mountains are
}

// Convert HSV to RGB color
func hsvToRGB(hsv HSV) color.RGBA {
	H, S, V := hsv.H, hsv.S, hsv.V
	
	// Constrain values
	H = math.Mod(H, 360)
	if S < 0 { S = 0 } else if S > 1 { S = 1 }
	if V < 0 { V = 0 } else if V > 1 { V = 1 }
	
	C := V * S
	X := C * (1 - math.Abs(math.Mod(H/60, 2)-1))
	M := V - C
	
	var R, G, B float64
	switch {
	case H < 60:
		R, G, B = C, X, 0
	case H < 120:
		R, G, B = X, C, 0
	case H < 180:
		R, G, B = 0, C, X
	case H < 240:
		R, G, B = 0, X, C
	case H < 300:
		R, G, B = X, 0, C
	default:
		R, G, B = C, 0, X
	}
	
	return color.RGBA{
		R: uint8((R + M) * 255),
		G: uint8((G + M) * 255),
		B: uint8((B + M) * 255),
		A: 255,
	}
}

// Add these helper functions for improved color transitions
func cosineInterpolate(a, b, t float64) float64 {
	ft := t * math.Pi
	f := (1 - math.Cos(ft)) * 0.5
	return a*(1-f) + b*f
}

func blend(colors []HSV, t float64) HSV {
	if t <= 0 {
		return colors[0]
	}
	if t >= 1 {
		return colors[len(colors)-1]
	}
	
	segment := t * float64(len(colors)-1)
	i := int(segment)
	t = segment - float64(i)
	
	if i+1 >= len(colors) {
		return colors[len(colors)-1]
	}
	
	// Cosine interpolation for smoother transitions
	return HSV{
		H: cosineInterpolate(colors[i].H, colors[i+1].H, t),
		S: cosineInterpolate(colors[i].S, colors[i+1].S, t),
		V: cosineInterpolate(colors[i].V, colors[i+1].V, t),
	}
}

// Replace getGradientParams with this improved version
func getGradientParams(timeOfDay float64) GradientParams {
	// Define key colors for different times of day
	keyColors := []struct {
		time float64
		sky  []HSV
		mountain HSV
	}{
		{ // Midnight
			time: 0.0,
			sky: []HSV{
				{H: 230, S: 0.6, V: 0.2},  // Deep blue top
				{H: 235, S: 0.5, V: 0.15}, // Middle
				{H: 240, S: 0.4, V: 0.1},  // Bottom
			},
			mountain: HSV{H: 235, S: 0.4, V: 0.1},
		},
		{ // Pre-dawn
			time: 0.2,
			sky: []HSV{
				{H: 240, S: 0.5, V: 0.3},  // Dark blue top
				{H: 260, S: 0.4, V: 0.2},  // Purple middle
				{H: 280, S: 0.3, V: 0.15}, // Deep purple bottom
			},
			mountain: HSV{H: 250, S: 0.3, V: 0.15},
		},
		{ // Dawn
			time: 0.3,
			sky: []HSV{
				{H: 200, S: 0.4, V: 0.6},  // Light blue top
				{H: 35, S: 0.7, V: 0.7},   // Orange middle
				{H: 20, S: 0.8, V: 0.8},   // Warm orange bottom
			},
			mountain: HSV{H: 30, S: 0.5, V: 0.3},
		},
		{ // Morning
			time: 0.4,
			sky: []HSV{
				{H: 195, S: 0.4, V: 0.9},  // Sky blue top
				{H: 200, S: 0.3, V: 0.8},  // Light blue middle
				{H: 205, S: 0.2, V: 0.7},  // Pale blue bottom
			},
			mountain: HSV{H: 200, S: 0.3, V: 0.4},
		},
		{ // Noon
			time: 0.5,
			sky: []HSV{
				{H: 210, S: 0.3, V: 0.9},  // Bright blue top
				{H: 205, S: 0.2, V: 0.85}, // Light blue middle
				{H: 200, S: 0.1, V: 0.8},  // Pale blue bottom
			},
			mountain: HSV{H: 205, S: 0.2, V: 0.5},
		},
		{ // Afternoon
			time: 0.7,
			sky: []HSV{
				{H: 210, S: 0.4, V: 0.8},  // Blue top
				{H: 215, S: 0.3, V: 0.7},  // Medium blue middle
				{H: 220, S: 0.2, V: 0.6},  // Light blue bottom
			},
			mountain: HSV{H: 215, S: 0.3, V: 0.4},
		},
		{ // Sunset
			time: 0.8,
			sky: []HSV{
				{H: 200, S: 0.5, V: 0.6},  // Deep blue top
				{H: 30, S: 0.8, V: 0.7},   // Orange middle
				{H: 15, S: 0.9, V: 0.8},   // Red-orange bottom
			},
			mountain: HSV{H: 20, S: 0.6, V: 0.3},
		},
		{ // Night
			time: 0.9,
			sky: []HSV{
				{H: 230, S: 0.6, V: 0.3},  // Dark blue top
				{H: 240, S: 0.5, V: 0.2},  // Deep blue middle
				{H: 250, S: 0.4, V: 0.1},  // Very deep blue bottom
			},
			mountain: HSV{H: 235, S: 0.4, V: 0.15},
		},
	}

	// Find the two time periods we're between
	var idx int
	for i := range keyColors {
		if timeOfDay < keyColors[i].time {
			idx = i - 1
			break
		}
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(keyColors)-1 {
		idx = len(keyColors) - 2
	}

	// Calculate progress between the two time periods
	t := (timeOfDay - keyColors[idx].time) / (keyColors[idx+1].time - keyColors[idx].time)
	t = smoothstep(t) // Apply smoothstep for better transitions

	// Create parameters based on the interpolation
	params := GradientParams{
		baseHue: cosineInterpolate(keyColors[idx].mountain.H, keyColors[idx+1].mountain.H, t),
		hueRange: 15, // Reduced range for more subtle variations
		satRange: [2]float64{
			cosineInterpolate(keyColors[idx].mountain.S-0.1, keyColors[idx+1].mountain.S-0.1, t),
			cosineInterpolate(keyColors[idx].mountain.S+0.1, keyColors[idx+1].mountain.S+0.1, t),
		},
		valRange: [2]float64{
			cosineInterpolate(keyColors[idx].mountain.V-0.1, keyColors[idx+1].mountain.V-0.1, t),
			cosineInterpolate(keyColors[idx].mountain.V+0.1, keyColors[idx+1].mountain.V+0.1, t),
		},
		mountainDepth: 0.2, // Consistent mountain depth
	}

	return params
}

// Replace generateColorSet with this improved version
func generateColorSet(params GradientParams) ColorSet {
	var result ColorSet

	// Generate sky gradient colors with smoother transitions
	for i := range result.skyColors {
		progress := float64(i) / float64(len(result.skyColors)-1)
		
		// Use subtle sine waves for variation
		hue := params.baseHue + params.hueRange*0.5*math.Sin(progress*math.Pi)
		sat := params.satRange[0] + (params.satRange[1]-params.satRange[0])*smoothstep(progress)
		val := params.valRange[1] - (params.valRange[1]-params.valRange[0])*smoothstep(progress)
		
		// Add very subtle variation
		hue += 2 * math.Sin(progress*2*math.Pi)
		sat += 0.05 * math.Sin(progress*3*math.Pi)
		val += 0.05 * math.Sin(progress*2*math.Pi)
		
		result.skyColors[i] = hsvToRGB(HSV{hue, sat, val})
	}

	// Generate mountain colors with proper depth perception
	for i := range result.mountainTints {
		progress := float64(i) / float64(len(result.mountainTints)-1)
		
		// Gradually adjust mountain colors for depth
		hue := params.baseHue + 5*progress // Slight hue shift for depth
		sat := params.satRange[0] * (1 - 0.2*progress)
		val := params.valRange[0] * (1 - params.mountainDepth*progress)
		
		result.mountainTints[i] = hsvToRGB(HSV{hue, sat, val})
	}

	return result
}

// Replace the getColorSetForTime function with this:
func getColorSetForTime(timeOfDay float64) ColorSet {
	params := getGradientParams(timeOfDay)
	return generateColorSet(params)
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
	stars        []struct{ x, y, brightness float64 }  // Add stars
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
	mountainImgs []*ebiten.Image  // Mountain layer images
	gameOver     bool
	nightMode    bool
	weather      int
	startTime    time.Time
	cycleTime    time.Duration
	weatherTimer float64 // counter for weather changes
	gameTime     float64 // time elapsed since game start (in seconds)
	initialTimeOfDay float64  // Random initial time of day (0.0 - 1.0)
	stuckToPlatform *Platform
	stuckTimer      float64    // For visual effect
	jumpPressed     bool       // Track jump button state
	canJumpRelease  bool       // Whether player can release from sticky platform
}

// loadImage loads an image from embedded assets
func loadImage(path string) *ebiten.Image {
	// Remove leading "./" from path if present
	if len(path) > 2 && path[:2] == "./" {
		path = path[2:]
	}

	imgBytes, err := gameAssets.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read embedded image: %v", err)
	}

	img, _, err := image.Decode(bytes.NewReader(imgBytes))
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
		stars:        make([]struct{ x, y, brightness float64 }, 100),  // Initialize stars
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
		initialTimeOfDay: rand.Float64(),
		mountainImgs: make([]*ebiten.Image, 3),
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
		X:    g.player.X - PlatformWidth/2,
		Y:    ScreenHeight - 30,
		Type: PlatformNormal,
	}

	// Generate random platforms
	for i := 1; i < PlatformCount; i++ {
		platformType := PlatformNormal
		
		// Platform type distribution
		rnd := rand.Float64()
		if rnd < 0.2 { // 20% chance for sticky platform
			platformType = PlatformSticky
		} else if rnd < 0.35 { // 15% chance for disappearing platform
			platformType = PlatformDisappearing
		}
		
		g.platforms[i] = Platform{
			X:          rand.Float64() * (ScreenWidth - PlatformWidth),
			Y:          float64(i) * (ScreenHeight / PlatformCount),
			Type:       platformType,
			State:      PlatformIntact,
			BreakTimer: 0,
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

	// Load mountain images
	g.mountainImgs = make([]*ebiten.Image, 3)
	for i := 0; i < 3; i++ {
		g.mountainImgs[i] = loadImage(fmt.Sprintf("./assets/mountains_%d.png", i))
	}

	// Initialize stars with random positions
	for i := range g.stars {
		g.stars[i].x = rand.Float64() * float64(ScreenWidth)
		g.stars[i].y = rand.Float64() * float64(ScreenHeight) * 0.7 // Stars in top 70% of screen
		g.stars[i].brightness = 0.3 + rand.Float64()*0.7 // Random brightness
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

	// Toggle weather with 'W' key
	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.weather = (g.weather + 1) % 3 // Cycle through weather types
		g.particles = g.particles[:0]   // Clear particles
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
	
	// Handle sticky platform release
	jumpKey := ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW)
	spaceKey := ebiten.IsKeyPressed(ebiten.KeySpace)
	
	// Check for jump key press
	if jumpKey || spaceKey {
		if !g.jumpPressed {
			// Key was just pressed
			if g.stuckToPlatform != nil {
				// Release from platform with a higher jump
				g.player.VelocityY = float64(JumpVelocity) * 1.2
				g.stuckToPlatform = nil
				g.stuckTimer = 0
			}
		}
		g.jumpPressed = true
	} else {
		g.jumpPressed = false
	}

	// Update platform states
	for i := range g.platforms {
		p := &g.platforms[i]
		
		// Update disappearing platform state
		if p.Type == PlatformDisappearing && p.State == PlatformBreaking {
			p.BreakTimer -= 1.0 / 60.0
			if p.BreakTimer <= 0 {
				p.State = PlatformBroken
			}
		}
		
		// Check for collision with player
		if g.player.X+PlayerWidth/3 >= p.X &&
			g.player.X-PlayerWidth/3 <= p.X+PlatformWidth &&
			g.player.Y+PlayerHeight/2 >= p.Y &&
			g.player.Y+PlayerHeight/2 <= p.Y+PlatformHeight &&
			g.player.VelocityY > 0 {
			
			// Skip broken platforms
			if p.Type == PlatformDisappearing && p.State == PlatformBroken {
				continue
			}
			
			if p.Type == PlatformSticky {
				// Stick to platform
				g.stuckToPlatform = p
				g.stuckTimer = 0
				g.player.VelocityY = 0
				g.player.Y = p.Y - PlayerHeight/2 // Align player with platform
				g.canJumpRelease = false // Require new jump press to release
			} else if p.Type == PlatformDisappearing && p.State == PlatformIntact {
				// Start breaking animation for disappearing platform
				p.State = PlatformBreaking
				p.BreakTimer = 0.3 // Time until platform breaks
				
				// Allow player to jump off it once
				jumpForce := float64(JumpVelocity)
				if g.player.BoostType == BoostJump {
					jumpForce *= 1.5
				}
				g.player.VelocityY = jumpForce
			} else {
				// Normal platform bounce
				jumpForce := float64(JumpVelocity)
				if g.player.BoostType == BoostJump {
					jumpForce *= 1.5
				}
				g.player.VelocityY = jumpForce
			}
		}
	}

	// Update stuck timer for animation
	if g.stuckToPlatform != nil {
		g.stuckTimer += 1.0 / 60.0
		// Keep player stuck to platform
		g.player.Y = g.stuckToPlatform.Y - PlayerHeight/2
		g.player.VelocityY = 0
	}

	// Update boost effects
	if g.player.BoostType != BoostNone {
		g.player.BoostTimer -= 1.0 / 60.0
		if g.player.BoostTimer <= 0 {
			g.player.BoostType = BoostNone
			g.player.BoostTimer = 0
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

	// Platform collisions are handled in the Update platform states section above

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
				
				// Reset platform state if it was broken
				if g.platforms[i].Type == PlatformDisappearing {
					g.platforms[i].State = PlatformIntact
				}
				
				// Generate a new platform type
				platformType := PlatformNormal
				rnd := rand.Float64()
				if rnd < 0.2 { // 20% chance for sticky platform
					platformType = PlatformSticky
				} else if rnd < 0.35 { // 15% chance for disappearing platform
					platformType = PlatformDisappearing
				}
				g.platforms[i].Type = platformType
				
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
	// Calculate current time of day (0.0 - 1.0)
	timeOfDay := math.Mod(float64(g.score)/DayCycleLength + g.initialTimeOfDay, 1.0)

	// Get color set for current time
	colorSet := getColorSetForTime(timeOfDay)

	// Draw sky gradient
	for y := 0; y < ScreenHeight; y++ {
		progress := float64(y) / float64(ScreenHeight)
		
		// Get base colors for interpolation
		baseColors := colorSet.skyColors
		
		// Calculate smooth color transition
		var color color.RGBA
		
		// Use continuous interpolation across all colors
		t := progress * float64(len(baseColors)-1)
		i := int(t)
		if i >= len(baseColors)-1 {
			color = baseColors[len(baseColors)-1]
		} else {
			// Get fractional progress between two colors
			frac := t - float64(i)
			
			// Use smoothstep for better color blending
			frac = smoothstep(frac)
			
			// Get the two colors to blend between
			c1 := baseColors[i]
			c2 := baseColors[i+1]
			
			// Interpolate in RGB space with gamma correction
			r := uint8(math.Pow((math.Pow(float64(c1.R)/255, 2.2)*(1-frac) + math.Pow(float64(c2.R)/255, 2.2)*frac), 1/2.2) * 255)
			g := uint8(math.Pow((math.Pow(float64(c1.G)/255, 2.2)*(1-frac) + math.Pow(float64(c2.G)/255, 2.2)*frac), 1/2.2) * 255)
			b := uint8(math.Pow((math.Pow(float64(c1.B)/255, 2.2)*(1-frac) + math.Pow(float64(c2.B)/255, 2.2)*frac), 1/2.2) * 255)
			color.R = r
			color.G = g
			color.B = b
			color.A = 255
		}
		
		// Apply subtle atmospheric perspective
		brightness := 1.0 - 0.15*math.Pow(progress, 2.0)
		color.R = uint8(float64(color.R) * brightness)
		color.G = uint8(float64(color.G) * brightness)
		color.B = uint8(float64(color.B) * brightness)
		
		ebitenutil.DrawRect(screen, 0, float64(y), ScreenWidth, 1, color)
	}

	// Draw stars during night time
	if timeOfDay > SunsetStart || timeOfDay < SunriseEnd {
		// Calculate star visibility
		starAlpha := 0.0
		if timeOfDay > SunsetStart && timeOfDay < SunsetEnd {
			// Fade in during sunset
			starAlpha = (timeOfDay - SunsetStart) / (SunsetEnd - SunsetStart)
		} else if timeOfDay > SunsetEnd || timeOfDay < SunriseStart {
			// Full visibility during night
			starAlpha = 1.0
		} else if timeOfDay < SunriseEnd {
			// Fade out during sunrise
			starAlpha = 1.0 - (timeOfDay / SunriseEnd)
		}

		// Draw stars with twinkling effect
		for _, star := range g.stars {
			// Calculate star position with parallax
			starX := math.Mod(star.x - g.camera*0.05, float64(ScreenWidth))
			if starX < 0 {
				starX += float64(ScreenWidth)
			}

			// Add twinkling effect
			twinkle := 0.7 + 0.3*math.Sin(g.gameTime*2+star.x*0.1)
			
			// Calculate final brightness
			brightness := star.brightness * twinkle * starAlpha
			
			// Draw star as a small white dot
			starColor := color.RGBA{
				R: uint8(255 * brightness),
				G: uint8(255 * brightness),
				B: uint8(255 * brightness),
				A: uint8(255 * brightness),
			}
			
			// Draw star with slight glow effect
			size := 1.0 + star.brightness*1.0
			ebitenutil.DrawCircle(screen, starX, star.y, size, starColor)
			
			// Add a subtle glow
			glowColor := color.RGBA{
				R: uint8(255 * brightness * 0.3),
				G: uint8(255 * brightness * 0.3),
				B: uint8(255 * brightness * 0.3),
				A: uint8(255 * brightness * 0.3),
			}
			ebitenutil.DrawCircle(screen, starX, star.y, size*2, glowColor)
		}
	}

	// Draw mountain layers
	for i := len(g.mountainImgs) - 1; i >= 0; i-- {
		op := &ebiten.DrawImageOptions{}
		
		// Calculate parallax offset
		parallaxOffset := g.camera * float64(i+1) * 0.15
		
		// Scale mountains
		scaleX := float64(ScreenWidth) / 1200.0 * 1.2
		scaleY := float64(ScreenHeight) / 800.0 * 1.5
		op.GeoM.Scale(scaleX, scaleY)
		
		// Position mountains
		yOffset := float64(ScreenHeight) * 0.3
		op.GeoM.Translate(-math.Mod(parallaxOffset, float64(ScreenWidth)), -yOffset)
		
		// Apply mountain tint
		tint := colorSet.mountainTints[i]
		op.ColorM.Scale(
			float64(tint.R)/255.0,
			float64(tint.G)/255.0,
			float64(tint.B)/255.0,
			1,
		)
		
		// Draw main layer and tiled copy
		screen.DrawImage(g.mountainImgs[i], op)
		op.GeoM.Reset()
		op.GeoM.Scale(scaleX, scaleY)
		op.GeoM.Translate(-math.Mod(parallaxOffset, float64(ScreenWidth))+float64(ScreenWidth), -yOffset)
		screen.DrawImage(g.mountainImgs[i], op)
	}

	// Draw clouds with adjusted transparency based on time of day
	for _, c := range g.clouds {
		op := &ebiten.DrawImageOptions{}
		sx := c.Width / CloudWidth
		sy := c.Height / CloudHeight
		op.GeoM.Scale(sx, sy)
		op.GeoM.Translate(c.X, c.Y)

		// Adjust cloud visibility based on time of day
		alpha := c.Alpha
		if timeOfDay > SunsetStart || timeOfDay < SunriseEnd {
			alpha *= 0.5 // Less visible clouds during night/twilight
		}
		op.ColorM.Scale(1, 1, 1, alpha)

		screen.DrawImage(g.cloudImg, op)
	}

	// Draw platforms
	for i := range g.platforms {
		p := &g.platforms[i]  // Get pointer to platform
		
		// Skip drawing broken platforms
		if p.Type == PlatformDisappearing && p.State == PlatformBroken {
			continue
		}
		
		if p.Type == PlatformSticky {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(p.X, p.Y)

			// Apply night mode color adjustment
			if g.nightMode {
				op.ColorM.Scale(0.7, 0.7, 0.9, 1)
			}

			// Yellow-amber color for sticky platforms
			op.ColorM.Scale(1.2, 1.0, 0.4, 1)
			
			// Add pulsing effect when player is stuck
			if p == g.stuckToPlatform {
				pulse := 0.3 + 0.2*math.Sin(g.stuckTimer*6.0)
				op.ColorM.Scale(1.0+pulse, 1.0+pulse, 0.5+pulse, 1)
				
				// Draw "Jump!" text
				ebitenutil.DebugPrintAt(screen, "Jump!", int(p.X)+20, int(p.Y)-15)
				
				// Draw sticky effect particles
				for i := 0; i < 3; i++ {
					if rand.Float64() < 0.7 {
						particleX := p.X + rand.Float64()*PlatformWidth
						particleY := p.Y + rand.Float64()*PlatformHeight/2
						particleColor := color.RGBA{255, 220, 100, 180}
						ebitenutil.DrawCircle(screen, particleX, particleY, 1.5, particleColor)
					}
				}
			}

			screen.DrawImage(g.platformImg, op)
		} else if p.Type == PlatformDisappearing {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(p.X, p.Y)

			// Apply night mode color adjustment
			if g.nightMode {
				op.ColorM.Scale(0.7, 0.7, 0.9, 1)
			}

			// Red color for disappearing platforms
			op.ColorM.Scale(1.0, 0.6, 0.6, 1)
			
			// Apply cracking animation effect
			if p.State == PlatformBreaking {
				// Make platform fade and shake as it breaks
				breakProgress := 1.0 - (p.BreakTimer / 0.3)
				op.ColorM.Scale(1, 1, 1, 1.0-breakProgress*0.5)
				
				// Add shaking effect
				shakeX := (rand.Float64()*2 - 1) * breakProgress * 3
				shakeY := (rand.Float64()*2 - 1) * breakProgress * 2
				op.GeoM.Translate(shakeX, shakeY)
				
				// Draw cracks
				for i := 0; i < 5; i++ {
					crackX1 := p.X + rand.Float64()*PlatformWidth
					crackY1 := p.Y + rand.Float64()*PlatformHeight
					crackX2 := crackX1 + (rand.Float64()*2-1)*10*breakProgress
					crackY2 := crackY1 + (rand.Float64()*2-1)*5*breakProgress
					ebitenutil.DrawLine(screen, crackX1, crackY1, crackX2, crackY2, color.RGBA{80, 80, 80, 200})
				}
			}

			screen.DrawImage(g.platformImg, op)
		} else {
			// Normal platform drawing
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(p.X, p.Y)

			// Apply night mode color adjustment
			if g.nightMode {
				op.ColorM.Scale(0.7, 0.7, 0.9, 1)
			}

			screen.DrawImage(g.platformImg, op)
		}
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
	ebitenutil.DebugPrintAt(screen, "W: Toggle Weather", 5, ScreenHeight-20)

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

	// Draw help text at the bottom
	ebitenutil.DebugPrintAt(screen, "Press UP/W or SPACE to release from sticky platforms!", 5, ScreenHeight-50)
}

// Layout implements ebiten.Game interface
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// lerpColor interpolates between two colors
func lerpColor(c1, c2 color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c1.R) + t*float64(c2.R-c1.R)),
		G: uint8(float64(c1.G) + t*float64(c2.G-c1.G)),
		B: uint8(float64(c1.B) + t*float64(c2.B-c1.B)),
		A: uint8(float64(c1.A) + t*float64(c2.A-c1.A)),
	}
}

// bezierPoint calculates a point on a Bézier curve
func bezierPoint(points []struct{ X, Y float64 }, t float64) (float64, float64) {
	n := len(points) - 1
	x, y := 0.0, 0.0
	
	for i := 0; i <= n; i++ {
		b := bernstein(n, i, t)
		x += points[i].X * b
		y += points[i].Y * b
	}
	
	return x, y
}

// bernstein calculates the Bernstein polynomial
func bernstein(n, i int, t float64) float64 {
	return float64(combination(n, i)) * math.Pow(t, float64(i)) * math.Pow(1-t, float64(n-i))
}

// combination calculates the binomial coefficient
func combination(n, k int) int {
	if k == 0 || k == n {
		return 1
	}
	if k > n {
		return 0
	}
	return combination(n-1, k-1) + combination(n-1, k)
}

// adjustColorBrightness adjusts the brightness of a color by a factor
func adjustColorBrightness(c color.RGBA, factor float64) color.RGBA {
	adjust := func(v uint8) uint8 {
		f := float64(v) * (1 + factor)
		if f < 0 {
			f = 0
		} else if f > 255 {
			f = 255
		}
		return uint8(f)
	}
	
	return color.RGBA{
		R: adjust(c.R),
		G: adjust(c.G),
		B: adjust(c.B),
		A: c.A,
	}
}

// Update mountainGradient for better performance
func mountainGradient(baseColor color.RGBA, skyBottom color.RGBA, height, maxHeight, timeOfDay float64) color.RGBA {
	// Calculate snow line based on height
	snowLine := maxHeight * 0.75
	snowAmount := math.Max(0, (height-snowLine)/(maxHeight-snowLine))
	
	// Adjust colors based on time of day (simplified calculation)
	sunlightFactor := 0.0
	if timeOfDay >= DayStart && timeOfDay <= DayEnd {
		sunlightFactor = 1.0
	} else if timeOfDay < DayStart {
		sunlightFactor = (timeOfDay - SunriseStart) / (DayStart - SunriseStart)
	} else if timeOfDay > DayEnd {
		sunlightFactor = 1.0 - (timeOfDay - DayEnd) / (SunsetStart - DayEnd)
	}
	
	// Use pre-calculated mountain colors
	mountainColors := []color.RGBA{
		{85, 85, 85, 255},    // Slate gray
		{102, 92, 84, 255},   // Warm gray
		{112, 128, 144, 255}, // Slate blue
	}
	
	// Get base mountain color (reduced random calls)
	baseColor = mountainColors[int(height/100)%len(mountainColors)]
	
	// Simplified color calculations
	heightFactor := height / maxHeight * 0.2
	r := uint8(float64(baseColor.R) * (1 + heightFactor))
	g := uint8(float64(baseColor.G) * (1 + heightFactor))
	b := uint8(float64(baseColor.B) * (1 + heightFactor))
	
	// Add snow effect
	if snowAmount > 0 {
		r = uint8(float64(r)*(1-snowAmount) + 245*snowAmount)
		g = uint8(float64(g)*(1-snowAmount) + 245*snowAmount)
		b = uint8(float64(b)*(1-snowAmount) + 250*snowAmount)
	}
	
	// Add sunlight (simplified)
	if sunlightFactor > 0 {
		sunFactor := sunlightFactor * 0.2
		r = uint8(math.Min(255, float64(r)*(1+sunFactor)))
		g = uint8(math.Min(255, float64(g)*(1+sunFactor)))
		b = uint8(math.Min(255, float64(b)*(1+sunFactor)))
	}
	
	return color.RGBA{r, g, b, 255}
}

// Add smoothstep function for better interpolation
func smoothstep(x float64) float64 {
	// Clamp between 0 and 1
	if x < 0 {
		x = 0
	}
	if x > 1 {
		x = 1
	}
	// Smooth interpolation curve
	return x * x * (3 - 2*x)
}