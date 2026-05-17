package core

type Player struct {
	Id       int
	CurLevel int
	Levels   []Level
	Stats    PlayerStats
}

type Level struct {
	FirstEntered int
	FinishedAt   int
	Monsters     int
	IsBossLevel  bool
	IsFirstEntry bool
	IsFinished   bool
}

type PlayerStats struct {
	Hp        int
	Status    string
	TotalTime int
	AvgTime   int
	BossTime  int
}

func NewPlayer(id int, info dungeon) Player {
	res := Player{
		Id:       id,
		CurLevel: -1,
		Levels:   make([]Level, info.floors),
		Stats: PlayerStats{
			Hp:        MaxHealth,
			Status:    StatusNewborn,
			TotalTime: 0,
			AvgTime:   0,
			BossTime:  0,
		},
	}

	for i := 0; i < len(res.Levels); i++ {
		res.Levels[i].Monsters = info.monsters
		res.Levels[i].IsFirstEntry = true
	}

	return res
}

type Event struct {
	EventTime     string
	EventTimeUnix int
	PlayerId      int
	EventId       int
	ExtraParam    any
}

type dungeon struct {
	opensAt  int
	closesAt int
	monsters int
	floors   int
}

type ReportInfo struct {
	Status    string
	Id        int
	TotalTime int
	AvgTime   int
	BossTime  int
	Hp        int
}
