package main

import (
	"log"
	
	"doodlejump/game"
	
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowSize(game.ScreenWidth*2, game.ScreenHeight*2)
	ebiten.SetWindowTitle("Doodle Jump")
	
	if err := ebiten.RunGame(game.NewGame()); err != nil {
		log.Fatal(err)
	}
}