package main

import (
	"context"
	"fmt"
	"sync"
)

// Генерация задач
func generateJobs(n int) <-chan int {
	ch := make(chan int)
	go func() {
		for i := 1; i <= n; i++ {
			ch <- i
		}
		close(ch)
	}()
	return ch
}

// Fan-out: Распределение задач между воркерами
func fanOut(ctx context.Context, jobs <-chan int, numWorkers int) []<-chan int {
	workerChannels := make([]<-chan int, 0, numWorkers)

	for i := 0; i < numWorkers; i++ {
		resultCh := make(chan int)

		go func() {
			defer close(resultCh)
			for {
				select {
				case job, ok := <-jobs:
					if !ok {
						return // Канал задач закрыт
					}
					// Обработка задачи (пример: возведение в квадрат)
					resultCh <- job * job
				case <-ctx.Done():
					return // Отмена через контекст
				}
			}
		}()

		workerChannels = append(workerChannels, resultCh)
	}
	return workerChannels
}

// Fan-in: Объединение результатов
func fanIn(channels []<-chan int) <-chan int {
	var wg sync.WaitGroup
	merged := make(chan int)

	wg.Add(len(channels))

	for _, ch := range channels {
		go func(c <-chan int) {
			defer wg.Done()
			for res := range c {
				merged <- res
			}
		}(ch)
	}

	// Горутина для закрытия итогового канала
	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged
}

func main() {
	// 1. Генерируем задачи
	jobs := generateJobs(5)

	// 2. Fan-out: распределяем задачи между 3 воркерами
	resultChannels := fanOut(context.Background(), jobs, 3)

	// 3. Fan-in: объединяем результаты из всех каналов
	mergedResults := fanIn(resultChannels)

	// 4. Результаты
	for res := range mergedResults {
		fmt.Printf("Получен результат: %d\n", res)
	}
}
