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

func mapAttributesToProto(attrs []model.ItemAttribute) []*inventorypb.ItemAttribute {
	var pbAttrs []*inventorypb.ItemAttribute
	for _, a := range attrs {
		pbAttrs = append(pbAttrs, &inventorypb.ItemAttribute{
			Key:             a.Key,
			Type:            a.Type,
			Value:           a.Value,
			CalculationExpr: a.CalculationExpr,
		})
	}
	return pbAttrs
}

func mapProtoToAttributes(pbAttrs []*inventorypb.ItemAttribute) []model.ItemAttribute {
	var attrs []model.ItemAttribute
	for _, a := range pbAttrs {
		if a == nil {
			continue
		}
		attrs = append(attrs, model.ItemAttribute{
			Key:             a.Key,
			Type:            a.Type,
			Value:           a.Value,
			CalculationExpr: a.CalculationExpr,
		})
	}
	return attrs
}

func mapItemToProto(item *model.InventoryItem) *inventorypb.Item {
	if item == nil {
		return nil
	}
	return &inventorypb.Item{
		Id:            item.ID,
		TemplateId:    item.TemplateID,
		PlayerId:      item.PlayerID,
		Name:          item.Name,
		Type:          inventorypb.ItemType(item.Type),
		Attack:        item.Attack,
		Defense:       item.Defense,
		Rarity:        item.Rarity,
		IsEquipped:    item.IsEquipped,
		Price:         item.Price,
		HasAttributes: item.HasAttributes,
		Attributes:    mapAttributesToProto(item.Attributes),
	}
}

func mapTemplateToProto(t *model.ItemTemplate) *inventorypb.ItemTemplate {
	if t == nil {
		return nil
	}
	return &inventorypb.ItemTemplate{
		TemplateId:    t.TemplateID,
		Name:          t.Name,
		Type:          inventorypb.ItemType(t.Type),
		BaseAttack:    t.BaseAttack,
		BaseDefense:   t.BaseDefense,
		Rarity:        t.Rarity,
		BasePrice:     t.BasePrice,
		HasAttributes: t.HasAttributes,
		Attributes:    mapAttributesToProto(t.Attributes),
	}
}

func mapProtoToTemplate(t *inventorypb.ItemTemplate) *model.ItemTemplate {
	if t == nil {
		return nil
	}
	return &model.ItemTemplate{
		TemplateID:    t.TemplateId,
		Name:          t.Name,
		Type:          model.ItemType(t.Type),
		BaseAttack:    t.BaseAttack,
		BaseDefense:   t.BaseDefense,
		Rarity:        t.Rarity,
		BasePrice:     t.BasePrice,
		HasAttributes: t.HasAttributes,
		Attributes:    mapProtoToAttributes(t.Attributes),
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
		pbTmpls = append(pbTmpls, mapTemplateToProto(t))
	}

	return &inventorypb.GetItemTemplatesResponse{Templates: pbTmpls}, nil
}

func (h *InventoryHandler) CreateItemTemplate(ctx context.Context, req *inventorypb.CreateItemTemplateRequest) (*inventorypb.CreateItemTemplateResponse, error) {
	modelTmpl := mapProtoToTemplate(req.Template)
	created, err := h.service.CreateItemTemplate(ctx, modelTmpl)
	if err != nil {
		return nil, err
	}

	return &inventorypb.CreateItemTemplateResponse{
		Success:  true,
		Template: mapTemplateToProto(created),
	}, nil
}

func (h *InventoryHandler) UpdateItemTemplate(ctx context.Context, req *inventorypb.UpdateItemTemplateRequest) (*inventorypb.UpdateItemTemplateResponse, error) {
	modelTmpl := mapProtoToTemplate(req.Template)
	updated, err := h.service.UpdateItemTemplate(ctx, modelTmpl)
	if err != nil {
		return nil, err
	}

	return &inventorypb.UpdateItemTemplateResponse{
		Success:  true,
		Template: mapTemplateToProto(updated),
	}, nil
}
