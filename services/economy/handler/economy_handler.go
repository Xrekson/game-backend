package handler

import (
	"context"

	"github.com/xrekson/game-backend/proto/economy"
	"github.com/xrekson/game-backend/services/economy/service"
)

type EconomyHandler struct {
	economypb.UnimplementedEconomyServiceServer
	service service.EconomyService
}

func NewEconomyHandler(service service.EconomyService) *EconomyHandler {
	return &EconomyHandler{service: service}
}

func (h *EconomyHandler) GetBalance(ctx context.Context, req *economypb.GetBalanceRequest) (*economypb.GetBalanceResponse, error) {
	b, incomePerSec, err := h.service.GetBalance(ctx, req.PlayerId)
	if err != nil {
		return nil, err
	}

	return &economypb.GetBalanceResponse{
		PlayerId:           b.PlayerID,
		Balance:            b.Balance,
		IncomePerSec:       incomePerSec,
		PrestigeLevel:      b.PrestigeLevel,
		PrestigeMultiplier: b.PrestigeMultiplier,
		LastClaimTime:      b.LastClaimTime.Unix(),
	}, nil
}

func (h *EconomyHandler) ClaimIncome(ctx context.Context, req *economypb.ClaimIncomeRequest) (*economypb.ClaimIncomeResponse, error) {
	b, claimed, err := h.service.ClaimIncome(ctx, req.PlayerId)
	if err != nil {
		return nil, err
	}

	return &economypb.ClaimIncomeResponse{
		PlayerId:      b.PlayerID,
		ClaimedAmount: claimed,
		NewBalance:    b.Balance,
	}, nil
}

func (h *EconomyHandler) GetUpgrades(ctx context.Context, req *economypb.GetUpgradesRequest) (*economypb.GetUpgradesResponse, error) {
	configs, levelMap, err := h.service.GetUpgrades(ctx, req.PlayerId)
	if err != nil {
		return nil, err
	}

	var upgrades []*economypb.UpgradeItem
	for _, cfg := range configs {
		upgrades = append(upgrades, &economypb.UpgradeItem{
			Id:                cfg.ID,
			Name:              cfg.Name,
			Level:             levelMap[cfg.ID],
			BaseCost:          cfg.BaseCost,
			CurrentCost:       cfg.BaseCost,
			IncomeBonusPerSec: cfg.IncomeBonusPerSec,
			Description:       cfg.Description,
		})
	}

	return &economypb.GetUpgradesResponse{Upgrades: upgrades}, nil
}

func (h *EconomyHandler) PurchaseUpgrade(ctx context.Context, req *economypb.PurchaseUpgradeRequest) (*economypb.PurchaseUpgradeResponse, error) {
	b, newIncome, u, err := h.service.PurchaseUpgrade(ctx, req.PlayerId, req.UpgradeId)
	if err != nil {
		return nil, err
	}

	return &economypb.PurchaseUpgradeResponse{
		Success:         true,
		NewBalance:      b.Balance,
		NewIncomePerSec: newIncome,
		Upgrade: &economypb.UpgradeItem{
			Id:    u.UpgradeID,
			Level: u.Level,
		},
	}, nil
}

func (h *EconomyHandler) Prestige(ctx context.Context, req *economypb.PrestigeRequest) (*economypb.PrestigeResponse, error) {
	b, incomePerSec, err := h.service.Prestige(ctx, req.PlayerId)
	if err != nil {
		return nil, err
	}

	return &economypb.PrestigeResponse{
		Success:               true,
		NewPrestigeLevel:      b.PrestigeLevel,
		NewPrestigeMultiplier: b.PrestigeMultiplier,
		NewBalance:            b.Balance,
		NewIncomePerSec:       incomePerSec,
	}, nil
}

func (h *EconomyHandler) DeductCurrency(ctx context.Context, req *economypb.DeductCurrencyRequest) (*economypb.DeductCurrencyResponse, error) {
	newBal, err := h.service.DeductCurrency(ctx, req.PlayerId, req.Amount)
	if err != nil {
		return nil, err
	}

	return &economypb.DeductCurrencyResponse{
		Success:    true,
		NewBalance: newBal,
	}, nil
}

func (h *EconomyHandler) AddCurrency(ctx context.Context, req *economypb.AddCurrencyRequest) (*economypb.AddCurrencyResponse, error) {
	newBal, err := h.service.AddCurrency(ctx, req.PlayerId, req.Amount)
	if err != nil {
		return nil, err
	}

	return &economypb.AddCurrencyResponse{
		Success:    true,
		NewBalance: newBal,
	}, nil
}
