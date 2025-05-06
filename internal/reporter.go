package internal

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// ReportLines формирует отчет в соответствии с README.
func (p *Processor) ReportLines() []string {
	type entry struct {
		id      int
		status  string
		total   time.Duration
		laps    []string
		pens    []string
		hits    int
		shots   int
		sortKey float64
	}

	var reps []entry
	// Собираем данные
	for id, comp := range p.competitors {
		var e entry
		e.id = id
		e.shots = p.cfg.Laps * p.cfg.FiringLines * 5
		e.hits = comp.Hits

		// круги
		prev := comp.ActualStart
		for i := 0; i < p.cfg.Laps; i++ {
			if i < len(comp.LapFinishes) && !prev.IsZero() {
				d := comp.LapFinishes[i].Sub(prev)
				sp := float64(p.cfg.LapLen) / d.Seconds()
				e.laps = append(e.laps, fmt.Sprintf("{%s, %.3f}", formatDur(d), sp))
				prev = comp.LapFinishes[i]
			} else {
				e.laps = append(e.laps, "{,}")
			}
		}

		// штрафные круги
		for i := range comp.PenaltyEntry {
			if i < len(comp.PenaltyExit) {
				d := comp.PenaltyExit[i].Sub(comp.PenaltyEntry[i])
				sp := float64(p.cfg.PenaltyLen) / d.Seconds()
				e.pens = append(e.pens, fmt.Sprintf("{%s, %.3f}", formatDur(d), sp))
			}
		}

		// статус и общее время
		if comp.Status == "Finished" && len(comp.LapFinishes) >= p.cfg.Laps {
			end := comp.LapFinishes[p.cfg.Laps-1]
			e.total = end.Sub(comp.ScheduledStart)
			e.status = fmt.Sprintf("[%s]", comp.Status)
			e.sortKey = e.total.Seconds()
		} else if comp.Status == "NotStarted" {
			e.status = fmt.Sprintf("[%s]", comp.Status)
			e.sortKey = 1e18
		} else {
			e.status = fmt.Sprintf("[%s]", comp.Status)
			e.sortKey = 1e19
		}

		reps = append(reps, e)
	}

	// сортировка по времени, затем по ID
	sort.Slice(reps, func(i, j int) bool {
		a, b := reps[i], reps[j]
		if a.sortKey != b.sortKey {
			return a.sortKey < b.sortKey
		}
		return a.id < b.id
	})

	// Формируем финальные строки
	var lines []string
	for _, e := range reps {
		lapsStr := "[" + strings.Join(e.laps, ", ") + "]"
		pensStr := "[" + strings.Join(e.pens, ", ") + "]"
		if e.status == "[Finished]" {
			lines = append(lines, fmt.Sprintf("%s %d %s %s %s %d/%d",
				e.status, e.id, formatDur(e.total), lapsStr, pensStr, e.hits, e.shots))
		} else {
			lines = append(lines, fmt.Sprintf("%s %d %s %s %d/%d",
				e.status, e.id, lapsStr, pensStr, e.hits, e.shots))
		}
	}
	return lines
}

// formatDur форматирует длительность в MM:SS.mmm или HH:MM:SS.mmm
func formatDur(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	ms := int(d.Milliseconds()) % 1000
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
	}
	return fmt.Sprintf("%02d:%02d.%03d", m, s, ms)
}
