module github.com/xrekson/game-backend/cmd

go 1.26.6

replace (
	github.com/xrekson/game-backend/gateway => ../gateway
	github.com/xrekson/game-backend/pkg => ../pkg
	github.com/xrekson/game-backend/proto => ../proto
	github.com/xrekson/game-backend/services/auth => ../services/auth
	github.com/xrekson/game-backend/services/economy => ../services/economy
	github.com/xrekson/game-backend/services/inventory => ../services/inventory
	github.com/xrekson/game-backend/services/npc => ../services/npc
	github.com/xrekson/game-backend/services/player => ../services/player
	github.com/xrekson/game-backend/services/trading => ../services/trading
	github.com/xrekson/game-backend/services/world => ../services/world
)

require (
	github.com/gocql/gocql v1.6.0
	github.com/xrekson/game-backend/gateway v0.0.0-00010101000000-000000000000
	github.com/xrekson/game-backend/pkg v0.0.0-00010101000000-000000000000
	github.com/xrekson/game-backend/proto v0.0.0-00010101000000-000000000000
	github.com/xrekson/game-backend/services/auth v0.0.0-00010101000000-000000000000
	github.com/xrekson/game-backend/services/economy v0.0.0-00010101000000-000000000000
	github.com/xrekson/game-backend/services/inventory v0.0.0-00010101000000-000000000000
	github.com/xrekson/game-backend/services/npc v0.0.0-00010101000000-000000000000
	github.com/xrekson/game-backend/services/player v0.0.0-00010101000000-000000000000
	github.com/xrekson/game-backend/services/trading v0.0.0-00010101000000-000000000000
	github.com/xrekson/game-backend/services/world v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.65.0
)

require (
	github.com/bytedance/sonic v1.11.6 // indirect
	github.com/bytedance/sonic/loader v0.1.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gin-gonic/gin v1.10.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.20.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/redis/go-redis/v9 v9.6.1 // indirect
	github.com/segmentio/kafka-go v0.4.47 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	golang.org/x/arch v0.8.0 // indirect
	golang.org/x/crypto v0.26.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.23.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240528184218-531527333157 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
