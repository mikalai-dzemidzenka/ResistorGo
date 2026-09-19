package main

import (
	"flag"
	"resistor"
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
	flag.Parse()
	cfg.Port = uint16(*port)

	if *dedicated {
		resistor.Dedicated(cfg)
		return
	}
	resistor.Game()
}
