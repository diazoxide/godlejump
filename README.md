# GodleJump

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue.svg)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A simple Doodle Jump-style endless jumping game written in Go using the [Ebitengine](https://ebitengine.org/) 2D game library. Inspired by the classic mobile game, this implementation features modern visual effects and smooth gameplay.

## Features

- **Smooth Character Movement**: Sprite character with directional animations
- **Endless Platforming**: Jump on dynamically generated platforms to climb higher
- **Bird Obstacles**: Avoid moving bird enemies that patrol horizontally
- **Dynamic Visual Effects**: 
  - Automatic day/night cycle with smooth color transitions
  - Weather system supporting clear, rain, and snow conditions
  - Animated floating clouds with varying opacity
- **Game Mechanics**: 
  - Real-time score tracking based on height achieved
  - Game over detection with instant restart capability
  - Responsive controls with keyboard input

## Controls

| Key | Action |
|-----|---------|
| `←` / `A` | Move left |
| `→` / `D` | Move right |
| `N` | Manually toggle night mode |
| `W` | Cycle weather (Clear → Rain → Snow) |
| `Space` | Restart after game over |

## How to Play

1. **Objective**: Control your character to jump on platforms and climb as high as possible
2. **Movement**: Use arrow keys or WASD to move left and right
3. **Avoid Obstacles**: Don't touch the bird enemies or you'll lose
4. **Scoring**: Your score increases based on the maximum height reached
5. **Game Over**: Falls below the screen boundary end the game
6. **Restart**: Press `Space` to immediately start a new game

## Requirements

- **Go**: Version 1.21 or higher
- **Operating System**: Windows, macOS, or Linux
- **Graphics**: OpenGL 2.1 compatible graphics card

## Installation

### Method 1: Clone and Build

```bash
# Clone the repository
git clone https://github.com/diazoxide/godlejump.git
cd godlejump

# Download dependencies
go mod tidy

# Build the executable
go build -o doodlejump .

# Run the game
./doodlejump
```

### Method 2: Direct Run

```bash
# Run directly without building
go run .
```

### Method 3: Install Globally

```bash
# Install to your GOPATH/bin
go install github.com/diazoxide/godlejump@latest

# Run from anywhere (ensure GOPATH/bin is in PATH)
doodlejump
```

## Visual Effects

### Day/Night Cycle
The game automatically transitions between day and night modes, creating a dynamic atmosphere with:
- Gradual background color changes
- Adjusted brightness for all game elements
- Smooth transitions that don't disrupt gameplay

### Weather System
Three distinct weather conditions enhance the visual experience:
- **Clear**: Standard sunny/clear conditions
- **Rain**: Animated blue raindrops falling from the sky
- **Snow**: Gentle white snowflakes drifting downward

### Environmental Elements
- **Clouds**: Multi-layered cloud system with varying sizes and transparency
- **Mountains**: Parallax scrolling background mountains for depth
- **Platforms**: Randomly generated platforms with consistent spacing

## Project Structure

```
godlejump/
├── main.go          # Entry point and window setup
├── game/            # Core game logic
│   ├── game.go      # Main game loop and rendering
│   ├── assets/      # Game assets (sprites, textures)
│   └── player.go    # Player character logic
├── go.mod           # Go module definition
├── go.sum           # Dependency checksums
└── LICENSE          # Apache 2.0 license
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Ebitengine](https://ebitengine.org/), a 2D game engine for Go
- Inspired by the original Doodle Jump mobile game
- Graphics and sprites created specifically for this project
