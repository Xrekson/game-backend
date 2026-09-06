package repository

import (
	"context"

	"github.com/xrekson/game-backend/pkg/cassandra"
	"github.com/xrekson/game-backend/services/world/model"
)

type MapRepository interface {
	GetMapLayout(ctx context.Context, mapID string) (*model.MapLayout, error)
	SeedMaps(ctx context.Context) error
}

type cassandraMapRepository struct {
	session *cassandra.Session
}

func NewMapRepository(session *cassandra.Session) MapRepository {
	repo := &cassandraMapRepository{session: session}
	if session != nil {
		_ = repo.SeedMaps(context.Background())
	}
	return repo
}

func (r *cassandraMapRepository) SeedMaps(ctx context.Context) error {
	if r.session == nil {
		return nil
	}

	createTableCQL := `
		CREATE TABLE IF NOT EXISTS dungeon_layouts (
			map_id text PRIMARY KEY,
			name text,
			width double,
			height double,
			is_safe_zone boolean,
			spawn_x double,
			spawn_y double
		);`
	if err := r.session.Query(createTableCQL).Exec(); err != nil {
		return err
	}

	insertCQL := `
		INSERT INTO dungeon_layouts (map_id, name, width, height, is_safe_zone, spawn_x, spawn_y)
		VALUES (?, ?, ?, ?, ?, ?, ?);`

	for _, m := range model.DefaultMaps {
		_ = r.session.Query(insertCQL, m.MapID, m.Name, m.Width, m.Height, m.IsSafeZone, m.SpawnPoint.X, m.SpawnPoint.Y).Exec()
	}

	return nil
}

func (r *cassandraMapRepository) GetMapLayout(ctx context.Context, mapID string) (*model.MapLayout, error) {
	if r.session != nil {
		queryCQL := `SELECT map_id, name, width, height, is_safe_zone, spawn_x, spawn_y FROM dungeon_layouts WHERE map_id = ? LIMIT 1;`
		m := &model.MapLayout{}
		var spawnX, spawnY float64
		err := r.session.Query(queryCQL, mapID).Scan(
			&m.MapID, &m.Name, &m.Width, &m.Height, &m.IsSafeZone, &spawnX, &spawnY,
		)
		if err == nil {
			m.SpawnPoint = model.Position{X: spawnX, Y: spawnY}
			return m, nil
		}
	}

	for _, def := range model.DefaultMaps {
		if def.MapID == mapID {
			return &def, nil
		}
	}
	return nil, nil
}
