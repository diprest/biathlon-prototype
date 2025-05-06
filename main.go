package main

import (
	"flag"
	"fmt"
	"log"

	"biathlon-prototype/internal"
)

func main() {
	cfgFile := flag.String("config", "config.json", "path to config.json")
	evFile := flag.String("events", "events.txt", "path to events.txt")
	flag.Parse()

	cfg, err := internal.LoadConfig(*cfgFile)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	events, err := internal.LoadEvents(*evFile)
	if err != nil {
		log.Fatalf("load events: %v", err)
	}

	proc := internal.NewProcessor(cfg)
	proc.Process(events)

	for _, line := range proc.LogLines() {
		fmt.Println(line)
	}
	fmt.Println()

	for _, line := range proc.ReportLines() {
		fmt.Println(line)
	}
}
