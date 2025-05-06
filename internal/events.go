package internal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type RawEvent struct {
	Timestamp  time.Time
	EventID    int
	Competitor int
	Params     []string
}

func LoadEvents(path string) ([]RawEvent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var evs []RawEvent
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		i1 := strings.Index(line, "[")
		i2 := strings.Index(line, "]")
		if i1 < 0 || i2 < 0 || i2 <= i1 {
			return nil, fmt.Errorf("invalid event line: %s", line)
		}
		tsStr := line[i1+1 : i2]
		ts, err := time.Parse("15:04:05.000", tsStr)
		if err != nil {
			return nil, fmt.Errorf("invalid timestamp: %w", err)
		}
		fields := strings.Fields(line[i2+1:])
		if len(fields) < 2 {
			return nil, fmt.Errorf("invalid event fields: %s", line)
		}

		var e RawEvent
		e.Timestamp = ts
		fmt.Sscanf(fields[0], "%d", &e.EventID)
		fmt.Sscanf(fields[1], "%d", &e.Competitor)
		if len(fields) > 2 {
			e.Params = fields[2:]
		}
		evs = append(evs, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return evs, nil
}
