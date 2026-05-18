package core

import "errors"

var (
	ErrAddedAlready = errors.New("added already")
	ErrNotFound     = errors.New("not found")
)

var (
	ErrNoMoreEvents = errors.New("no more events")
)

var (
	ErrImpossibleMove       = errors.New("player makes imposible move")
	ErrPlayerDisqualified   = errors.New("player must be disqualified")
	ErrPlayerIsDisqualified = errors.New("player is disqualified")
	ErrPlayerDead           = errors.New("played died")
)
