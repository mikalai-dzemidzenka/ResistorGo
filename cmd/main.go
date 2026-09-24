package main

import (
	"flag"
	"resistor"
	"strings"
)

func main() {
	cfg := resistor.DefaultServerConfig()
	dedicated := flag.Bool("dedicated", false, "run a headless dedicated server")
	port := flag.Uint("port", resistor.DefaultPort, "server UDP port")
	flag.IntVar(&cfg.MaxPlayers, "maxplayers", cfg.MaxPlayers, "maximum human players")
	flag.IntVar(&cfg.MaxSpectators, "maxspectators", cfg.MaxSpectators, "maximum spectators")
	flag.IntVar(&cfg.Bots, "bots", 0, "number of server bots")
	flag.BoolVar(&cfg.LagComp, "lagcomp", true, "lag compensation for hit tests")
	// Only a development build reads a level off the disk; a release plays the
	// one it was built with, so this does nothing there.
	flag.StringVar(&resistor.LevelPath, "level", resistor.LevelPath, "level file to play (development builds only)")
	flag.StringVar(&resistor.Gamemode, "gamemode", "", "gamemode the server runs, from scripts/gamemodes/<name>.lua")
	filterscripts := flag.String("filterscripts", "", "comma-separated filterscripts the server runs, from scripts/filterscripts/<name>.lua")
	flag.Parse()
	cfg.Port = uint16(*port)
	for _, name := range strings.Split(*filterscripts, ",") {
		if name = strings.TrimSpace(name); name != "" {
			resistor.Filterscripts = append(resistor.Filterscripts, name)
		}
	}
	// read after the flags, like the Host screen's own configuration
	cfg.Gamemode, cfg.Filterscripts = resistor.Gamemode, resistor.Filterscripts

	if *dedicated {
		resistor.Dedicated(cfg)
		return
	}
	resistor.Game()
}
