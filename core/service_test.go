package core_test

import (
	"testing"

	"github.com/Razzle131/dungeon-challenge/config"
	. "github.com/Razzle131/dungeon-challenge/core"
	mock_core "github.com/Razzle131/dungeon-challenge/core/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

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
			eventTime int
		}
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name: "OK",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{{}}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				player := Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{{}}}
				player.CurLevel = 0
				player.Stats.Status = StatusInRun
				player.Levels[0].IsFirstEntry = false
				player.Levels[0].FirstEntered = 0

				r.EXPECT().UpdatePlayer(t.Context(), player).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Double entry",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Disqualified",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name: "Must be disqualified",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}}, eventTime: 0},
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
			eventTime int
		}
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name: "OK",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{Monsters: 1}}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{Monsters: 0, IsFinished: true}}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "No more monsters",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{Monsters: 0}}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Boss floor",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{IsBossLevel: true}}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Outside of dungeon",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{{}}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Disqualified",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name: "Must be disqualified",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}}, eventTime: 0},
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
			eventTime int
		}
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name: "OK",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{Monsters: 1}, {Monsters: 1, IsFirstEntry: true}}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, CurLevel: 1, Levels: []Level{{Monsters: 1}, {Monsters: 1}}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Empty floor",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{Monsters: 1}, {Monsters: 0, IsFirstEntry: true}}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, CurLevel: 1, Levels: []Level{{Monsters: 1}, {Monsters: 0, IsFinished: true}}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "No more levels",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{}}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Outside of dungeon",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{{}}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Disqualified",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name: "Must be disqualified",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}}, eventTime: 0},
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
		name          string
		input         Player
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name:  "OK",
			input: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, CurLevel: 1, Levels: []Level{{Monsters: 1}, {Monsters: 1}}},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, CurLevel: 0, Levels: []Level{{Monsters: 1}, {Monsters: 1}}}).Return(nil)
			},
			expected: nil,
		},
		{
			name:          "No more levels",
			input:         Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{}}},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name:          "Outside of dungeon",
			input:         Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{{}}},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name:          "Disqualified",
			input:         Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name:  "Must be disqualified",
			input: Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}},
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

			res := service.PlayerMovesPrevFloor(t.Context(), testCase.input, 0)

			assert.Equal(t, testCase.expected, res)
		})
	}
}

func TestPlayerEntersBoss(t *testing.T) {
	testCases := []struct {
		name          string
		input         Player
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name:  "OK",
			input: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{}}},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{IsBossLevel: true}}}).Return(nil)
			},
			expected: nil,
		},
		{
			name:          "Outside of dungeon",
			input:         Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{{}}},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name:          "Disqualified",
			input:         Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name:  "Must be disqualified",
			input: Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}},
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

			res := service.PlayerEntersBoss(t.Context(), testCase.input, 0)

			assert.Equal(t, testCase.expected, res)
		})
	}
}

func TestPlayerKilledBoss(t *testing.T) {
	testCases := []struct {
		name  string
		input struct {
			player    Player
			eventTime int
		}
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name: "OK",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{IsBossLevel: true}}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{IsBossLevel: true, IsFinished: true}}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Finished floor",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{IsBossLevel: true, IsFinished: true}}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Not boss floor",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{}}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Outside of dungeon",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{{}}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Disqualified",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name: "Must be disqualified",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}}, eventTime: 0},
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
			eventTime int
		}
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name: "OK success",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{IsFinished: true}}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusSuccess}, Levels: []Level{{IsFinished: true}}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "OK failure",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{{}}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusFail}, Levels: []Level{{}}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Outside of dungeon",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{{}}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Disqualified",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name: "Must be disqualified",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}}, eventTime: 0},
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
			eventTime int
		}
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name: "OK",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}, Levels: []Level{}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}, Levels: []Level{}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "OK outside of dungeon",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}, Levels: []Level{}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}, Levels: []Level{}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Finished dungeon",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusFail}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Disqualified",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}, eventTime: 0},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name: "Must be disqualified",
			input: struct {
				player    Player
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}}, eventTime: 0},
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
			player Player
			amount int
		}
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name: "OK",
			input: struct {
				player Player
				amount int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}}, amount: MaxHealth - 1},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusInRun, Hp: MaxHealth - 1}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "OK overheal",
			input: struct {
				player Player
				amount int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}}, amount: MaxHealth + 1},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusInRun, Hp: MaxHealth}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Negative heal",
			input: struct {
				player Player
				amount int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}}, amount: -10},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Outside of dungeon",
			input: struct {
				player Player
				amount int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}}},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Disqualified",
			input: struct {
				player Player
				amount int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name: "Must be disqualified",
			input: struct {
				player Player
				amount int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}}},
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

			res := service.PlayerHealed(t.Context(), testCase.input.player, testCase.input.amount, 0)

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
			eventTime int
		}
		repoBehaviour func(r *mock_core.MockPlayerRepository)
		expected      error
	}{
		{
			name: "OK",
			input: struct {
				player    Player
				amount    int
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun, Hp: 100}}, amount: 80},
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
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun, Hp: 100}}, amount: 100},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {
				r.EXPECT().UpdatePlayer(t.Context(), Player{Id: 1, Stats: PlayerStats{Status: StatusFail}}).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Negative amount",
			input: struct {
				player    Player
				amount    int
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusInRun}}, amount: -10},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Outside of dungeon",
			input: struct {
				player    Player
				amount    int
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusRegistered}}},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrImpossibleMove,
		},
		{
			name: "Disqualified",
			input: struct {
				player    Player
				amount    int
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusDisqual}}},
			repoBehaviour: func(r *mock_core.MockPlayerRepository) {},
			expected:      ErrPlayerIsDisqualified,
		},
		{
			name: "Must be disqualified",
			input: struct {
				player    Player
				amount    int
				eventTime int
			}{player: Player{Id: 1, Stats: PlayerStats{Status: StatusNewborn}}, amount: 0},
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
