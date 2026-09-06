package templaterepo

import (
	"context"

	"github.com/xrekson/game-backend/pkg/cassandra"
	"github.com/xrekson/game-backend/services/inventory/model"
)

type TemplateRepository interface {
	GetTemplates(ctx context.Context) ([]*model.ItemTemplate, error)
	GetTemplateByID(ctx context.Context, templateID string) (*model.ItemTemplate, error)
	SeedTemplates(ctx context.Context) error
}

type cassandraTemplateRepository struct {
	session *cassandra.Session
}

func NewTemplateRepository(session *cassandra.Session) TemplateRepository {
	repo := &cassandraTemplateRepository{session: session}
	if session != nil {
		_ = repo.SeedTemplates(context.Background())
	}
	return repo
}

func (r *cassandraTemplateRepository) SeedTemplates(ctx context.Context) error {
	if r.session == nil {
		return nil
	}

	createTableCQL := `
		CREATE TABLE IF NOT EXISTS item_templates (
			template_id text PRIMARY KEY,
			name text,
			type int,
			base_attack double,
			base_defense double,
			rarity text,
			base_price double
		);`
	if err := r.session.Query(createTableCQL).Exec(); err != nil {
		return err
	}

	insertCQL := `
		INSERT INTO item_templates (template_id, name, type, base_attack, base_defense, rarity, base_price)
		VALUES (?, ?, ?, ?, ?, ?, ?);`

	for _, tmpl := range model.DefaultItemTemplates {
		_ = r.session.Query(insertCQL,
			tmpl.TemplateID, tmpl.Name, int(tmpl.Type), tmpl.BaseAttack, tmpl.BaseDefense, tmpl.Rarity, tmpl.BasePrice,
		).Exec()
	}

	return nil
}

func (r *cassandraTemplateRepository) GetTemplates(ctx context.Context) ([]*model.ItemTemplate, error) {
	if r.session == nil {
		var templates []*model.ItemTemplate
		for i := range model.DefaultItemTemplates {
			templates = append(templates, &model.DefaultItemTemplates[i])
		}
		return templates, nil
	}

	queryCQL := `SELECT template_id, name, type, base_attack, base_defense, rarity, base_price FROM item_templates;`
	iter := r.session.Query(queryCQL).Iter()

	var templates []*model.ItemTemplate
	var templateID, name, rarity string
	var itemType int
	var attack, defense, price float64

	for iter.Scan(&templateID, &name, &itemType, &attack, &defense, &rarity, &price) {
		templates = append(templates, &model.ItemTemplate{
			TemplateID:  templateID,
			Name:        name,
			Type:        model.ItemType(itemType),
			BaseAttack:  attack,
			BaseDefense: defense,
			Rarity:      rarity,
			BasePrice:   price,
		})
	}

	if err := iter.Close(); err != nil || len(templates) == 0 {
		var defaultList []*model.ItemTemplate
		for i := range model.DefaultItemTemplates {
			defaultList = append(defaultList, &model.DefaultItemTemplates[i])
		}
		return defaultList, nil
	}

	return templates, nil
}

func (r *cassandraTemplateRepository) GetTemplateByID(ctx context.Context, templateID string) (*model.ItemTemplate, error) {
	if r.session != nil {
		queryCQL := `SELECT template_id, name, type, base_attack, base_defense, rarity, base_price FROM item_templates WHERE template_id = ? LIMIT 1;`
		tmpl := &model.ItemTemplate{}
		var itemType int
		err := r.session.Query(queryCQL, templateID).Scan(
			&tmpl.TemplateID, &tmpl.Name, &itemType, &tmpl.BaseAttack, &tmpl.BaseDefense, &tmpl.Rarity, &tmpl.BasePrice,
		)
		if err == nil {
			tmpl.Type = model.ItemType(itemType)
			return tmpl, nil
		}
	}

	for _, t := range model.DefaultItemTemplates {
		if t.TemplateID == templateID {
			return &t, nil
		}
	}
	return nil, nil
}
