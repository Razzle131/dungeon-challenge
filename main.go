package main

import (
	"context"
	"flag"

	"github.com/Razzle131/dungeon-challenge/adapters/events"
	"github.com/Razzle131/dungeon-challenge/adapters/players"
	"github.com/Razzle131/dungeon-challenge/adapters/printer"
	"github.com/Razzle131/dungeon-challenge/config"
	"github.com/Razzle131/dungeon-challenge/core"
)

func main() {
	var configPath string
	var inputPath string
	flag.StringVar(&configPath, "config", "config.json", "path to configuration file")
	flag.StringVar(&inputPath, "events", "events", "path to events file")
	flag.Parse()

	cfg := config.MustLoad(configPath)

	eventReader, err := events.New(inputPath)
	if err != nil {
		panic(err)
	}

	printer := printer.New()
	playerRepo := players.New()

	service := core.New(playerRepo, eventReader, printer, cfg)

	service.ProcessEvents(context.Background())
}
