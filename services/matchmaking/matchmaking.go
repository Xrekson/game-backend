package matchmaking

import "context"

type Matchmaker interface {
	FindMatch(ctx context.Context, playerID string) (string, error)
}
