# Doodle Jump Clone

A simple Doodle Jump style game written in Go using the Ebitengine game library.

## Controls

- Left Arrow / A: Move left
- Right Arrow / D: Move right
- N: Toggle night mode manually
- W: Cycle through weather types (Clear, Rain, Snow)
- Space: Restart after game over

## How to Play

- Control the green character to jump on platforms and climb higher
- Avoid the bird obstacles or you'll lose
- If you fall below the screen, the game ends
- Your score increases as you climb higher
- Press Space to restart after game over

## Features

- Character sprite that changes direction based on movement
- Platforms to jump on
- Bird obstacles that move horizontally
- Dynamic floating clouds in the background
- Day/night cycle that changes the game's appearance
- Weather system with clear, rain and snow conditions
- Score tracking
- Game over state with restart option

## Visual Elements

- **Day/Night Cycle**: The game automatically cycles between day and night modes, changing the background color and adjusting the brightness of all game elements.
- **Cloud System**: Clouds with varying sizes and transparency float across the sky.
- **Weather Effects**:
  - Clear: Standard gameplay with no precipitation
  - Rain: Blue raindrops fall from the sky
  - Snow: White snowflakes drift down slowly

## Building and Running

```bash
# Build the game
go build

# Run the game
./
```

## Requirements

- Go 1.16 or higher
- Ebitengine library (installed automatically with go mod tidy)
