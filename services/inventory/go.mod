module github.com/xrekson/game-backend/services/inventory

go 1.26.6

replace (
	github.com/xrekson/game-backend/pkg => ../../pkg
	github.com/xrekson/game-backend/proto => ../../proto
)

require (
	github.com/google/uuid v1.6.0
	github.com/xrekson/game-backend/pkg v0.0.0-00010101000000-000000000000
	github.com/xrekson/game-backend/proto v0.0.0-00010101000000-000000000000
)

require (
	github.com/gocql/gocql v1.6.0 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	golang.org/x/crypto v0.26.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240528184218-531527333157 // indirect
	google.golang.org/grpc v1.65.0 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
