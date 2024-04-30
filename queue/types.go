package queue

import "context"

type Options struct {
	MaxWorkers   int
	MaxQueueSize int
}

type Task interface {
	Process() error
}

type Dispatcher interface {
	Enqueue(task Task) error
	Start(ctx context.Context) error
}
