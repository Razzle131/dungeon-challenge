package events

import (
	"bufio"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Razzle131/dungeon-challenge/core"
)

type EventFileReader struct {
	file    *os.File
	scanner *bufio.Scanner
	logger  *slog.Logger
}

func New(filePath string) (*EventFileReader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	return &EventFileReader{
		file:    file,
		scanner: bufio.NewScanner(file),
		logger:  slog.Default(),
	}, nil
}

func (e *EventFileReader) CloseFileOrLog() {
	if err := e.file.Close(); err != nil {
		e.logger.Error("closing file", "error", err.Error())
	}
}

func (e *EventFileReader) GetNextEvent() (core.Event, bool, error) {
	scanRes := e.scanner.Scan()
	if !scanRes {
		return core.Event{}, scanRes, core.ErrNoMoreEvents
	}

	line := e.scanner.Text()
	tmp := strings.Split(strings.TrimSpace(line), " ")

	eventTime, err := time.Parse("[15:04:05]", tmp[0])
	if err != nil {
		return core.Event{}, scanRes, err
	}

	playerId, err := strconv.Atoi(tmp[1])
	if err != nil {
		return core.Event{}, scanRes, err
	}

	eventId, err := strconv.Atoi(tmp[2])
	if err != nil {
		return core.Event{}, scanRes, err
	}

	var extraParam any = ""
	if len(tmp) > 3 && eventId == 9 {
		extraParam = tmp[3]
	}
	if len(tmp) > 3 && (eventId == 10 || eventId == 11) {
		extraParam, err = strconv.Atoi(tmp[3])
		if err != nil {
			return core.Event{}, scanRes, err
		}
	}

	return core.Event{
		EventTime:     tmp[0][1 : len(tmp[0])-1],
		EventTimeUnix: eventTime.Unix(),
		PlayerId:      playerId,
		EventId:       eventId,
		ExtraParam:    extraParam,
	}, scanRes, nil
}
