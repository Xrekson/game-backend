package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/xrekson/game-backend/pkg/errors"
	"github.com/xrekson/game-backend/services/inventory/model"
	"github.com/xrekson/game-backend/services/inventory/repository"
	"github.com/xrekson/game-backend/services/inventory/templaterepo"
)

type InventoryService interface {
	GetInventory(ctx context.Context, playerID string) ([]*model.InventoryItem, error)
	EquipItem(ctx context.Context, playerID, itemID string) (*model.InventoryItem, error)
	UnequipItem(ctx context.Context, playerID, itemID string) (*model.InventoryItem, error)
	AddItem(ctx context.Context, playerID, templateID string) (*model.InventoryItem, error)
	RemoveItem(ctx context.Context, playerID, itemID string) error
	GetItemTemplates(ctx context.Context) ([]*model.ItemTemplate, error)
	CreateItemTemplate(ctx context.Context, tmpl *model.ItemTemplate) (*model.ItemTemplate, error)
	UpdateItemTemplate(ctx context.Context, tmpl *model.ItemTemplate) (*model.ItemTemplate, error)
}

type inventoryService struct {
	repo         repository.InventoryRepository
	templateRepo templaterepo.TemplateRepository
}

func NewInventoryService(repo repository.InventoryRepository, templateRepo templaterepo.TemplateRepository) InventoryService {
	return &inventoryService{
		repo:         repo,
		templateRepo: templateRepo,
	}
}

func (s *inventoryService) GetInventory(ctx context.Context, playerID string) ([]*model.InventoryItem, error) {
	items, err := s.repo.GetInventory(ctx, playerID)
	if err != nil {
		return nil, errors.Internal("failed to fetch inventory", err)
	}
	return items, nil
}

func (s *inventoryService) EquipItem(ctx context.Context, playerID, itemID string) (*model.InventoryItem, error) {
	item, err := s.repo.GetItemByID(ctx, itemID)
	if err != nil || item == nil {
		return nil, errors.NotFound("item not found", err)
	}
	if item.PlayerID != playerID {
		return nil, errors.Unauthorized("not owner of item", nil)
	}

	// Unequip currently equipped item of the same slot type
	if err := s.repo.UnequipAllOfType(ctx, playerID, item.Type); err != nil {
		return nil, errors.Internal("failed to unequip existing item slot", err)
	}

	if err := s.repo.SetEquipped(ctx, itemID, true); err != nil {
		return nil, errors.Internal("failed to equip item", err)
	}

	item.IsEquipped = true
	return item, nil
}

func (s *inventoryService) UnequipItem(ctx context.Context, playerID, itemID string) (*model.InventoryItem, error) {
	item, err := s.repo.GetItemByID(ctx, itemID)
	if err != nil || item == nil {
		return nil, errors.NotFound("item not found", err)
	}
	if item.PlayerID != playerID {
		return nil, errors.Unauthorized("not owner of item", nil)
	}

	if err := s.repo.SetEquipped(ctx, itemID, false); err != nil {
		return nil, errors.Internal("failed to unequip item", err)
	}

	item.IsEquipped = false
	return item, nil
}

func (s *inventoryService) AddItem(ctx context.Context, playerID, templateID string) (*model.InventoryItem, error) {
	tmpl, err := s.templateRepo.GetTemplateByID(ctx, templateID)
	if err != nil || tmpl == nil {
		return nil, errors.NotFound("item template not found", err)
	}

	item := &model.InventoryItem{
		ID:            uuid.New().String(),
		PlayerID:      playerID,
		TemplateID:    tmpl.TemplateID,
		Name:          tmpl.Name,
		Type:          tmpl.Type,
		Attack:        tmpl.BaseAttack,
		Defense:       tmpl.BaseDefense,
		Rarity:        tmpl.Rarity,
		IsEquipped:    false,
		Price:         tmpl.BasePrice,
		HasAttributes: tmpl.HasAttributes,
		Attributes:    tmpl.Attributes,
		CreatedAt:     time.Now(),
	}

	if err := s.repo.AddItem(ctx, item); err != nil {
		return nil, errors.Internal("failed to add item to inventory", err)
	}

	return item, nil
}

func (s *inventoryService) RemoveItem(ctx context.Context, playerID, itemID string) error {
	item, err := s.repo.GetItemByID(ctx, itemID)
	if err != nil || item == nil {
		return errors.NotFound("item not found", err)
	}
	if item.PlayerID != playerID {
		return errors.Unauthorized("not owner of item", nil)
	}

	if err := s.repo.RemoveItem(ctx, itemID); err != nil {
		return errors.Internal("failed to remove item", err)
	}
	return nil
}

func (s *inventoryService) GetItemTemplates(ctx context.Context) ([]*model.ItemTemplate, error) {
	tmpls, err := s.templateRepo.GetTemplates(ctx)
	if err != nil {
		return nil, errors.Internal("failed to fetch templates", err)
	}
	return tmpls, nil
}

func (s *inventoryService) CreateItemTemplate(ctx context.Context, tmpl *model.ItemTemplate) (*model.ItemTemplate, error) {
	if tmpl.TemplateID == "" {
		tmpl.TemplateID = "tmpl_" + uuid.New().String()[:8]
	}
	if err := s.templateRepo.SaveTemplate(ctx, tmpl); err != nil {
		return nil, errors.Internal("failed to save item template", err)
	}
	return tmpl, nil
}

func (s *inventoryService) UpdateItemTemplate(ctx context.Context, tmpl *model.ItemTemplate) (*model.ItemTemplate, error) {
	if tmpl.TemplateID == "" {
		return nil, errors.BadRequest("template_id is required", nil)
	}
	if err := s.templateRepo.UpdateTemplate(ctx, tmpl); err != nil {
		return nil, errors.Internal("failed to update item template", err)
	}
	return tmpl, nil
}
