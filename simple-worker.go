package goworker

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

type (
	WorkerPool interface {
		Start() chan struct{}
		AddJobs(jobs ...JobRunner)
		Stop()
	}

	workerPool struct {
		poolSize    int32
		running     *atomic.Bool
		quiting     chan chan struct{}
		done        chan struct{}
		jobPool     chan JobRunner
		idleTimeout time.Duration
		ctx         context.Context
		ctxCancel   context.CancelFunc
	}

	JobRunner func(ctx context.Context) error
)

func NewWorkerPool(options ...Options) WorkerPool {
	wp := &workerPool{}
	for _, option := range options {
		option(wp)
	}

	if wp.ctx == nil {
		ctx, cancelFunc := context.WithCancel(context.Background())
		wp.ctx = ctx
		wp.ctxCancel = cancelFunc
	}

	if wp.poolSize <= 0 {
		wp.poolSize = 4
	}

	if wp.idleTimeout == 0 {
		wp.idleTimeout = 1 * time.Minute
	}

	wp.running = &atomic.Bool{}
	wp.jobPool = make(chan JobRunner, wp.poolSize)
	wp.quiting = make(chan chan struct{}, 1)
	wp.done = make(chan struct{})

	return wp
}

func (wp workerPool) Start() chan struct{} {
	if wp.running.Load() {
		fmt.Println("working already started")
		return nil
	}

	wp.running.Store(true)

	go func() {
		for {
			select {
			case job := <-wp.jobPool:
				err := job(wp.ctx)
				if err != nil {
					fmt.Println(err)
				}
			case <-time.After(wp.idleTimeout):
				fmt.Println("pool idle...stopping")
				wp.Stop()
			case <-wp.ctx.Done():
				fmt.Println("context finalized...stopping")
				wp.Stop()
			case quit := <-wp.quiting:
				fmt.Println("stopping worker")
				wp.running.Store(false)
				quit <- struct{}{}
				return
			}
		}
	}()

	return wp.done
}

func (wp workerPool) AddJobs(jobs ...JobRunner) {
	if !wp.running.Load() {
		fmt.Println("worker not running")
		return
	}

	for _, job := range jobs {
		wp.jobPool <- job
	}
}

func (wp workerPool) Stop() {
	if !wp.running.Load() {
		fmt.Println("worker not running")
		return
	}

	quit := make(chan struct{})
	go func(q chan struct{}) {
		<-q
		fmt.Println("stop finalized... closing channels")
		wp.closeChannels(q)
		wp.ctxCancel()
		wp.done <- struct{}{}
	}(quit)
	wp.quiting <- quit
}

func (wp workerPool) closeChannels(quit chan struct{}) {
	close(quit)
	close(wp.quiting)
	close(wp.jobPool)
}

type (
	Options func(wp *workerPool)
)

func WithPoolSize(size int) Options {
	return func(wp *workerPool) {
		wp.poolSize = int32(size)
	}
}

func WithIdleTimeout(timeout time.Duration) Options {
	return func(wp *workerPool) {
		wp.idleTimeout = timeout
	}
}

func WithContext(ctx context.Context) Options {
	return func(wp *workerPool) {
		wpCtx, cancelFunc := context.WithCancel(ctx)
		wp.ctx = wpCtx
		wp.ctxCancel = cancelFunc
	}
}
