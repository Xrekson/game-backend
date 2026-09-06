module github.com/xrekson/game-backend/gateway

go 1.26.6

replace (
	github.com/xrekson/game-backend/pkg => ../pkg
	github.com/xrekson/game-backend/proto => ../proto
)

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/xrekson/game-backend/pkg v0.0.0-00010101000000-000000000000
	github.com/xrekson/game-backend/proto v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.65.0
)
