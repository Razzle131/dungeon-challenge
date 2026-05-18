package core

const (
	MaxHealth        = 100
	StatusSuccess    = "SUCCESS"
	StatusFail       = "FAIL"
	StatusDisqual    = "DISQUAL"
	StatusInRun      = "INRUN"
	StatusRegistered = "REGISTERED"
	StatusNewborn    = ""
)

const secondsPerDay = 24 * 60 * 60

const (
	eventRegister = iota + 1
	eventEnter
	eventKillMonster
	eventNextFloor
	eventPrevFloor
	eventEnterBoss
	eventKillBoss
	eventLeft
	eventCannotContinue
	eventHealed
	eventDamaged
)
