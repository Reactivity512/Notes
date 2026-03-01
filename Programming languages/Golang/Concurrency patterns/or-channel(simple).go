package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Task — простая задача с ID и временем выполнения
type Task struct {
	ID       int
	Duration time.Duration
}

// worker выполняет задачу и отправляет результат в канал
func worker(ctx context.Context, task Task, result chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("Воркер %d начал работу (длительность: %v)\n", task.ID, task.Duration)

	// Имитация работы с возможностью отмены
	select {
	case <-time.After(task.Duration):
		// Работа успешно завершена
		result <- fmt.Sprintf("Воркер %d завершён", task.ID)
	case <-ctx.Done():
		// Контекст отменен — прерываем работу
		fmt.Printf("Воркер %d отменён: %v\n", task.ID, ctx.Err())
		return
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Создаем отменяемый контекст
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Освобождаем ресурсы

	result := make(chan string, 3) // Буферизированный канал для 3 результатов
	var wg sync.WaitGroup

	// Запускаем 3 воркера с разным временем выполнения
	tasks := []Task{
		{ID: 1, Duration: 3 * time.Second},
		{ID: 2, Duration: 1 * time.Second}, // Самый быстрый
		{ID: 3, Duration: 5 * time.Second},
	}

	for _, task := range tasks {
		wg.Add(1)
		go worker(ctx, task, result, &wg)
	}

	// Ждём ПЕРВЫЙ ответ
	firstResult := <-result
	fmt.Printf("\nПервый ответ: %s\n\n", firstResult)

	// Отменяем контекст — остальные воркеры получат сигнал и завершатся
	cancel()

	// Ждём, пока все воркеры корректно завершат работу (обработают отмену)
	wg.Wait()
	close(result)

	fmt.Println("Все воркеры завершены")
}
