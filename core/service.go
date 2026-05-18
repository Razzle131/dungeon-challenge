package core

import (
	"context"
	"errors"
	"log/slog"
	"slices"

	"github.com/Razzle131/dungeon-challenge/config"
)

type Service struct {
	playerRepo  PlayerRepository
	eventReader EventReader
	printer     Printer
	dungeonInfo Dungeon
	logger      *slog.Logger
}

func New(playerRepo PlayerRepository, eventReader EventReader, printer Printer, cfg config.Config) *Service {
	return &Service{
		playerRepo:  playerRepo,
		eventReader: eventReader,
		printer:     printer,
		dungeonInfo: Dungeon{
			OpensAt:  cfg.OpenAt,
			ClosesAt: cfg.OpenAt + int64(cfg.DurationUnix),
			Monsters: cfg.Monsters,
			Floors:   cfg.Floors,
		},
		logger: slog.Default(),
	}
}

const int64minValue = -1<<63 + 1

func (s *Service) ProcessEvents(ctx context.Context) {
	var prevEvent Event
	prevEvent.EventTimeUnix = int64minValue
	for {
		curEvent := s.getNextEvent()
		if curEvent.EventId == 0 {
			break
		}

		for prevEvent.EventTimeUnix > curEvent.EventTimeUnix {
			curEvent.EventTimeUnix += secondsPerDay
		}

		var player Player
		player, err := s.playerRepo.GetPlayerById(ctx, curEvent.PlayerId)
		if errors.Is(err, ErrNotFound) {
			player, err = s.playerRepo.AddPlayer(ctx, NewPlayer(curEvent.PlayerId, s.dungeonInfo))
			if err != nil {
				s.logger.Error("process event add player", "error", err)
			}
		} else if err != nil {
			s.logger.Error("process event get player", "error", err)
		}

		s.processEvent(ctx, curEvent, player)

		prevEvent = curEvent
	}

	s.printer.PrintReportLabel()

	players := s.playerRepo.GetAllPlayers(ctx)
	for _, player := range players {
		s.printer.PrintReport(FormReportInfo(player, s.dungeonInfo))
	}
}

func (s *Service) getNextEvent() Event {
	curEvent, scanned, err := s.eventReader.GetNextEvent()
	for scanned && err != nil {
		curEvent, scanned, err = s.eventReader.GetNextEvent()
	}
	return curEvent
}

func (s *Service) processEvent(ctx context.Context, event Event, player Player) {
	var err error
	switch event.EventId {
	case eventRegister:
		err = withPrint(ctx, player, event, s.RegisterPlayer, s.printer.PrintEventRegister)
	case eventEnter:
		err = withPrint(ctx, player, event, s.PlayerEntersDungeon, s.printer.PrintEventEnter)
	case eventKillMonster:
		err = withPrint(ctx, player, event, s.PlayerKillsMonster, s.printer.PrintEventKillMonster)
	case eventNextFloor:
		err = withPrint(ctx, player, event, s.PlayerMovesNextFloor, s.printer.PrintEventNextFloor)
	case eventPrevFloor:
		err = withPrint(ctx, player, event, s.PlayerMovesPrevFloor, s.printer.PrintEventPrevFloor)
	case eventEnterBoss:
		err = withPrint(ctx, player, event, s.PlayerEntersBoss, s.printer.PrintEventEnterBoss)
	case eventKillBoss:
		err = withPrint(ctx, player, event, s.PlayerKilledBoss, s.printer.PrintEventKillBoss)
	case eventLeft:
		err = withPrint(ctx, player, event, s.PlayerLeftDungeon, s.printer.PrintEventLeft)
	case eventCannotContinue:
		err = withPrint(ctx, player, event, s.PlayerCannotContinue, s.printer.PrintEventCannotContinue)
	case eventHealed:
		amount, ok := event.ExtraParam.(int)
		if !ok {
			break
		}

		err = s.PlayerHealed(ctx, player, amount, event.EventTimeUnix)
		if err == nil {
			s.printer.PrintEventHealed(event)
		}
	case eventDamaged:
		amount, ok := event.ExtraParam.(int)
		if !ok {
			break
		}

		isDead, err := s.PlayerDamaged(ctx, player, amount, event.EventTimeUnix)
		if err == nil {
			s.printer.PrintEventDamaged(event)
		}
		if isDead {
			s.printer.PrintEventDead(event)
		}
	}

	if errors.Is(err, ErrImpossibleMove) {
		s.printer.PrintEventImpossibleMove(event)
	}

	if errors.Is(err, ErrPlayerDisqualified) {
		s.printer.PrintEventDisqualified(event)
	}
}

func withPrint(ctx context.Context, player Player, e Event, call func(ctx context.Context, player Player, eventTime int64) error, printCall func(e Event)) error {
	err := call(ctx, player, e.EventTimeUnix)
	if err != nil {
		return err

	}
	printCall(e)
	return nil
}

func (s *Service) RegisterPlayer(ctx context.Context, player Player, eventTime int64) error {
	if player.Stats.Status == StatusDisqual {
		return ErrPlayerIsDisqualified
	}

	if player.Stats.Status != StatusNewborn {
		return ErrImpossibleMove
	}

	player.Stats.Status = StatusRegistered
	err := s.playerRepo.UpdatePlayer(ctx, player)
	if errors.Is(err, ErrNotFound) {
		s.DisqualAndWrite(ctx, player)
		return ErrPlayerDisqualified
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) PlayerEntersDungeon(ctx context.Context, player Player, eventTime int64) error {
	err := s.commonChecks(ctx, player, eventTime, []string{StatusRegistered})
	if err != nil {
		return err
	}

	player.CurLevel = 0
	player.Stats.Status = StatusInRun
	player.Levels[0].IsFirstEntry = false
	player.Levels[0].FirstEntered = eventTime
	player.Stats.TotalTime = eventTime
	if player.Levels[0].Monsters <= 0 {
		player.Levels[0].IsFinished = true
		player.Levels[0].FinishedAt = eventTime
	}

	err = s.playerRepo.UpdatePlayer(ctx, player)
	if errors.Is(err, ErrNotFound) {
		s.DisqualAndWrite(ctx, player)
		return ErrPlayerDisqualified
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) PlayerKillsMonster(ctx context.Context, player Player, eventTime int64) error {
	err := s.commonChecks(ctx, player, eventTime, []string{StatusInRun})
	if err != nil {
		return err
	}

	if player.Levels[player.CurLevel].IsBossLevel {
		return ErrImpossibleMove
	}

	if player.Levels[player.CurLevel].Monsters <= 0 {
		return ErrImpossibleMove
	}

	player.Levels[player.CurLevel].Monsters--
	if player.Levels[player.CurLevel].Monsters <= 0 {
		player.Levels[player.CurLevel].FinishedAt = eventTime
		player.Levels[player.CurLevel].IsFinished = true
	}

	err = s.playerRepo.UpdatePlayer(ctx, player)
	if errors.Is(err, ErrNotFound) {
		s.DisqualAndWrite(ctx, player)
		return ErrPlayerDisqualified
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) PlayerMovesNextFloor(ctx context.Context, player Player, eventTime int64) error {
	err := s.commonChecks(ctx, player, eventTime, []string{StatusInRun})
	if err != nil {
		return err
	}

	if player.CurLevel >= len(player.Levels)-1 {
		return ErrImpossibleMove
	}

	player.CurLevel++
	if player.Levels[player.CurLevel].IsFirstEntry {
		player.Levels[player.CurLevel].FirstEntered = eventTime
		player.Levels[player.CurLevel].IsFirstEntry = false
	}

	if player.Levels[player.CurLevel].Monsters <= 0 && !player.Levels[player.CurLevel].IsFinished && !player.Levels[player.CurLevel].IsBossLevel {
		player.Levels[player.CurLevel].FinishedAt = eventTime
		player.Levels[player.CurLevel].IsFinished = true
	}

	err = s.playerRepo.UpdatePlayer(ctx, player)
	if errors.Is(err, ErrNotFound) {
		s.DisqualAndWrite(ctx, player)
		return ErrPlayerDisqualified
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) PlayerMovesPrevFloor(ctx context.Context, player Player, eventTime int64) error {
	err := s.commonChecks(ctx, player, eventTime, []string{StatusInRun})
	if err != nil {
		return err
	}

	if player.CurLevel <= 0 {
		return ErrImpossibleMove
	}

	player.CurLevel--
	err = s.playerRepo.UpdatePlayer(ctx, player)
	if errors.Is(err, ErrNotFound) {
		s.DisqualAndWrite(ctx, player)
		return ErrPlayerDisqualified
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) PlayerEntersBoss(ctx context.Context, player Player, eventTime int64) error {
	err := s.commonChecks(ctx, player, eventTime, []string{StatusInRun})
	if err != nil {
		return err
	}

	if !player.Levels[player.CurLevel].IsBossLevel {
		player.Levels[player.CurLevel].IsBossLevel = true
		player.Levels[player.CurLevel].IsFinished = false
		player.Levels[player.CurLevel].Monsters = 0
	}

	err = s.playerRepo.UpdatePlayer(ctx, player)
	if errors.Is(err, ErrNotFound) {
		s.DisqualAndWrite(ctx, player)
		return ErrPlayerDisqualified
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) PlayerKilledBoss(ctx context.Context, player Player, eventTime int64) error {
	err := s.commonChecks(ctx, player, eventTime, []string{StatusInRun})
	if err != nil {
		return err
	}

	if !player.Levels[player.CurLevel].IsBossLevel {
		return ErrImpossibleMove
	}

	if player.Levels[player.CurLevel].IsFinished {
		return ErrImpossibleMove
	}

	player.Levels[player.CurLevel].FinishedAt = eventTime
	player.Levels[player.CurLevel].IsFinished = true

	err = s.playerRepo.UpdatePlayer(ctx, player)
	if errors.Is(err, ErrNotFound) {
		s.DisqualAndWrite(ctx, player)
		return ErrPlayerDisqualified
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) PlayerLeftDungeon(ctx context.Context, player Player, eventTime int64) error {
	err := s.commonChecks(ctx, player, eventTime, []string{StatusInRun})
	if err != nil {
		return err
	}

	player.Stats.Status = StatusSuccess
	for _, level := range player.Levels {
		if !level.IsFinished {
			player.Stats.Status = StatusFail
			break
		}
	}

	player.Stats.TotalTime = eventTime - player.Stats.TotalTime

	err = s.playerRepo.UpdatePlayer(ctx, player)
	if errors.Is(err, ErrNotFound) {
		s.DisqualAndWrite(ctx, player)
		return ErrPlayerDisqualified
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) PlayerCannotContinue(ctx context.Context, player Player, eventTime int64) error {
	err := s.commonChecks(ctx, player, s.dungeonInfo.OpensAt, []string{StatusInRun, StatusRegistered})
	if err != nil {
		return err
	}

	if player.Stats.Status == StatusInRun {
		player.Stats.TotalTime = eventTime - player.Stats.TotalTime
	}
	player.Stats.Status = StatusDisqual

	err = s.playerRepo.UpdatePlayer(ctx, player)
	if errors.Is(err, ErrNotFound) {
		s.DisqualAndWrite(ctx, player)
		return ErrPlayerDisqualified
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) PlayerHealed(ctx context.Context, player Player, amount int, eventTime int64) error {
	err := s.commonChecks(ctx, player, eventTime, []string{StatusInRun})
	if err != nil {
		return err
	}

	if amount < 0 {
		return ErrImpossibleMove
	}

	player.Stats.Hp = min(player.Stats.Hp+amount, MaxHealth)

	err = s.playerRepo.UpdatePlayer(ctx, player)
	if errors.Is(err, ErrNotFound) {
		s.DisqualAndWrite(ctx, player)
		return ErrPlayerDisqualified
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) PlayerDamaged(ctx context.Context, player Player, amount int, eventTime int64) (bool, error) {
	err := s.commonChecks(ctx, player, eventTime, []string{StatusInRun})
	if err != nil {
		return false, err
	}

	if amount < 0 {
		return false, ErrImpossibleMove
	}

	isDead := false
	player.Stats.Hp = max(player.Stats.Hp-amount, 0)
	if player.Stats.Hp <= 0 {
		player.Stats.Status = StatusFail
		isDead = true
		player.Stats.TotalTime = eventTime - player.Stats.TotalTime
	}

	err = s.playerRepo.UpdatePlayer(ctx, player)
	if errors.Is(err, ErrNotFound) {
		s.DisqualAndWrite(ctx, player)
		return false, ErrPlayerDisqualified
	}
	if err != nil {
		return false, err
	}

	return isDead, nil
}

func (s *Service) DisqualAndWrite(ctx context.Context, p Player) {
	p.Stats.Status = StatusDisqual
	err := s.playerRepo.UpdatePlayer(ctx, p)
	if err != nil {
		s.logger.Error("disqualify", "error", err)
	}
}

func (s *Service) commonChecks(ctx context.Context, player Player, eventTime int64, neededStatuses []string) error {
	if player.Stats.Status == StatusDisqual {
		return ErrPlayerIsDisqualified
	}

	if player.Stats.Status == StatusNewborn {
		s.DisqualAndWrite(ctx, player)
		return ErrPlayerDisqualified
	}

	if s.dungeonInfo.OpensAt > eventTime || eventTime > s.dungeonInfo.ClosesAt {
		return ErrImpossibleMove
	}

	if !slices.Contains(neededStatuses, player.Stats.Status) {
		return ErrImpossibleMove
	}

	return nil
}

func FormReportInfo(p Player, info Dungeon) ReportInfo {
	if p.Stats.Status == StatusInRun {
		p.Stats.TotalTime = info.ClosesAt - p.Stats.TotalTime
	}

	if p.Stats.Status != StatusSuccess && p.Stats.Status != StatusFail {
		p.Stats.Status = StatusDisqual
	}

	res := ReportInfo{
		Status:    p.Stats.Status,
		Id:        p.Id,
		TotalTime: p.Stats.TotalTime,
		AvgTime:   0,
		BossTime:  0,
		Hp:        p.Stats.Hp,
	}

	var avgTime int64
	floorsCompleted := 0
	for _, level := range p.Levels {
		if level.IsFinished && !level.IsBossLevel {
			avgTime += level.FinishedAt - level.FirstEntered
			floorsCompleted++
		}
		if level.IsFinished && level.IsBossLevel {
			res.BossTime = level.FinishedAt - level.FirstEntered
		}
	}

	if floorsCompleted > 0 {
		res.AvgTime = avgTime / int64(floorsCompleted)
	}

	return res
}
