module github.com/xrekson/game-backend/services/trading

go 1.26.6

replace (
	github.com/xrekson/game-backend/pkg => ../../pkg
	github.com/xrekson/game-backend/proto => ../../proto
)

require (
	github.com/google/uuid v1.6.0
	github.com/xrekson/game-backend/pkg v0.0.0-00010101000000-000000000000
	github.com/xrekson/game-backend/proto v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.65.0
)
