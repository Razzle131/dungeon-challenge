package players

import (
	"context"
	"slices"

	"github.com/Razzle131/dungeon-challenge/core"
)

type PlayerCacheRepository struct {
	players map[int]core.Player
}

func New() *PlayerCacheRepository {
	return &PlayerCacheRepository{
		players: make(map[int]core.Player),
	}
}

func (r *PlayerCacheRepository) AddPlayer(ctx context.Context, p core.Player) (core.Player, error) {
	if _, found := r.players[p.Id]; found {
		return core.Player{}, core.ErrAddedAlready
	}

	r.players[p.Id] = p
	return p, nil
}

func (r *PlayerCacheRepository) UpdatePlayer(ctx context.Context, p core.Player) error {
	if _, found := r.players[p.Id]; !found {
		return core.ErrNotFound
	}

	r.players[p.Id] = p
	return nil
}

func (r *PlayerCacheRepository) GetPlayerById(ctx context.Context, id int) (core.Player, error) {
	res, found := r.players[id]
	if !found {
		return core.Player{}, core.ErrNotFound
	}

	return res, nil
}

func (r *PlayerCacheRepository) GetAllPlayers(ctx context.Context) []core.Player {
	res := make([]core.Player, 0, len(r.players))
	for _, p := range r.players {
		res = append(res, p)
	}

	slices.SortFunc(res, func(a, b core.Player) int {
		return a.Id - b.Id
	})

	return res
}
