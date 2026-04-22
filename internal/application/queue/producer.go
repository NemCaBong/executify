package queue

import "context"

type Producer interface {
	Publish(ctx context.Context, queueName string, message []byte) error
	Enqueue(ctx context.Context, queueName string, message []byte) error
}
