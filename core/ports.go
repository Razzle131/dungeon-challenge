package core

import (
	"context"
)

type EventReader interface {
	GetNextEvent() (Event, error)
}

type PlayerRepository interface {
	AddPlayer(ctx context.Context, p Player) (Player, error)
	UpdatePlayer(ctx context.Context, p Player) error
	GetPlayerById(ctx context.Context, id int) (Player, error)
	GetAllPlayers(ctx context.Context) []Player
}

type Printer interface {
	PrintEventRegister(e Event)
	PrintEventEnter(e Event)
	PrintEventKillMonster(e Event)
	PrintEventNextFloor(e Event)
	PrintEventPrevFloor(e Event)
	PrintEventEnterBoss(e Event)
	PrintEventKillBoss(e Event)
	PrintEventLeft(e Event)
	PrintEventCannotContinue(e Event)
	PrintEventHealed(e Event)
	PrintEventDamaged(e Event)
	PrintEventDisqualified(e Event)
	PrintEventDead(e Event)
	PrintEventImpossibleMove(e Event)
	PrintReportLabel()
	PrintReport(info ReportInfo)
}
