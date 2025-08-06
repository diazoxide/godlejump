# GodleJump

A simple Doodle Jump style game written in Go using the [Ebitengine](https://ebitengine.org/) 2D game library.

## Features

- **Character Movement**: Sprite character that changes direction based on movement
- **Platform Jumping**: Jump on platforms to climb higher and increase your score
- **Obstacles**: Avoid bird enemies that move horizontally across the screen
- **Dynamic Environment**: 
  - Day/night cycle that automatically changes the game's appearance
  - Weather system with clear, rain, and snow conditions
  - Floating clouds with varying sizes and transparency
- **Game Mechanics**: Score tracking and game over state with restart functionality

## Controls

| Key | Action |
|-----|---------|
| `←` / `A` | Move left |
| `→` / `D` | Move right |
| `N` | Toggle night mode manually |
| `W` | Cycle through weather types (Clear, Rain, Snow) |
| `Space` | Restart after game over |

## How to Play

1. Control the green character to jump on platforms and climb higher
2. Avoid the bird obstacles or you'll lose the game
3. If you fall below the screen, the game ends
4. Your score increases as you climb higher
5. Press `Space` to restart after game over

## Installation and Setup

### Prerequisites

- Go 1.23.2 or higher
- Git (for cloning the repository)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/diazoxide/godlejump.git
   cd godlejump
   ```

2. Download dependencies:
   ```bash
   go mod tidy
   ```

3. Build the game:
   ```bash
   go build -o godlejump
   ```

4. Run the game:
   ```bash
   ./godlejump
   ```

### Quick Start

For a one-liner to run the game directly:
```bash
go run main.go
```

## Visual Effects

- **Day/Night Cycle**: Automatically cycles between day and night modes, changing background colors and adjusting brightness of all game elements
- **Weather System**: 
  - **Clear**: Standard gameplay with no precipitation
  - **Rain**: Blue raindrops fall from the sky
  - **Snow**: White snowflakes drift down slowly
- **Dynamic Clouds**: Clouds with varying sizes and transparency float across the sky

## License

This project is available under the terms specified in the [LICENSE](LICENSE) file.
