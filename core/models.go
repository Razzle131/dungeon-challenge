package core

type Player struct {
	Id       int
	CurLevel int
	Levels   []Level
	Stats    PlayerStats
}

type Level struct {
	FirstEntered int64
	FinishedAt   int64
	Monsters     int
	IsBossLevel  bool
	IsFirstEntry bool
	IsFinished   bool
}

type PlayerStats struct {
	Hp        int
	Status    string
	TotalTime int64
	AvgTime   int64
	BossTime  int64
}

func NewPlayer(id int, info Dungeon) Player {
	res := Player{
		Id:       id,
		CurLevel: -1,
		Levels:   make([]Level, info.Floors),
		Stats: PlayerStats{
			Hp:        MaxHealth,
			Status:    StatusNewborn,
			TotalTime: 0,
			AvgTime:   0,
			BossTime:  0,
		},
	}

	for i := 0; i < len(res.Levels); i++ {
		res.Levels[i].Monsters = info.Monsters
		res.Levels[i].IsFirstEntry = true
	}

	return res
}

type Event struct {
	EventTime     string
	EventTimeUnix int64
	PlayerId      int
	EventId       int
	ExtraParam    any
}

type Dungeon struct {
	OpensAt  int64
	ClosesAt int64
	Monsters int
	Floors   int
}

type ReportInfo struct {
	Status    string
	Id        int
	TotalTime int64
	AvgTime   int64
	BossTime  int64
	Hp        int
}
