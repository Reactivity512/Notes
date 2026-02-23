package main

import (
	"context"
	"fmt"
	"time"
)

func worker(ctx context.Context, name string) {
	for {
		select {
		case <-time.After(500 * time.Millisecond):
			fmt.Printf("%s: работаю\n", name)
		case <-ctx.Done():
			// Получили сигнал об отмене
			fmt.Printf("%s: остановлен, причина: %v\n", name, ctx.Err())
			return
		}
	}
}

func main() {
	// 1. Создаём родительский контекст с отменой
	parentCtx, cancel := context.WithCancel(context.Background())

	// 2. Запускаем воркера
	go worker(parentCtx, "воркер1")

	// 3. Создаём дочерний контекст с таймаутом
	childCtx, childCancel := context.WithTimeout(parentCtx, 1*time.Second)
	defer childCancel()
	go worker(childCtx, "воркер2")

	// 4. Даём поработать 2 секунды
	time.Sleep(2 * time.Second)

	// 5. Отменяем родительский контекст
	fmt.Println("Главный: отменяю контекст")
	cancel()

	// 6. Ждём завершения горутин
	time.Sleep(500 * time.Millisecond)
	fmt.Println("Главный: завершился")
}
