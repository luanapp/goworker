package goworker

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestRunTaskBeforeStarting(t *testing.T) {
	workerPool := NewWorkerPool(
		WithContext(context.Background()),
		WithIdleTimeout(3*time.Second),
		WithPoolSize(2),
	)

	workerPool.AddJobs(
		func(ctx context.Context) error {
			fmt.Println("job 1 running")
			time.Sleep(200 * time.Millisecond)
			return nil
		},
		func(ctx context.Context) error {
			fmt.Println("job 2 running")
			time.Sleep(2 * time.Second)
			return errors.New("this is error")
		},
		func(ctx context.Context) error {
			fmt.Println("job 3 running")
			time.Sleep(1 * time.Second)
			return nil
		},
	)

}

func TestMultipleStartCalls(t *testing.T) {
	workerPool := NewWorkerPool(
		WithContext(context.Background()),
		WithIdleTimeout(3*time.Second),
		WithPoolSize(2),
	)

	done := workerPool.Start()
	workerPool.Start()

	<-done
}

func TestLifecycle(t *testing.T) {
	workerPool := NewWorkerPool(
		WithContext(context.Background()),
		WithIdleTimeout(3*time.Second),
		WithPoolSize(2),
	)

	done := workerPool.Start()
	workerPool.AddJobs(
		func(ctx context.Context) error {
			fmt.Println("job 1 running")
			time.Sleep(200 * time.Millisecond)
			return nil
		},
		func(ctx context.Context) error {
			fmt.Println("job 2 running")
			time.Sleep(2 * time.Second)
			return errors.New("this is error")
		},
		func(ctx context.Context) error {
			fmt.Println("job 3 running")
			time.Sleep(1 * time.Second)
			return nil
		},
	)

	time.Sleep(time.Second)

	for i := 0; i < 10; i++ {
		i := i
		workerPool.AddJobs(
			func(ctx context.Context) error {
				fmt.Printf("job %d running\n", i)
				return nil
			},
		)
	}
	<-done
}

func BenchmarkWorkers(b *testing.B) {
	workerPool := NewWorkerPool(
		WithContext(context.Background()),
		WithIdleTimeout(3*time.Second),
		WithPoolSize(2),
	)

	workerPool.Start()

	for i := 0; i < b.N; i++ {
		workerPool.AddJobs(
			func(ctx context.Context) error {
				return nil
			},
		)

	}
	workerPool.Stop()
}
