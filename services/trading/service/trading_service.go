package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/xrekson/game-backend/pkg/errors"
	"github.com/xrekson/game-backend/pkg/events"
	"github.com/xrekson/game-backend/proto/economy"
	"github.com/xrekson/game-backend/proto/inventory"
	"github.com/xrekson/game-backend/services/trading/model"
	"github.com/xrekson/game-backend/services/trading/repository"
)

type TradingService interface {
	CreateListing(ctx context.Context, sellerID, itemID string, price float64) (*model.TradeListing, error)
	GetListings(ctx context.Context, status string) ([]*model.TradeListing, error)
	BuyListing(ctx context.Context, buyerID, listingID string) (*model.TradeListing, error)
	CancelListing(ctx context.Context, sellerID, listingID string) error
}

type tradingService struct {
	repo            repository.TradingRepository
	economyClient   economypb.EconomyServiceClient
	inventoryClient inventorypb.InventoryServiceClient
}

func NewTradingService(
	repo repository.TradingRepository,
	econClient economypb.EconomyServiceClient,
	invClient inventorypb.InventoryServiceClient,
) TradingService {
	return &tradingService{
		repo:            repo,
		economyClient:   econClient,
		inventoryClient: invClient,
	}
}

func (s *tradingService) CreateListing(ctx context.Context, sellerID, itemID string, price float64) (*model.TradeListing, error) {
	if price <= 0 {
		return nil, errors.BadRequest("price must be greater than 0", nil)
	}

	invRes, err := s.inventoryClient.GetInventory(ctx, &inventorypb.GetInventoryRequest{PlayerId: sellerID})
	if err != nil {
		return nil, errors.Internal("failed fetching inventory from inventory service", err)
	}

	var targetItem *inventorypb.Item
	for _, item := range invRes.Items {
		if item.Id == itemID {
			targetItem = item
			break
		}
	}
	if targetItem == nil {
		return nil, errors.NotFound("seller does not own specified item", nil)
	}

	listing := &model.TradeListing{
		ID:        uuid.New().String(),
		SellerID:  sellerID,
		ItemID:    itemID,
		ItemName:  targetItem.Name,
		Price:     price,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateListing(ctx, listing); err != nil {
		return nil, errors.Internal("failed creating trade listing", err)
	}

	evtPayload, _ := json.Marshal(events.TradeCreatedEvent{
		TradeID:   listing.ID,
		SellerID:  sellerID,
		ItemID:    itemID,
		Price:     price,
		Timestamp: listing.CreatedAt,
	})

	_ = s.repo.SaveOutboxEvent(ctx, &model.OutboxEvent{
		ID:        uuid.New().String(),
		Topic:     events.TopicTradingCreated,
		EventKey:  listing.ID,
		Payload:   string(evtPayload),
		Processed: false,
		CreatedAt: time.Now(),
	})

	return listing, nil
}

func (s *tradingService) GetListings(ctx context.Context, status string) ([]*model.TradeListing, error) {
	listings, err := s.repo.GetListings(ctx, status)
	if err != nil {
		return nil, errors.Internal("failed fetching trade listings", err)
	}
	return listings, nil
}

func (s *tradingService) BuyListing(ctx context.Context, buyerID, listingID string) (*model.TradeListing, error) {
	listing, err := s.repo.GetListingByID(ctx, listingID)
	if err != nil || listing == nil {
		return nil, errors.NotFound("trade listing not found", err)
	}
	if listing.Status != "active" {
		return nil, errors.BadRequest("trade listing is no longer active", nil)
	}
	if listing.SellerID == buyerID {
		return nil, errors.BadRequest("cannot buy your own listing", nil)
	}

	// 1. Deduct currency from buyer
	deductRes, err := s.economyClient.DeductCurrency(ctx, &economypb.DeductCurrencyRequest{
		PlayerId: buyerID,
		Amount:   listing.Price,
	})
	if err != nil || !deductRes.Success {
		return nil, errors.BadRequest(fmt.Sprintf("buyer has insufficient funds (cost: %.2f)", listing.Price), err)
	}

	// 2. Add currency to seller
	_, err = s.economyClient.AddCurrency(ctx, &economypb.AddCurrencyRequest{
		PlayerId: listing.SellerID,
		Amount:   listing.Price,
	})
	if err != nil {
		// Rollback buyer deduction
		_, _ = s.economyClient.AddCurrency(ctx, &economypb.AddCurrencyRequest{PlayerId: buyerID, Amount: listing.Price})
		return nil, errors.Internal("failed transferring funds to seller", err)
	}

	// 3. Transfer item ownership from seller to buyer
	_, err = s.inventoryClient.RemoveItem(ctx, &inventorypb.RemoveItemRequest{PlayerId: listing.SellerID, ItemId: listing.ItemID})
	if err != nil {
		// Rollback payments
		_, _ = s.economyClient.AddCurrency(ctx, &economypb.AddCurrencyRequest{PlayerId: buyerID, Amount: listing.Price})
		_, _ = s.economyClient.DeductCurrency(ctx, &economypb.DeductCurrencyRequest{PlayerId: listing.SellerID, Amount: listing.Price})
		return nil, errors.Internal("failed removing item from seller", err)
	}

	// Add item template to buyer inventory
	_, _ = s.inventoryClient.AddItem(ctx, &inventorypb.AddItemRequest{PlayerId: buyerID, TemplateId: "wep_iron_sword"})

	// 4. Update listing status
	listing.Status = "completed"
	listing.UpdatedAt = time.Now()
	_ = s.repo.UpdateListingStatus(ctx, listing.ID, "completed")

	// 5. Write to transactional outbox
	evtPayload, _ := json.Marshal(events.TradeCompletedEvent{
		TradeID:   listing.ID,
		BuyerID:   buyerID,
		SellerID:  listing.SellerID,
		ItemID:    listing.ItemID,
		Price:     listing.Price,
		Timestamp: time.Now(),
	})

	_ = s.repo.SaveOutboxEvent(ctx, &model.OutboxEvent{
		ID:        uuid.New().String(),
		Topic:     events.TopicTradingCompleted,
		EventKey:  listing.ID,
		Payload:   string(evtPayload),
		Processed: false,
		CreatedAt: time.Now(),
	})

	return listing, nil
}

func (s *tradingService) CancelListing(ctx context.Context, sellerID, listingID string) error {
	listing, err := s.repo.GetListingByID(ctx, listingID)
	if err != nil || listing == nil {
		return errors.NotFound("listing not found", err)
	}
	if listing.SellerID != sellerID {
		return errors.Unauthorized("not owner of listing", nil)
	}
	if listing.Status != "active" {
		return errors.BadRequest("listing is not active", nil)
	}

	return s.repo.UpdateListingStatus(ctx, listingID, "cancelled")
}
