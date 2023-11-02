package goworker

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestRunTaskBeforeStarting(t *testing.T) {
	pool := NewWorkerPool(
		WithContext(context.Background()),
		WithIdleTimeout(3*time.Second),
		WithPoolSize(2),
	)

	pool.AddJobs(
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
	pool := NewWorkerPool(
		WithContext(context.Background()),
		WithIdleTimeout(3*time.Second),
		WithPoolSize(2),
	)

	done := pool.Start()
	pool.Start()

	<-done
}

func TestLifecycle(t *testing.T) {
	pool := NewWorkerPool(
		WithContext(context.Background()),
		WithIdleTimeout(3*time.Second),
		WithPoolSize(2),
	)

	done := pool.Start()
	pool.AddJobs(
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
		pool.AddJobs(
			func(ctx context.Context) error {
				fmt.Printf("job %d running\n", i)
				return nil
			},
		)
	}

	<-done
}

func TestParallelRun(t *testing.T) {
	pool := NewWorkerPool(
		WithContext(context.Background()),
		WithIdleTimeout(3*time.Second),
		WithPoolSize(5),
	)

	done := pool.Start()
	t.Parallel()
	for i := 0; i < 10; i++ {
		i := i
		pool.AddJobs(
			func(ctx context.Context) error {
				fmt.Printf("job %d running\n", i)
				return nil
			},
		)
	}

	<-done
}
