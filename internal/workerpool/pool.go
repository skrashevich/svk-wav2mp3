package workerpool

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Task представляет задачу для воркера.
type Task interface {
	Execute() error
}

// TaskResult содержит результат выполнения задачи.
type TaskResult struct {
	Task   Task
	Output any
	Error  error
}

// WorkerPool управляет пулом воркеров для параллельного выполнения задач.
type WorkerPool struct {
	numWorkers int
	taskQueue  chan Task
	results    chan TaskResult
	workers    []chan Task
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	done       chan struct{}
	Stats      *PoolStats
}

// PoolStats содержит статистику пула воркеров.
type PoolStats struct {
	TotalTasks  int
	Completed   int
	Failed      int
	Processing  int
	ElapsedTime int64 // в микросекундах
}

// NewWorkerPool создает новый пул воркеров.
func NewWorkerPool(ctx context.Context, numWorkers int, bufferSize int) *WorkerPool {
	if numWorkers <= 0 {
		numWorkers = 4 // default
	}
	if bufferSize <= 0 {
		bufferSize = 100 // default
	}

	ctx, cancel := context.WithCancel(ctx)
	wp := &WorkerPool{
		numWorkers: numWorkers,
		taskQueue:  make(chan Task, bufferSize),
		results:    make(chan TaskResult, bufferSize),
		workers:    make([]chan Task, numWorkers),
		ctx:        ctx,
		cancel:     cancel,
		done:       make(chan struct{}),
		Stats:      &PoolStats{},
	}

	// Запускаем воркеров
	wp.startWorkers()

	return wp
}

// Submit добавляет задачу в пул для выполнения.
func (wp *WorkerPool) Submit(task Task) error {
	select {
	case wp.taskQueue <- task:
		wp.Stats.TotalTasks++
		wp.Stats.Processing++
		return nil
	case <-wp.ctx.Done():
		return errors.New("pool is shutting down")
	}
}

// GetResults возвращает канал результатов.
func (wp *WorkerPool) GetResults() <-chan TaskResult {
	return wp.results
}

// Shutdown gracefully shuts down the pool.
func (wp *WorkerPool) Shutdown() {
	wp.cancel()
	close(wp.taskQueue)
	wp.wg.Wait()
	close(wp.done)
}

// Close закрывает результаты.
func (wp *WorkerPool) Close() {
	close(wp.results)
}

// IsDone возвращает true если все задачи завершены.
func (wp *WorkerPool) IsDone() bool {
	select {
	case <-wp.done:
		return true
	default:
		return false
	}
}

// startWorkers запускает воркеров.
func (wp *WorkerPool) startWorkers() {
	for i := 0; i < wp.numWorkers; i++ {
		wp.workers[i] = make(chan Task)
		wp.wg.Add(1)

		go func(id int) {
			defer wp.wg.Done()
			wp.workerLoop(id)
		}(i)
	}
}

// workerLoop обрабатывает задачи для одного воркера.
func (wp *WorkerPool) workerLoop(workerID int) {
	for task := range wp.workers[workerID] {
		// Проверяем cancelled
		select {
		case <-wp.ctx.Done():
			wp.Stats.Failed++
			wp.Stats.Processing--
			wp.results <- TaskResult{Task: task, Error: errors.New("worker cancelled")}
			continue
		default:
		}

		startTime := time.Now()
		err := task.Execute()
		elapsed := time.Since(startTime)

		wp.Stats.Processing--
		wp.Stats.Completed++

		wp.results <- TaskResult{
			Task:   task,
			Error:  err,
			Output: elapsed,
		}
	}
}
