package monitoring

import (
	"context"
	"log"
	"sync"
	"time"
)

// Task represents a monitoring task
type Task struct {
	DeviceID   int64
	DeviceType string
	IPAddress  string
	Execute    func(ctx context.Context) error
}

// Result represents a monitoring result
type Result struct {
	DeviceID  int64
	Status    string // "online", "offline", "unknown"
	Latency   time.Duration
	Error     error
	Timestamp time.Time
	Details   map[string]interface{}
}

// WorkerPool manages a pool of workers for monitoring tasks
type WorkerPool struct {
	workers    int
	taskQueue  chan Task
	results    chan Result
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	onResult   func(Result)
	running    bool
	mu         sync.Mutex
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers int, onResult func(Result)) *WorkerPool {
	pool := &WorkerPool{
		workers:  workers,
		onResult: onResult,
	}

	return pool
}

// Start starts the worker pool
func (p *WorkerPool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return
	}

	// Create new channels and context for this session
	p.ctx, p.cancel = context.WithCancel(context.Background())
	p.taskQueue = make(chan Task, 100)
	p.results = make(chan Result, 100)
	p.running = true

	// Start workers
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	// Start result processor
	go p.processResults()

	log.Printf("Worker pool started with %d workers", p.workers)
}

// Stop stops the worker pool
func (p *WorkerPool) Stop() {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return
	}
	p.running = false
	p.mu.Unlock()

	p.cancel()
	close(p.taskQueue)
	p.wg.Wait()
	close(p.results)
	log.Println("Worker pool stopped")
}

// Submit submits a task to the pool
func (p *WorkerPool) Submit(task Task) {
	select {
	case p.taskQueue <- task:
	case <-p.ctx.Done():
		return
	default:
		log.Printf("Task queue full, dropping task for device %d", task.DeviceID)
	}
}

// worker processes tasks from the queue
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case task, ok := <-p.taskQueue:
			if !ok {
				return
			}

			start := time.Now()

			// Create context with timeout
			ctx, cancel := context.WithTimeout(p.ctx, 10*time.Second)

			err := task.Execute(ctx)
			latency := time.Since(start)

			cancel()

			status := "online"
			if err != nil {
				status = "offline"
			}

			p.results <- Result{
				DeviceID:  task.DeviceID,
				Status:    status,
				Latency:   latency,
				Error:     err,
				Timestamp: time.Now(),
			}

		case <-p.ctx.Done():
			return
		}
	}
}

// processResults handles results from workers
func (p *WorkerPool) processResults() {
	for result := range p.results {
		if p.onResult != nil {
			p.onResult(result)
		}
	}
}
