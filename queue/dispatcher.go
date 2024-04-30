package queue

import (
	"context"
	"errors"
	"sync"
)

type TaskDispatcher struct {
	Opts     Options
	Queue    chan Task
	Finished bool
}

func (d *TaskDispatcher) Enqueue(task Task) error {
	if d.Finished {
		return errors.New(`queue is closed`)
	}
	d.Queue <- task
	return nil
}

func (d *TaskDispatcher) Start(ctx context.Context) error {
	wg := sync.WaitGroup{}
	errChan := make(chan error, 1)

	for i := 0; i < d.Opts.MaxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					d.Finished = true
					errChan <- ctx.Err()
					return
				case task := <-d.Queue:
					task.Process()
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(errChan)
	}()

	err := <-errChan
	return err
}

func NewTaskDispatcher(opts Options) *TaskDispatcher {
	return &TaskDispatcher{
		Opts:     opts,
		Queue:    make(chan Task, opts.MaxQueueSize),
		Finished: false,
	}
}
