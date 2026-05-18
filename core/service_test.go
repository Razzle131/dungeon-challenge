package core_test

import (
	"testing"

	"github.com/Razzle131/dungeon-challenge/config"
	. "github.com/Razzle131/dungeon-challenge/core"
	mock_core "github.com/Razzle131/dungeon-challenge/core/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// использовал testify для более удобного вывода результатов тестов, если фейлятся

func TestRegisterPlayer(t *testing.T) {
	testCases := []struct {
		name             string
		input            Player
		repoBehaviour    func(r *mock_core.MockPlayerRepository)
		printerBehaviour func(p *mock_core.MockPrinter)
		expected         error
	}{
		{
			name:  "OK",
			input: Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}}).Return(nil)
			},
			expected: nil,
		},
		{
			name:  "OK closed dungeon",
			input: Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}}).Return(nil)
			},
			expected: nil,
		},
		{
			name:          "Double registration",
			input:         Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name:          "Disqualified",
			input:         Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_core.NewMockPlayerRepository(c)
			testCase.repoBehaviour(repo)

			service := New(repo, nil, nil, config.Config{})

			res := service.RegisterPlayer(t.Context(), testCase.input, 0)

			assert.Equal(t, testCase.expected, res)
		})
	}
}

func TestPlayerEntersDungeon(t *testing.T) {
	testCases := []struct {
		name  string
		input struct {
			player    Player
			eventTime int64
		}
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name: "OK",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{{Monsters: 1, IsFirstEntry: true}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				player := Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{{Monsters: 1}}}
				player.CurLevel = 0
				player.Stats.Status = StatusInRun
				player.Levels[0].IsFirstEntry = false
				player.Levels[0].FirstEntered = 0

				r.EXPECT().UpdatePlayer(t.Context(), player).Return(nil)
			},
			expected: nil,
		},
		{
			name: "OK empty floor",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{{Monsters: 0, IsFirstEntry: true}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				player := Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{{Monsters: 0}}}
				player.CurLevel = 0
				player.Stats.Status = StatusInRun
				player.Levels[0].IsFirstEntry = false
				player.Levels[0].FirstEntered = 0
				player.Levels[0].IsFinished = true
				player.Levels[0].FinishedAt = 0

				r.EXPECT().UpdatePlayer(t.Context(), player).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Dungeon closed",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{{}}},
				eventTime: 1,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Double entry",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Disqualified",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name: "Must be disqualified",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}).Return(nil)
			},
			expected: ErrPlayerDisqualified,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_core.NewMockPlayerRepository(c)
			testCase.repoBehaviour(repo)

			service := New(repo, nil, nil, config.Config{Monsters: 1})

			res := service.PlayerEntersDungeon(t.Context(), testCase.input.player, testCase.input.eventTime)

			assert.Equal(t, testCase.expected, res)
		})
	}
}

func TestPlayerKillsMonster(t *testing.T) {
	testCases := []struct {
		name  string
		input struct {
			player    Player
			eventTime int64
		}
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name: "OK",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{Monsters: 1}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{Monsters: 0, IsFinished: true}}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Dungeon closed",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}},
				eventTime: 1,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "No more monsters",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{Monsters: 0}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Boss floor",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{IsBossLevel: true}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Outside of dungeon",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{{}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Disqualified",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name: "Must be disqualified",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}).Return(nil)
			},
			expected: ErrPlayerDisqualified,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_core.NewMockPlayerRepository(c)
			testCase.repoBehaviour(repo)

			service := New(repo, nil, nil, config.Config{})

			res := service.PlayerKillsMonster(t.Context(), testCase.input.player, testCase.input.eventTime)

			assert.Equal(t, testCase.expected, res)
		})
	}
}

func TestPlayerMovesNextFloor(t *testing.T) {
	testCases := []struct {
		name  string
		input struct {
			player    Player
			eventTime int64
		}
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name: "OK",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{Monsters: 1}, {Monsters: 1, IsFirstEntry: true}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, CurLevel: 1, Levels: []Level{{Monsters: 1}, {Monsters: 1}}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Empty floor",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{Monsters: 1}, {Monsters: 0, IsFirstEntry: true}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, CurLevel: 1, Levels: []Level{{Monsters: 1}, {Monsters: 0, IsFinished: true}}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Dungeon closed",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, CurLevel: 1, Levels: []Level{{}, {}}},
				eventTime: 1,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "No more levels",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Outside of dungeon",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{{}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Disqualified",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name: "Must be disqualified",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}).Return(nil)
			},
			expected: ErrPlayerDisqualified,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_core.NewMockPlayerRepository(c)
			testCase.repoBehaviour(repo)

			service := New(repo, nil, nil, config.Config{})

			res := service.PlayerMovesNextFloor(t.Context(), testCase.input.player, testCase.input.eventTime)

			assert.Equal(t, testCase.expected, res)
		})
	}
}

func TestPlayerMovesPrevFloor(t *testing.T) {
	testCases := []struct {
		name  string
		input struct {
			player    Player
			eventTime int64
		}
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name: "OK",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, CurLevel: 1, Levels: []Level{{Monsters: 1}, {Monsters: 1}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, CurLevel: 0, Levels: []Level{{Monsters: 1}, {Monsters: 1}}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Dungeon closed",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, CurLevel: 1, Levels: []Level{{}, {}}},
				eventTime: 1,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "No more levels",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Outside of dungeon",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{{}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Disqualified",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name: "Must be disqualified",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}).Return(nil)
			},
			expected: ErrPlayerDisqualified,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_core.NewMockPlayerRepository(c)
			testCase.repoBehaviour(repo)

			service := New(repo, nil, nil, config.Config{})

			res := service.PlayerMovesPrevFloor(t.Context(), testCase.input.player, testCase.input.eventTime)

			assert.Equal(t, testCase.expected, res)
		})
	}
}

func TestPlayerEntersBoss(t *testing.T) {
	testCases := []struct {
		name  string
		input struct {
			player    Player
			eventTime int64
		}
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name: "OK",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{IsBossLevel: true}}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Dungeon closed",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}},
				eventTime: 1,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Outside of dungeon",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{{}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Disqualified",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name: "Must be disqualified",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}).Return(nil)
			},
			expected: ErrPlayerDisqualified,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_core.NewMockPlayerRepository(c)
			testCase.repoBehaviour(repo)

			service := New(repo, nil, nil, config.Config{})

			res := service.PlayerEntersBoss(t.Context(), testCase.input.player, testCase.input.eventTime)

			assert.Equal(t, testCase.expected, res)
		})
	}
}

func TestPlayerKilledBoss(t *testing.T) {
	testCases := []struct {
		name  string
		input struct {
			player    Player
			eventTime int64
		}
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name: "OK",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{IsBossLevel: true}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{IsBossLevel: true, IsFinished: true}}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Finished floor",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{IsBossLevel: true, IsFinished: true}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Dungeon closed",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun, Hp: 100}},
				eventTime: 1,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Not boss floor",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Outside of dungeon",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{{}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Disqualified",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name: "Must be disqualified",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}).Return(nil)
			},
			expected: ErrPlayerDisqualified,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_core.NewMockPlayerRepository(c)
			testCase.repoBehaviour(repo)

			service := New(repo, nil, nil, config.Config{})

			res := service.PlayerKilledBoss(t.Context(), testCase.input.player, testCase.input.eventTime)

			assert.Equal(t, testCase.expected, res)
		})
	}
}

func TestPlayerLeftDungeon(t *testing.T) {
	testCases := []struct {
		name  string
		input struct {
			player    Player
			eventTime int64
		}
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name: "OK success",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{IsFinished: true}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusSuccess}, Levels: []Level{{IsFinished: true}}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "OK failure",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusFail}, Levels: []Level{{}}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Dungeon closed",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun, Hp: 100}},
				eventTime: 1,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Outside of dungeon",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{{}}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Disqualified",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name: "Must be disqualified",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}).Return(nil)
			},
			expected: ErrPlayerDisqualified,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_core.NewMockPlayerRepository(c)
			testCase.repoBehaviour(repo)

			service := New(repo, nil, nil, config.Config{})

			res := service.PlayerLeftDungeon(t.Context(), testCase.input.player, testCase.input.eventTime)

			assert.Equal(t, testCase.expected, res)
		})
	}
}

func TestPlayerCannotContinue(t *testing.T) {
	testCases := []struct {
		name  string
		input struct {
			player    Player
			eventTime int64
		}
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name: "OK",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "OK outside of dungeon",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "OK dungeon closed",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}},
				eventTime: -1,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Finished dungeon",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusFail}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Disqualified",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name: "Must be disqualified",
			input: struct {
				player    Player
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}},
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}).Return(nil)
			},
			expected: ErrPlayerDisqualified,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_core.NewMockPlayerRepository(c)
			testCase.repoBehaviour(repo)

			service := New(repo, nil, nil, config.Config{})

			res := service.PlayerCannotContinue(t.Context(), testCase.input.player, testCase.input.eventTime)

			assert.Equal(t, testCase.expected, res)
		})
	}
}

func TestPlayerHealed(t *testing.T) {
	testCases := []struct {
		name  string
		input struct {
			player    Player
			amount    int
			eventTime int64
		}
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name: "OK",
			input: struct {
				player    Player
				amount    int
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}},
				amount:    MaxHealth - 1,
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusInRun, Hp: MaxHealth - 1}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "OK overheal",
			input: struct {
				player    Player
				amount    int
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}},
				amount:    MaxHealth + 1,
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusInRun, Hp: MaxHealth}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Dungeon closed",
			input: struct {
				player    Player
				amount    int
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun, Hp: 100}},
				amount:    0,
				eventTime: -1,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Negative heal",
			input: struct {
				player    Player
				amount    int
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}},
				amount:    -10,
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Outside of dungeon",
			input: struct {
				player    Player
				amount    int
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}},
				amount:    0,
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Disqualified",
			input: struct {
				player    Player
				amount    int
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}},
				amount:    0,
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name: "Must be disqualified",
			input: struct {
				player    Player
				amount    int
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}},
				amount:    0,
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}).Return(nil)
			},
			expected: ErrPlayerDisqualified,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_core.NewMockPlayerRepository(c)
			testCase.repoBehaviour(repo)

			service := New(repo, nil, nil, config.Config{})

			res := service.PlayerHealed(t.Context(), testCase.input.player, testCase.input.amount, testCase.input.eventTime)

			assert.Equal(t, testCase.expected, res)
		})
	}
}

func TestPlayerDamaged(t *testing.T) {
	testCases := []struct {
		name  string
		input struct {
			player    Player
			amount    int
			eventTime int64
		}
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name: "OK",
			input: struct {
				player    Player
				amount    int
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun, Hp: 100}},
				amount:    80,
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusInRun, Hp: 20}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "OK dead",
			input: struct {
				player    Player
				amount    int
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun, Hp: 100}},
				amount:    100,
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusFail}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Dungeon closed",
			input: struct {
				player    Player
				amount    int
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun, Hp: 100}},
				amount:    0,
				eventTime: -1,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Negative amount",
			input: struct {
				player    Player
				amount    int
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}},
				amount:    -10,
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Outside of dungeon",
			input: struct {
				player    Player
				amount    int
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}},
				amount:    0,
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Disqualified",
			input: struct {
				player    Player
				amount    int
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}},
				amount:    0,
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name: "Must be disqualified",
			input: struct {
				player    Player
				amount    int
				eventTime int64
			}{
				player:    Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}},
				amount:    0,
				eventTime: 0,
			},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}).Return(nil)
			},
			expected: ErrPlayerDisqualified,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_core.NewMockPlayerRepository(c)
			testCase.repoBehaviour(repo)

			service := New(repo, nil, nil, config.Config{})

			_, res := service.PlayerDamaged(t.Context(), testCase.input.player, testCase.input.amount, testCase.input.eventTime)

			assert.Equal(t, testCase.expected, res)
		})
	}
}

func TestFormReportInfo(t *testing.T) {
	testCases := []struct {
		name  string
		input struct {
			player Player
			info   Dungeon
		}
		expected ReportInfo
	}{
		{
			name: "OK default",
			input: struct {
				player Player
				info   Dungeon
			}{
				player: Player{Id: 1, Stats: PlayerStats{Status: StatusFail, TotalTime: 20, Hp: 100}, Levels: []Level{{FirstEntered: 0, IsFinished: true, FinishedAt: 10}}},
				info:   Dungeon{},
			},
			expected: ReportInfo{Status: StatusFail, Id: 1, TotalTime: 20, AvgTime: 10, BossTime: 0, Hp: 100},
		},
		{
			name: "OK only boss",
			input: struct {
				player Player
				info   Dungeon
			}{
				player: Player{Id: 1, Stats: PlayerStats{Status: StatusSuccess, TotalTime: 20, Hp: 100}, Levels: []Level{{FirstEntered: 0, IsFinished: true, FinishedAt: 10, IsBossLevel: true}}},
				info:   Dungeon{},
			},
			expected: ReportInfo{Status: StatusSuccess, Id: 1, TotalTime: 20, AvgTime: 0, BossTime: 10, Hp: 100},
		},
		{
			name: "OK still in run",
			input: struct {
				player Player
				info   Dungeon
			}{
				player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun, TotalTime: 0, Hp: 100}, Levels: []Level{{FirstEntered: 0}}},
				info:   Dungeon{ClosesAt: 10},
			},
			expected: ReportInfo{Status: StatusDisqual, Id: 1, TotalTime: 10, AvgTime: 0, BossTime: 0, Hp: 100},
		},
		{
			name: "OK no floors comleted",
			input: struct {
				player Player
				info   Dungeon
			}{
				player: Player{Id: 1, Stats: PlayerStats{Status: StatusFail, TotalTime: 20, Hp: 100}, Levels: []Level{{FirstEntered: 0}}},
				info:   Dungeon{},
			},
			expected: ReportInfo{Status: StatusFail, Id: 1, TotalTime: 20, AvgTime: 0, BossTime: 0, Hp: 100},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			res := FormReportInfo(testCase.input.player, testCase.input.info)
			assert.Equal(t, testCase.expected, res)
		})
	}
}
