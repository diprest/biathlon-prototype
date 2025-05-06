package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

// Config
type Config struct {
	Laps        int
	LapLen      int
	PenaltyLen  int
	FiringLines int
	Start       time.Time
	StartDelta  time.Duration
}

func LoadConfig(path string) (Config, error) {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var raw struct {
		Laps        int    `json:"laps"`
		LapLen      int    `json:"lapLen"`
		PenaltyLen  int    `json:"penaltyLen"`
		FiringLines int    `json:"firingLines"`
		Start       string `json:"start"`
		StartDelta  string `json:"startDelta"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return Config{}, err
	}

	start, err := time.Parse("15:04:05", raw.Start)
	if err != nil {
		return Config{}, fmt.Errorf("invalid start time: %w", err)
	}

	parts := strings.Split(raw.StartDelta, ":")
	if len(parts) != 3 {
		return Config{}, fmt.Errorf("invalid startDelta: %s", raw.StartDelta)
	}
	dur, err := time.ParseDuration(
		parts[0] + "h" + parts[1] + "m" + parts[2] + "s",
	)
	if err != nil {
		return Config{}, fmt.Errorf("invalid startDelta: %w", err)
	}

	return Config{
		Laps:        raw.Laps,
		LapLen:      raw.LapLen,
		PenaltyLen:  raw.PenaltyLen,
		FiringLines: raw.FiringLines,
		Start:       start,
		StartDelta:  dur,
	}, nil
}
