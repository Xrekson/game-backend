package service

import (
	"context"
	"fmt"

	"github.com/xrekson/game-backend/pkg/errors"
	"github.com/xrekson/game-backend/proto/economy"
	"github.com/xrekson/game-backend/proto/inventory"
	"github.com/xrekson/game-backend/services/npc/model"
	"github.com/xrekson/game-backend/services/npc/repository"
)

type NPCService interface {
	GetNPCsByMapID(ctx context.Context, mapID string) ([]*model.NPC, error)
	BuyFromNPC(ctx context.Context, playerID, npcID, itemTemplateID string) (string, float64, float64, error)
}

type npcService struct {
	repo          repository.NPCRepository
	economyClient economypb.EconomyServiceClient
	inventoryClient inventorypb.InventoryServiceClient
}

func NewNPCService(
	repo repository.NPCRepository,
	econClient economypb.EconomyServiceClient,
	invClient inventorypb.InventoryServiceClient,
) NPCService {
	return &npcService{
		repo:            repo,
		economyClient:   econClient,
		inventoryClient: invClient,
	}
}

func (s *npcService) GetNPCsByMapID(ctx context.Context, mapID string) ([]*model.NPC, error) {
	npcs, err := s.repo.GetNPCsByMapID(ctx, mapID)
	if err != nil {
		return nil, errors.Internal("failed to fetch NPCs", err)
	}
	return npcs, nil
}

func (s *npcService) BuyFromNPC(ctx context.Context, playerID, npcID, itemTemplateID string) (string, float64, float64, error) {
	npc, err := s.repo.GetNPCByID(ctx, npcID)
	if err != nil || npc == nil {
		return "", 0, 0, errors.NotFound("NPC not found", err)
	}

	hasItem := false
	for _, id := range npc.ShopItemTemplateIDs {
		if id == itemTemplateID {
			hasItem = true
			break
		}
	}
	if !hasItem {
		return "", 0, 0, errors.BadRequest("NPC does not sell this item", nil)
	}

	// Fetch template info from inventory service
	tmplsRes, err := s.inventoryClient.GetItemTemplates(ctx, &inventorypb.GetItemTemplatesRequest{})
	if err != nil {
		return "", 0, 0, errors.Internal("failed fetching item templates from inventory service", err)
	}

	var price float64
	found := false
	for _, t := range tmplsRes.Templates {
		if t.TemplateId == itemTemplateID {
			price = t.BasePrice
			found = true
			break
		}
	}
	if !found {
		return "", 0, 0, errors.NotFound("item template not found", nil)
	}

	// Deduct funds from economy service
	deductRes, err := s.economyClient.DeductCurrency(ctx, &economypb.DeductCurrencyRequest{
		PlayerId: playerID,
		Amount:   price,
	})
	if err != nil || !deductRes.Success {
		return "", 0, 0, errors.BadRequest(fmt.Sprintf("failed deducting funds or insufficient gold (cost: %.2f)", price), err)
	}

	// Add item to player inventory
	addRes, err := s.inventoryClient.AddItem(ctx, &inventorypb.AddItemRequest{
		PlayerId:   playerID,
		TemplateId: itemTemplateID,
	})
	if err != nil || !addRes.Success {
		// Rollback deducted currency if inventory addition fails
		_, _ = s.economyClient.AddCurrency(ctx, &economypb.AddCurrencyRequest{
			PlayerId: playerID,
			Amount:   price,
		})
		return "", 0, 0, errors.Internal("failed adding item to inventory", err)
	}

	return addRes.Item.Id, price, deductRes.NewBalance, nil
}
