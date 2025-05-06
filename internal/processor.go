package internal

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type competitor struct {
	RegisteredAt   time.Time
	ScheduledStart time.Time
	ActualStart    time.Time
	LapFinishes    []time.Time
	PenaltyEntry   []time.Time
	PenaltyExit    []time.Time
	Hits           int
	Status         string
}

// Processor хранит состояние всех участников и формирует логи
type Processor struct {
	cfg         Config
	competitors map[int]*competitor
	logLines    []string
}

// NewProcessor создаёт новый процессор с заданной конфигурацией
func NewProcessor(cfg Config) *Processor {
	return &Processor{
		cfg:         cfg,
		competitors: make(map[int]*competitor),
	}
}

// Process обрабатывает входящие события и дописывает строки в лог
func (p *Processor) Process(events []RawEvent) {
	for _, e := range events {
		id := e.Competitor
		comp, ok := p.competitors[id]
		if !ok {
			comp = &competitor{}
			p.competitors[id] = comp
		}
		ts := e.Timestamp

		switch e.EventID {
		case 1:
			// регистрация участника
			comp.RegisteredAt = ts
			p.logLines = append(p.logLines,
				fmt.Sprintf("[%s] The competitor(%d) registered",
					ts.Format("15:04:05.000"), id))

		case 2:
			// жребий: назначено запланированное время старта
			st, _ := time.Parse("15:04:05.000", e.Params[0])
			comp.ScheduledStart = st
			p.logLines = append(p.logLines,
				fmt.Sprintf("[%s] The start time for the competitor(%d) was set by a draw to %s",
					ts.Format("15:04:05.000"), id, st.Format("15:04:05.000")))

		case 3:
			// выход на стартовую линию
			p.logLines = append(p.logLines,
				fmt.Sprintf("[%s] The competitor(%d) is on the start line",
					ts.Format("15:04:05.000"), id))

		case 4:
			// фактический старт
			comp.ActualStart = ts
			p.logLines = append(p.logLines,
				fmt.Sprintf("[%s] The competitor(%d) has started",
					ts.Format("15:04:05.000"), id))

		case 5:
			// вход на огневой рубеж
			line := e.Params[0]
			p.logLines = append(p.logLines,
				fmt.Sprintf("[%s] The competitor(%d) is on the firing range(%s)",
					ts.Format("15:04:05.000"), id, line))

		case 6:
			// попадание в мишень
			comp.Hits++
			p.logLines = append(p.logLines,
				fmt.Sprintf("[%s] The target(%s) has been hit by competitor(%d)",
					ts.Format("15:04:05.000"), e.Params[0], id))

		case 7:
			// выход с огневого рубежа
			p.logLines = append(p.logLines,
				fmt.Sprintf("[%s] The competitor(%d) left the firing range",
					ts.Format("15:04:05.000"), id))

		case 8:
			// вход в штрафные круги
			comp.PenaltyEntry = append(comp.PenaltyEntry, ts)
			p.logLines = append(p.logLines,
				fmt.Sprintf("[%s] The competitor(%d) entered the penalty laps",
					ts.Format("15:04:05.000"), id))

		case 9:
			// выход из штрафных кругов
			comp.PenaltyExit = append(comp.PenaltyExit, ts)
			p.logLines = append(p.logLines,
				fmt.Sprintf("[%s] The competitor(%d) left the penalty laps",
					ts.Format("15:04:05.000"), id))

		case 10:
			// финиш круга
			comp.LapFinishes = append(comp.LapFinishes, ts)
			p.logLines = append(p.logLines,
				fmt.Sprintf("[%s] The competitor(%d) ended the main lap",
					ts.Format("15:04:05.000"), id))

		case 11:
			// отказ от продолжения
			msg := strings.Join(e.Params, " ")
			comp.Status = "NotFinished"
			p.logLines = append(p.logLines,
				fmt.Sprintf("[%s] The competitor(%d) can't continue: %s",
					ts.Format("15:04:05.000"), id, msg))
		}
	}

	// Автоназначение ScheduledStart из конфига для тех, кто не получил жребием
	type regEntry struct {
		id int
		at time.Time
	}
	var toSched []regEntry
	for id, comp := range p.competitors {
		if comp.ScheduledStart.IsZero() && !comp.RegisteredAt.IsZero() {
			toSched = append(toSched, regEntry{id, comp.RegisteredAt})
		}
	}
	sort.Slice(toSched, func(i, j int) bool {
		return toSched[i].at.Before(toSched[j].at)
	})
	for idx, e := range toSched {
		comp := p.competitors[e.id]
		comp.ScheduledStart = p.cfg.Start.Add(time.Duration(idx) * p.cfg.StartDelta)
		p.logLines = append(p.logLines,
			fmt.Sprintf("[%s] The start time for the competitor(%d) was scheduled to %s",
				comp.ScheduledStart.Format("15:04:05.000"), e.id, comp.ScheduledStart.Format("15:04:05.000")))
	}

	// Проверка статусов и генерация исходящих (дисквалификация/финиш)
	for id, comp := range p.competitors {
		if comp.ActualStart.IsZero() {
			comp.Status = "NotStarted"
			p.logLines = append(p.logLines,
				fmt.Sprintf("[%s] The competitor(%d) disqualified (no start)",
					comp.ScheduledStart.Format("15:04:05.000"), id))
		} else if comp.Status == "" && len(comp.LapFinishes) == p.cfg.Laps {
			comp.Status = "Finished"
			end := comp.LapFinishes[len(comp.LapFinishes)-1]
			p.logLines = append(p.logLines,
				fmt.Sprintf("[%s] The competitor(%d) finished",
					end.Format("15:04:05.000"), id))
		}
	}
}

// LogLines возвращает сформированный лог (не гарантируется сортировка —
// если нужна строгая хронология, можно отсортировать по префиксу времени)
func (p *Processor) LogLines() []string {
	return p.logLines
}
