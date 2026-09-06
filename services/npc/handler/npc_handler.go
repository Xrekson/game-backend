package handler

import (
	"context"

	"github.com/xrekson/game-backend/proto/npc"
	"github.com/xrekson/game-backend/services/npc/service"
)

type NPCHandler struct {
	npcpb.UnimplementedNPCServiceServer
	service service.NPCService
}

func NewNPCHandler(service service.NPCService) *NPCHandler {
	return &NPCHandler{service: service}
}

func (h *NPCHandler) GetNPCList(ctx context.Context, req *npcpb.GetNPCListRequest) (*npcpb.GetNPCListResponse, error) {
	npcs, err := h.service.GetNPCsByMapID(ctx, req.MapId)
	if err != nil {
		return nil, err
	}

	var pbNPCs []*npcpb.NPC
	for _, n := range npcs {
		pbNPCs = append(pbNPCs, &npcpb.NPC{
			Id:                   n.ID,
			Name:                 n.Name,
			MapId:                n.MapID,
			Dialogue:             n.Dialogue,
			ShopItemTemplateIds: n.ShopItemTemplateIDs,
		})
	}

	return &npcpb.GetNPCListResponse{Npcs: pbNPCs}, nil
}

func (h *NPCHandler) BuyFromNPC(ctx context.Context, req *npcpb.BuyFromNPCRequest) (*npcpb.BuyFromNPCResponse, error) {
	itemID, pricePaid, newBalance, err := h.service.BuyFromNPC(ctx, req.PlayerId, req.NpcId, req.ItemTemplateId)
	if err != nil {
		return nil, err
	}

	return &npcpb.BuyFromNPCResponse{
		Success:    true,
		ItemId:     itemID,
		PricePaid:  pricePaid,
		NewBalance: newBalance,
	}, nil
}
