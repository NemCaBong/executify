package redis

import (
	"context"

	"github.com/NemCaBong/executify/internal/application/queue"
	"github.com/redis/go-redis/v9"
)

type RedisProducer struct {
	client *redis.Client
}

func NewRedisProducer(client *redis.Client) queue.Producer {
	return &RedisProducer{
		client: client,
	}
}

func (p *RedisProducer) Publish(ctx context.Context, queueName string, message []byte) error {
	return p.client.LPush(ctx, queueName, message).Err()
}

func (p *RedisProducer) Enqueue(ctx context.Context, queueName string, message []byte) error {
	return p.client.RPush(ctx, queueName, message).Err()
}
