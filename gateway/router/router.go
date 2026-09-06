package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/xrekson/game-backend/gateway/middleware"
	"github.com/xrekson/game-backend/pkg/auth"
	"github.com/xrekson/game-backend/pkg/redis"
	"github.com/xrekson/game-backend/proto/auth"
	"github.com/xrekson/game-backend/proto/economy"
	"github.com/xrekson/game-backend/proto/inventory"
	"github.com/xrekson/game-backend/proto/npc"
	"github.com/xrekson/game-backend/proto/player"
	"github.com/xrekson/game-backend/proto/trading"
	"github.com/xrekson/game-backend/proto/world"
)

type Clients struct {
	Auth      authpb.AuthServiceClient
	Player    playerpb.PlayerServiceClient
	Economy   economypb.EconomyServiceClient
	Inventory inventorypb.InventoryServiceClient
	Trading   tradingpb.TradingServiceClient
	NPC       npcpb.NPCServiceClient
	World     worldpb.WorldServiceClient
}

func SetupRouter(jwtMgr *auth.JWTManager, redisClient *redis.Client, clients Clients) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimiterMiddleware(redisClient, 120))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "api-gateway"})
	})

	api := r.Group("/api/v1")

	// Public Auth Endpoints
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", func(c *gin.Context) {
			var req authpb.RegisterRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			res, err := clients.Auth.Register(c.Request.Context(), &req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})

		authGroup.POST("/login", func(c *gin.Context) {
			var req authpb.LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			res, err := clients.Auth.Login(c.Request.Context(), &req)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})
	}

	// Protected Endpoints
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(jwtMgr))

	// Player Profile
	playerGroup := protected.Group("/player")
	{
		playerGroup.GET("/profile", func(c *gin.Context) {
			userID := c.GetString("user_id")
			res, err := clients.Player.GetProfile(c.Request.Context(), &playerpb.GetProfileRequest{PlayerId: userID})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})
	}

	// Economy
	econGroup := protected.Group("/economy")
	{
		econGroup.GET("/balance", func(c *gin.Context) {
			userID := c.GetString("user_id")
			res, err := clients.Economy.GetBalance(c.Request.Context(), &economypb.GetBalanceRequest{PlayerId: userID})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})

		econGroup.POST("/claim", func(c *gin.Context) {
			userID := c.GetString("user_id")
			res, err := clients.Economy.ClaimIncome(c.Request.Context(), &economypb.ClaimIncomeRequest{PlayerId: userID})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})

		econGroup.GET("/upgrades", func(c *gin.Context) {
			userID := c.GetString("user_id")
			res, err := clients.Economy.GetUpgrades(c.Request.Context(), &economypb.GetUpgradesRequest{PlayerId: userID})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})

		econGroup.POST("/upgrade", func(c *gin.Context) {
			userID := c.GetString("user_id")
			var req struct {
				UpgradeID string `json:"upgrade_id"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			res, err := clients.Economy.PurchaseUpgrade(c.Request.Context(), &economypb.PurchaseUpgradeRequest{
				PlayerId:  userID,
				UpgradeId: req.UpgradeID,
			})
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})

		econGroup.POST("/prestige", func(c *gin.Context) {
			userID := c.GetString("user_id")
			res, err := clients.Economy.Prestige(c.Request.Context(), &economypb.PrestigeRequest{PlayerId: userID})
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})
	}

	// Inventory
	invGroup := protected.Group("/inventory")
	{
		invGroup.GET("/items", func(c *gin.Context) {
			userID := c.GetString("user_id")
			res, err := clients.Inventory.GetInventory(c.Request.Context(), &inventorypb.GetInventoryRequest{PlayerId: userID})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})

		invGroup.POST("/equip", func(c *gin.Context) {
			userID := c.GetString("user_id")
			var req struct {
				ItemID string `json:"item_id"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			res, err := clients.Inventory.EquipItem(c.Request.Context(), &inventorypb.EquipItemRequest{PlayerId: userID, ItemId: req.ItemID})
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})

		invGroup.GET("/templates", func(c *gin.Context) {
			res, err := clients.Inventory.GetItemTemplates(c.Request.Context(), &inventorypb.GetItemTemplatesRequest{})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})
	}

	// Trading
	tradingGroup := protected.Group("/trading")
	{
		tradingGroup.GET("/listings", func(c *gin.Context) {
			status := c.Query("status")
			res, err := clients.Trading.GetListings(c.Request.Context(), &tradingpb.GetListingsRequest{Status: status})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})

		tradingGroup.POST("/create", func(c *gin.Context) {
			userID := c.GetString("user_id")
			var req struct {
				ItemID string  `json:"item_id"`
				Price  float64 `json:"price"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			res, err := clients.Trading.CreateListing(c.Request.Context(), &tradingpb.CreateListingRequest{
				SellerId: userID,
				ItemId:   req.ItemID,
				Price:    req.Price,
			})
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})

		tradingGroup.POST("/buy", func(c *gin.Context) {
			userID := c.GetString("user_id")
			var req struct {
				ListingID string `json:"listing_id"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			res, err := clients.Trading.BuyListing(c.Request.Context(), &tradingpb.BuyListingRequest{
				BuyerId:   userID,
				ListingId: req.ListingID,
			})
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})
	}

	// NPC
	npcGroup := protected.Group("/npc")
	{
		npcGroup.GET("/list", func(c *gin.Context) {
			mapID := c.Query("map_id")
			res, err := clients.NPC.GetNPCList(c.Request.Context(), &npcpb.GetNPCListRequest{MapId: mapID})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})

		npcGroup.POST("/buy", func(c *gin.Context) {
			userID := c.GetString("user_id")
			var req struct {
				NPCID          string `json:"npc_id"`
				ItemTemplateID string `json:"item_template_id"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			res, err := clients.NPC.BuyFromNPC(c.Request.Context(), &npcpb.BuyFromNPCRequest{
				PlayerId:       userID,
				NpcId:          req.NPCID,
				ItemTemplateId: req.ItemTemplateID,
			})
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})
	}

	// World
	worldGroup := protected.Group("/world")
	{
		worldGroup.GET("/map", func(c *gin.Context) {
			mapID := c.Query("map_id")
			res, err := clients.World.GetMapLayout(c.Request.Context(), &worldpb.GetMapLayoutRequest{MapId: mapID})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})

		worldGroup.POST("/move", func(c *gin.Context) {
			userID := c.GetString("user_id")
			var req struct {
				MapID string  `json:"map_id"`
				X     float64 `json:"x"`
				Y     float64 `json:"y"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			res, err := clients.World.ValidateMovement(c.Request.Context(), &worldpb.ValidateMovementRequest{
				PlayerId:    userID,
				MapId:       req.MapID,
				NewPosition: &worldpb.Position{X: req.X, Y: req.Y},
			})
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})

		worldGroup.POST("/combat", func(c *gin.Context) {
			userID := c.GetString("user_id")
			var req struct {
				EnemyID string  `json:"enemy_id"`
				Damage  float64 `json:"damage"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			res, err := clients.World.ResolveCombat(c.Request.Context(), &worldpb.ResolveCombatRequest{
				PlayerId: userID,
				EnemyId:  req.EnemyID,
				Damage:   req.Damage,
			})
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, res)
		})
	}

	return r
}
