package printer

import (
	"fmt"

	"github.com/Razzle131/dungeon-challenge/core"
)

type FmtPrinter struct{}

func New() *FmtPrinter {
	return &FmtPrinter{}
}

func (p *FmtPrinter) PrintEventRegister(e core.Event) {
	fmt.Printf("[%s] Player [%v] registered\n", e.EventTime, e.PlayerId)
}

func (p *FmtPrinter) PrintEventEnter(e core.Event) {
	fmt.Printf("[%s] Player [%v] entered the dungeon\n", e.EventTime, e.PlayerId)
}

func (p *FmtPrinter) PrintEventKillMonster(e core.Event) {
	fmt.Printf("[%s] Player [%v] killed the monster\n", e.EventTime, e.PlayerId)
}

func (p *FmtPrinter) PrintEventNextFloor(e core.Event) {
	fmt.Printf("[%s] Player [%v] went to the next floor\n", e.EventTime, e.PlayerId)
}

func (p *FmtPrinter) PrintEventPrevFloor(e core.Event) {
	fmt.Printf("[%s] Player [%v] went to the previous floor\n", e.EventTime, e.PlayerId)
}

func (p *FmtPrinter) PrintEventEnterBoss(e core.Event) {
	fmt.Printf("[%s] Player [%v] entered the boss's floor\n", e.EventTime, e.PlayerId)
}

func (p *FmtPrinter) PrintEventKillBoss(e core.Event) {
	fmt.Printf("[%s] Player [%v] killed the boss\n", e.EventTime, e.PlayerId)
}

func (p *FmtPrinter) PrintEventLeft(e core.Event) {
	fmt.Printf("[%s] Player [%v] left the dungeon\n", e.EventTime, e.PlayerId)
}

func (p *FmtPrinter) PrintEventCannotContinue(e core.Event) {
	fmt.Printf("[%s] Player [%v] cannot continue due to [%s]\n", e.EventTime, e.PlayerId, e.ExtraParam)
}

func (p *FmtPrinter) PrintEventHealed(e core.Event) {
	fmt.Printf("[%s] Player [%v] has restored [%v] of health\n", e.EventTime, e.PlayerId, e.ExtraParam)
}

func (p *FmtPrinter) PrintEventDamaged(e core.Event) {
	fmt.Printf("[%s] Player [%v] recieved [%v] of damage\n", e.EventTime, e.PlayerId, e.ExtraParam)
}

func (p *FmtPrinter) PrintEventDisqualified(e core.Event) {
	fmt.Printf("[%s] Player [%v] is disqualified\n", e.EventTime, e.PlayerId)
}

func (p *FmtPrinter) PrintEventDead(e core.Event) {
	fmt.Printf("[%s] Player [%v] is dead\n", e.EventTime, e.PlayerId)
}

func (p *FmtPrinter) PrintEventImpossibleMove(e core.Event) {
	fmt.Printf("[%s] Player [%v] makes imposible move [%v]\n", e.EventTime, e.PlayerId, e.EventId)
}

func (p *FmtPrinter) PrintReportLabel() {
	fmt.Println("Final report:")
}

func (p *FmtPrinter) PrintReport(info core.ReportInfo) {
	fmt.Printf("[%s] %v [%s, %s, %s] HP:%v\n",
		info.Status,
		info.Id,
		unixToTimeFormat(info.TotalTime),
		unixToTimeFormat(info.AvgTime),
		unixToTimeFormat(info.BossTime),
		info.Hp)
}

const (
	secondsPerHour   = 3600
	secondsPerMinute = 60
)

func unixToTimeFormat(t int64) string {
	hours := t / secondsPerHour
	minutes := (t % secondsPerHour) / secondsPerMinute
	seconds := t % secondsPerMinute

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}
