package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/xrekson/game-backend/pkg/kafka"
	"github.com/xrekson/game-backend/pkg/logger"
	"github.com/xrekson/game-backend/services/trading/repository"
)

type OutboxRelay struct {
	repo     repository.TradingRepository
	producer *kafka.Producer
	ticker   *time.Ticker
	stopChan chan struct{}
}

func NewOutboxRelay(repo repository.TradingRepository, producer *kafka.Producer) *OutboxRelay {
	return &OutboxRelay{
		repo:     repo,
		producer: producer,
		stopChan: make(chan struct{}),
	}
}

func (r *OutboxRelay) Start(interval time.Duration) {
	r.ticker = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-r.ticker.C:
				r.processPending()
			case <-r.stopChan:
				r.ticker.Stop()
				return
			}
		}
	}()
	logger.Info("Transactional Outbox Relay started with interval %v", interval)
}

func (r *OutboxRelay) Stop() {
	close(r.stopChan)
}

func (r *OutboxRelay) processPending() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pendingEvents, err := r.repo.GetPendingOutboxEvents(ctx, 50)
	if err != nil || len(pendingEvents) == 0 {
		return
	}

	for _, evt := range pendingEvents {
		var rawPayload any
		if err := json.Unmarshal([]byte(evt.Payload), &rawPayload); err != nil {
			rawPayload = evt.Payload
		}

		err := r.producer.Publish(ctx, evt.Topic, evt.EventKey, rawPayload)
		if err == nil {
			_ = r.repo.MarkOutboxEventProcessed(ctx, evt.ID)
		} else {
			logger.Error("Outbox Relay failed publishing event %s to topic %s: %v", evt.ID, evt.Topic, err)
		}
	}
}
