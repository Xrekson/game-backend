package templaterepo

import (
	"context"
	"encoding/json"

	"github.com/xrekson/game-backend/pkg/cassandra"
	"github.com/xrekson/game-backend/services/inventory/model"
)

type TemplateRepository interface {
	GetTemplates(ctx context.Context) ([]*model.ItemTemplate, error)
	GetTemplateByID(ctx context.Context, templateID string) (*model.ItemTemplate, error)
	SaveTemplate(ctx context.Context, template *model.ItemTemplate) error
	UpdateTemplate(ctx context.Context, template *model.ItemTemplate) error
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
			base_price double,
			has_attributes boolean,
			attributes_json text
		);`
	if err := r.session.Query(createTableCQL).Exec(); err != nil {
		return err
	}

	for _, tmpl := range model.DefaultItemTemplates {
		_ = r.SaveTemplate(ctx, &tmpl)
	}

	return nil
}

func (r *cassandraTemplateRepository) SaveTemplate(ctx context.Context, template *model.ItemTemplate) error {
	if r.session == nil {
		return nil
	}

	attrsJSON, _ := json.Marshal(template.Attributes)
	insertCQL := `
		INSERT INTO item_templates (template_id, name, type, base_attack, base_defense, rarity, base_price, has_attributes, attributes_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);`

	return r.session.Query(insertCQL,
		template.TemplateID, template.Name, int(template.Type), template.BaseAttack, template.BaseDefense,
		template.Rarity, template.BasePrice, template.HasAttributes, string(attrsJSON),
	).Exec()
}

func (r *cassandraTemplateRepository) UpdateTemplate(ctx context.Context, template *model.ItemTemplate) error {
	return r.SaveTemplate(ctx, template)
}

func (r *cassandraTemplateRepository) GetTemplates(ctx context.Context) ([]*model.ItemTemplate, error) {
	if r.session == nil {
		var templates []*model.ItemTemplate
		for i := range model.DefaultItemTemplates {
			templates = append(templates, &model.DefaultItemTemplates[i])
		}
		return templates, nil
	}

	queryCQL := `SELECT template_id, name, type, base_attack, base_defense, rarity, base_price, has_attributes, attributes_json FROM item_templates;`
	iter := r.session.Query(queryCQL).Iter()

	var templates []*model.ItemTemplate
	var templateID, name, rarity, attrsJSON string
	var itemType int
	var attack, defense, price float64
	var hasAttributes bool

	for iter.Scan(&templateID, &name, &itemType, &attack, &defense, &rarity, &price, &hasAttributes, &attrsJSON) {
		var attrs []model.ItemAttribute
		if attrsJSON != "" {
			_ = json.Unmarshal([]byte(attrsJSON), &attrs)
		}

		templates = append(templates, &model.ItemTemplate{
			TemplateID:    templateID,
			Name:          name,
			Type:          model.ItemType(itemType),
			BaseAttack:    attack,
			BaseDefense:   defense,
			Rarity:        rarity,
			BasePrice:     price,
			HasAttributes: hasAttributes,
			Attributes:    attrs,
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
		queryCQL := `SELECT template_id, name, type, base_attack, base_defense, rarity, base_price, has_attributes, attributes_json FROM item_templates WHERE template_id = ? LIMIT 1;`
		tmpl := &model.ItemTemplate{}
		var itemType int
		var attrsJSON string
		err := r.session.Query(queryCQL, templateID).Scan(
			&tmpl.TemplateID, &tmpl.Name, &itemType, &tmpl.BaseAttack, &tmpl.BaseDefense, &tmpl.Rarity, &tmpl.BasePrice, &tmpl.HasAttributes, &attrsJSON,
		)
		if err == nil {
			tmpl.Type = model.ItemType(itemType)
			if attrsJSON != "" {
				_ = json.Unmarshal([]byte(attrsJSON), &tmpl.Attributes)
			}
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
