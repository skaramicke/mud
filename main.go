package main

import (
	"mud/game"
	"mud/telnet"
)

func main() {
	gameInstance := game.NewGame()
	server := telnet.NewServer(gameInstance)
	server.Start()
}