package repository

import (
	"context"

	"github.com/xrekson/game-backend/pkg/cassandra"
	"github.com/xrekson/game-backend/services/npc/model"
)

type NPCRepository interface {
	GetNPCsByMapID(ctx context.Context, mapID string) ([]*model.NPC, error)
	GetNPCByID(ctx context.Context, npcID string) (*model.NPC, error)
	SeedNPCs(ctx context.Context) error
}

type cassandraNPCRepository struct {
	session *cassandra.Session
}

func NewNPCRepository(session *cassandra.Session) NPCRepository {
	repo := &cassandraNPCRepository{session: session}
	if session != nil {
		_ = repo.SeedNPCs(context.Background())
	}
	return repo
}

func (r *cassandraNPCRepository) SeedNPCs(ctx context.Context) error {
	if r.session == nil {
		return nil
	}

	createTableCQL := `
		CREATE TABLE IF NOT EXISTS npcs (
			id text PRIMARY KEY,
			name text,
			map_id text,
			dialogue text,
			shop_item_template_ids list<text>
		);`
	if err := r.session.Query(createTableCQL).Exec(); err != nil {
		return err
	}

	createIndexCQL := `CREATE INDEX IF NOT EXISTS idx_npcs_map_id ON npcs(map_id);`
	_ = r.session.Query(createIndexCQL).Exec()

	insertCQL := `
		INSERT INTO npcs (id, name, map_id, dialogue, shop_item_template_ids)
		VALUES (?, ?, ?, ?, ?);`

	for _, n := range model.DefaultNPCs {
		_ = r.session.Query(insertCQL, n.ID, n.Name, n.MapID, n.Dialogue, n.ShopItemTemplateIDs).Exec()
	}

	return nil
}

func (r *cassandraNPCRepository) GetNPCsByMapID(ctx context.Context, mapID string) ([]*model.NPC, error) {
	if r.session != nil {
		var queryCQL string
		var args []any

		if mapID != "" {
			queryCQL = `SELECT id, name, map_id, dialogue, shop_item_template_ids FROM npcs WHERE map_id = ?;`
			args = append(args, mapID)
		} else {
			queryCQL = `SELECT id, name, map_id, dialogue, shop_item_template_ids FROM npcs;`
		}

		iter := r.session.Query(queryCQL, args...).Iter()
		var npcs []*model.NPC
		var id, name, mID, dialogue string
		var templates []string

		for iter.Scan(&id, &name, &mID, &dialogue, &templates) {
			npcs = append(npcs, &model.NPC{
				ID:                   id,
				Name:                 name,
				MapID:                mID,
				Dialogue:             dialogue,
				ShopItemTemplateIDs: templates,
			})
		}

		if err := iter.Close(); err == nil && len(npcs) > 0 {
			return npcs, nil
		}
	}

	var fallback []*model.NPC
	for i := range model.DefaultNPCs {
		if model.DefaultNPCs[i].MapID == mapID || mapID == "" {
			fallback = append(fallback, &model.DefaultNPCs[i])
		}
	}
	return fallback, nil
}

func (r *cassandraNPCRepository) GetNPCByID(ctx context.Context, npcID string) (*model.NPC, error) {
	if r.session != nil {
		queryCQL := `SELECT id, name, map_id, dialogue, shop_item_template_ids FROM npcs WHERE id = ? LIMIT 1;`
		n := &model.NPC{}
		var templates []string
		err := r.session.Query(queryCQL, npcID).Scan(&n.ID, &n.Name, &n.MapID, &n.Dialogue, &templates)
		if err == nil {
			n.ShopItemTemplateIDs = templates
			return n, nil
		}
	}

	for _, npc := range model.DefaultNPCs {
		if npc.ID == npcID {
			return &npc, nil
		}
	}
	return nil, nil
}
