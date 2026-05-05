package redis

import (
	"context"

	"github.com/redis/go-redis/v9"

	"github.com/NemCaBong/executify/internal/application/queue"
)

type Producer struct {
	client *redis.Client
}

func NewRedisProducer(client *redis.Client) queue.Producer {
	return &Producer{
		client: client,
	}
}

func (p *Producer) Publish(ctx context.Context, queueName string, message []byte) error {
	return p.client.LPush(ctx, queueName, message).Err()
}

func (p *Producer) Enqueue(ctx context.Context, queueName string, message []byte) error {
	return p.client.RPush(ctx, queueName, message).Err()
}
