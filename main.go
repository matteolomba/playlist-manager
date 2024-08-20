package main

import (
	"playlist-manager/internal/config"
	"playlist-manager/internal/spotify"
	log "playlist-manager/pkg/logger"
	"playlist-manager/pkg/terminal"
)

func init() {
	log.Init(config.Envs.LogLevel)
	log.Info("Logger: inizializzato")
	spotify.Init()
}

func main() {
	//-> Terminale
	err := terminal.Display()
	if err != nil {
		log.Fatal(err)
	}
}
