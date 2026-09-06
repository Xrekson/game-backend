package handler

import (
	"context"

	"github.com/xrekson/game-backend/proto/inventory"
	"github.com/xrekson/game-backend/services/inventory/model"
	"github.com/xrekson/game-backend/services/inventory/service"
)

type InventoryHandler struct {
	inventorypb.UnimplementedInventoryServiceServer
	service service.InventoryService
}

func NewInventoryHandler(service service.InventoryService) *InventoryHandler {
	return &InventoryHandler{service: service}
}

func mapItemToProto(item *model.InventoryItem) *inventorypb.Item {
	if item == nil {
		return nil
	}
	return &inventorypb.Item{
		Id:         item.ID,
		TemplateId: item.TemplateID,
		PlayerId:   item.PlayerID,
		Name:       item.Name,
		Type:       inventorypb.ItemType(item.Type),
		Attack:     item.Attack,
		Defense:    item.Defense,
		Rarity:     item.Rarity,
		IsEquipped: item.IsEquipped,
		Price:      item.Price,
	}
}

func (h *InventoryHandler) GetInventory(ctx context.Context, req *inventorypb.GetInventoryRequest) (*inventorypb.GetInventoryResponse, error) {
	items, err := h.service.GetInventory(ctx, req.PlayerId)
	if err != nil {
		return nil, err
	}

	var pbItems []*inventorypb.Item
	for _, item := range items {
		pbItems = append(pbItems, mapItemToProto(item))
	}

	return &inventorypb.GetInventoryResponse{Items: pbItems}, nil
}

func (h *InventoryHandler) EquipItem(ctx context.Context, req *inventorypb.EquipItemRequest) (*inventorypb.EquipItemResponse, error) {
	item, err := h.service.EquipItem(ctx, req.PlayerId, req.ItemId)
	if err != nil {
		return nil, err
	}

	return &inventorypb.EquipItemResponse{
		Success: true,
		Item:    mapItemToProto(item),
	}, nil
}

func (h *InventoryHandler) UnequipItem(ctx context.Context, req *inventorypb.UnequipItemRequest) (*inventorypb.UnequipItemResponse, error) {
	item, err := h.service.UnequipItem(ctx, req.PlayerId, req.ItemId)
	if err != nil {
		return nil, err
	}

	return &inventorypb.UnequipItemResponse{
		Success: true,
		Item:    mapItemToProto(item),
	}, nil
}

func (h *InventoryHandler) AddItem(ctx context.Context, req *inventorypb.AddItemRequest) (*inventorypb.AddItemResponse, error) {
	item, err := h.service.AddItem(ctx, req.PlayerId, req.TemplateId)
	if err != nil {
		return nil, err
	}

	return &inventorypb.AddItemResponse{
		Success: true,
		Item:    mapItemToProto(item),
	}, nil
}

func (h *InventoryHandler) RemoveItem(ctx context.Context, req *inventorypb.RemoveItemRequest) (*inventorypb.RemoveItemResponse, error) {
	if err := h.service.RemoveItem(ctx, req.PlayerId, req.ItemId); err != nil {
		return nil, err
	}
	return &inventorypb.RemoveItemResponse{Success: true}, nil
}

func (h *InventoryHandler) GetItemTemplates(ctx context.Context, req *inventorypb.GetItemTemplatesRequest) (*inventorypb.GetItemTemplatesResponse, error) {
	tmpls, err := h.service.GetItemTemplates(ctx)
	if err != nil {
		return nil, err
	}

	var pbTmpls []*inventorypb.ItemTemplate
	for _, t := range tmpls {
		pbTmpls = append(pbTmpls, &inventorypb.ItemTemplate{
			TemplateId:  t.TemplateID,
			Name:        t.Name,
			Type:        inventorypb.ItemType(t.Type),
			BaseAttack:  t.BaseAttack,
			BaseDefense: t.BaseDefense,
			Rarity:      t.Rarity,
			BasePrice:   t.BasePrice,
		})
	}

	return &inventorypb.GetItemTemplatesResponse{Templates: pbTmpls}, nil
}
