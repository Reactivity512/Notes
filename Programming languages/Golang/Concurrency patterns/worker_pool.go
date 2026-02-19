package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// Task представляет единицу работы
type Task struct {
	// ID задачи для логирования и трекинга
	ID string

	// Payload - сами данные для обработки
	Payload interface{}

	// Обработчик задачи (функция, которую нужно выполнить)
	Handler func(ctx context.Context, payload interface{}) (interface{}, error)
}

// Result представляет результат выполнения задачи
type Result struct {
	TaskID string
	Output interface{}
	Err    error
}

// WorkerPool - основная структура пула
type WorkerPool struct {
	// Конфигурация
	numWorkers int
	queueSize  int

	// Каналы
	tasksCh   chan Task
	resultsCh chan Result

	// Управление жизненным циклом
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Состояние
	started int32 // atomic flag
	stopped int32 // atomic flag

	// Статистика (опционально, для мониторинга)
	tasksSubmitted int64
	tasksCompleted int64
	tasksFailed    int64

	// Защита от паник
	panicHandler func(interface{})

	// Канал для сигнала о завершении всех задач
	doneCh chan struct{}
}

// NewWorkerPool создает новый пул воркеров
func NewWorkerPool(numWorkers int, queueSize int) *WorkerPool {
	if numWorkers <= 0 {
		numWorkers = 1
	}
	if queueSize <= 0 {
		queueSize = 100
	}

	return &WorkerPool{
		numWorkers:   numWorkers,
		queueSize:    queueSize,
		tasksCh:      make(chan Task, queueSize),
		resultsCh:    make(chan Result, queueSize),
		doneCh:       make(chan struct{}),
		panicHandler: defaultPanicHandler,
	}
}

// SetPanicHandler устанавливает кастомный обработчик паник
func (wp *WorkerPool) SetPanicHandler(handler func(interface{})) {
	wp.panicHandler = handler
}

// defaultPanicHandler - обработчик паник по умолчанию
func defaultPanicHandler(r interface{}) {
	fmt.Printf("Worker recovered from panic: %v\n", r)
}

// Start запускает воркеров
func (wp *WorkerPool) Start() error {
	if !atomic.CompareAndSwapInt32(&wp.started, 0, 1) {
		return fmt.Errorf("worker pool already started")
	}

	// Создаем контекст для graceful shutdown
	wp.ctx, wp.cancel = context.WithCancel(context.Background())

	// Запускаем воркеров
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	// Запускаем мониторинг завершения
	go wp.monitor()

	return nil
}

// worker - основная логика воркера
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	fmt.Printf("Worker %d started\n", id)

	for {
		select {
		case <-wp.ctx.Done():
			// Получили сигнал на завершение
			fmt.Printf("Worker %d shutting down\n", id)
			return

		case task, ok := <-wp.tasksCh:
			if !ok {
				// Канал задач закрыт
				fmt.Printf("Worker %d: tasks channel closed\n", id)
				return
			}

			// Обрабатываем задачу
			wp.processTask(task)
		}
	}
}

// processTask обрабатывает одну задачу с защитой от паник
func (wp *WorkerPool) processTask(task Task) {
	atomic.AddInt64(&wp.tasksSubmitted, 1)

	// Канал для результата внутри задачи (чтобы не блокировать воркер на панике)
	resultCh := make(chan Result, 1)

	// Запускаем обработку в отдельной горутине для изоляции паник
	go func() {
		defer func() {
			if r := recover(); r != nil {
				wp.panicHandler(r)
				resultCh <- Result{
					TaskID: task.ID,
					Err:    fmt.Errorf("panic in task handler: %v", r),
				}
			}
		}()

		// Выполняем хендлер задачи с контекстом
		output, err := task.Handler(wp.ctx, task.Payload)
		resultCh <- Result{
			TaskID: task.ID,
			Output: output,
			Err:    err,
		}
	}()

	// Ждем результат или отмены контекста
	select {
	case <-wp.ctx.Done():
		// Контекст отменен во время выполнения задачи
		wp.sendResult(Result{
			TaskID: task.ID,
			Err:    fmt.Errorf("task cancelled: %w", wp.ctx.Err()),
		})

	case result := <-resultCh:
		wp.sendResult(result)
	}
}

// sendResult отправляет результат и обновляет статистику
func (wp *WorkerPool) sendResult(result Result) {
	atomic.AddInt64(&wp.tasksCompleted, 1)
	if result.Err != nil {
		atomic.AddInt64(&wp.tasksFailed, 1)
	}

	// Неблокирующая отправка результата
	select {
	case wp.resultsCh <- result:
		// Результат отправлен
	default:
		// Канал результатов переполнен - логируем проблему
		fmt.Printf("Warning: results channel full, dropping result for task %s\n", result.TaskID)
	}
}

// Submit отправляет задачу в пул (блокирующая версия)
func (wp *WorkerPool) Submit(task Task) error {
	if atomic.LoadInt32(&wp.stopped) == 1 {
		return fmt.Errorf("worker pool already stopped")
	}

	if atomic.LoadInt32(&wp.started) == 0 {
		return fmt.Errorf("worker pool not started")
	}

	select {
	case wp.tasksCh <- task:
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	}
}

// SubmitWithTimeout отправляет задачу с таймаутом
func (wp *WorkerPool) SubmitWithTimeout(task Task, timeout <-chan time.Time) error {
	select {
	case wp.tasksCh <- task:
		return nil
	case <-timeout:
		return fmt.Errorf("submit timeout")
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	}
}

// TrySubmit пытается отправить задачу неблокирующим образом
func (wp *WorkerPool) TrySubmit(task Task) error {
	select {
	case wp.tasksCh <- task:
		return nil
	default:
		return fmt.Errorf("task queue is full")
	}
}

// Results возвращает канал результатов
func (wp *WorkerPool) Results() <-chan Result {
	return wp.resultsCh
}

// Stop останавливает пул и ждет завершения всех задач
func (wp *WorkerPool) Stop() error {
	if !atomic.CompareAndSwapInt32(&wp.stopped, 0, 1) {
		return fmt.Errorf("worker pool already stopped")
	}

	// Отменяем контекст, чтобы воркеры поняли, что пора закругляться
	if wp.cancel != nil {
		wp.cancel()
	}

	// Ждем завершения всех воркеров
	wp.wg.Wait()

	// Закрываем каналы после завершения всех воркеров
	close(wp.tasksCh)
	close(wp.resultsCh)

	// Сигнализируем о полном завершении
	close(wp.doneCh)

	return nil
}

// StopAndWait останавливает пул и ждет обработки всех оставшихся задач
func (wp *WorkerPool) StopAndWait() error {
	// Закрываем канал задач - новые задачи не принимаем
	close(wp.tasksCh)

	// Ждем, пока воркеры дообработают все задачи в очереди
	wp.wg.Wait()

	// Отменяем контекст для освобождения ресурсов
	if wp.cancel != nil {
		wp.cancel()
	}

	// Закрываем канал результатов
	close(wp.resultsCh)
	close(wp.doneCh)

	atomic.StoreInt32(&wp.stopped, 1)

	return nil
}

// Done возвращает канал, который закрывается при полной остановке пула
func (wp *WorkerPool) Done() <-chan struct{} {
	return wp.doneCh
}

// monitor следит за состоянием пула (опционально)
func (wp *WorkerPool) monitor() {
	<-wp.ctx.Done()
	// Можно добавить логику мониторинга, например, вывод статистики
	fmt.Printf("Pool shutting down. Stats: submitted=%d, completed=%d, failed=%d\n",
		atomic.LoadInt64(&wp.tasksSubmitted),
		atomic.LoadInt64(&wp.tasksCompleted),
		atomic.LoadInt64(&wp.tasksFailed))
}

// Stats возвращает текущую статистику
func (wp *WorkerPool) Stats() (submitted, completed, failed int64) {
	return atomic.LoadInt64(&wp.tasksSubmitted),
		atomic.LoadInt64(&wp.tasksCompleted),
		atomic.LoadInt64(&wp.tasksFailed)
}

// IsRunning проверяет, запущен ли пул
func (wp *WorkerPool) IsRunning() bool {
	return atomic.LoadInt32(&wp.started) == 1 && atomic.LoadInt32(&wp.stopped) == 0
}

func main() {
	// Создаем пул с 5 воркерами и очередью на 100 задач
	pool := NewWorkerPool(5, 100)

	// Запускаем пул
	if err := pool.Start(); err != nil {
		log.Fatal(err)
	}

	// Гарантируем остановку пула при выходе
	defer pool.Stop()

	// Запускаем потребитель результатов
	go func() {
		for result := range pool.Results() {
			if result.Err != nil {
				fmt.Printf("Task %s failed: %v\n", result.TaskID, result.Err)
			} else {
				fmt.Printf("Task %s completed: %v\n", result.TaskID, result.Output)
			}
		}
	}()

	// Отправляем задачи
	for i := 0; i < 20; i++ {
		taskID := fmt.Sprintf("task-%d", i)
		task := Task{
			ID:      taskID,
			Payload: i,
			Handler: func(ctx context.Context, payload interface{}) (interface{}, error) {
				num := payload.(int)

				// Имитация работы
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(time.Duration(num%3+1) * time.Second):
					// Успешное выполнение
				}

				// Симулируем ошибку для некоторых задач
				if num%5 == 0 {
					return nil, fmt.Errorf("simulated error for %d", num)
				}

				return num * 2, nil
			},
		}

		// Пробуем отправить с таймаутом
		if err := pool.SubmitWithTimeout(task, time.After(100*time.Millisecond)); err != nil {
			fmt.Printf("Failed to submit %s: %v\n", taskID, err)
		}
	}

	// Даем время на обработку
	time.Sleep(10 * time.Second)

	// Останавливаем пул с ожиданием
	if err := pool.StopAndWait(); err != nil {
		log.Fatal(err)
	}

	// Ждем полной остановки
	<-pool.Done()
	fmt.Println("Pool stopped completely")
}
