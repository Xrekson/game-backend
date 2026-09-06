package handler

import (
	"context"

	"github.com/xrekson/game-backend/proto/trading"
	"github.com/xrekson/game-backend/services/trading/service"
)

type TradingHandler struct {
	tradingpb.UnimplementedTradingServiceServer
	service service.TradingService
}

func NewTradingHandler(service service.TradingService) *TradingHandler {
	return &TradingHandler{service: service}
}

func (h *TradingHandler) CreateListing(ctx context.Context, req *tradingpb.CreateListingRequest) (*tradingpb.CreateListingResponse, error) {
	listing, err := h.service.CreateListing(ctx, req.SellerId, req.ItemId, req.Price)
	if err != nil {
		return nil, err
	}

	return &tradingpb.CreateListingResponse{
		Listing: &tradingpb.TradeListing{
			Id:        listing.ID,
			SellerId:  listing.SellerID,
			ItemId:    listing.ItemID,
			ItemName:  listing.ItemName,
			Price:     listing.Price,
			Status:    listing.Status,
			CreatedAt: listing.CreatedAt.Unix(),
		},
	}, nil
}

func (h *TradingHandler) GetListings(ctx context.Context, req *tradingpb.GetListingsRequest) (*tradingpb.GetListingsResponse, error) {
	listings, err := h.service.GetListings(ctx, req.Status)
	if err != nil {
		return nil, err
	}

	var pbListings []*tradingpb.TradeListing
	for _, l := range listings {
		pbListings = append(pbListings, &tradingpb.TradeListing{
			Id:        l.ID,
			SellerId:  l.SellerID,
			ItemId:    l.ItemID,
			ItemName:  l.ItemName,
			Price:     l.Price,
			Status:    l.Status,
			CreatedAt: l.CreatedAt.Unix(),
		})
	}

	return &tradingpb.GetListingsResponse{Listings: pbListings}, nil
}

func (h *TradingHandler) BuyListing(ctx context.Context, req *tradingpb.BuyListingRequest) (*tradingpb.BuyListingResponse, error) {
	listing, err := h.service.BuyListing(ctx, req.BuyerId, req.ListingId)
	if err != nil {
		return nil, err
	}

	return &tradingpb.BuyListingResponse{
		Success: true,
		Listing: &tradingpb.TradeListing{
			Id:        listing.ID,
			SellerId:  listing.SellerID,
			ItemId:    listing.ItemID,
			ItemName:  listing.ItemName,
			Price:     listing.Price,
			Status:    listing.Status,
			CreatedAt: listing.CreatedAt.Unix(),
		},
	}, nil
}

func (h *TradingHandler) CancelListing(ctx context.Context, req *tradingpb.CancelListingRequest) (*tradingpb.CancelListingResponse, error) {
	if err := h.service.CancelListing(ctx, req.SellerId, req.ListingId); err != nil {
		return nil, err
	}
	return &tradingpb.CancelListingResponse{Success: true}, nil
}
