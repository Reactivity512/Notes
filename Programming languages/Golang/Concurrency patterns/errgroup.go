package main

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	// Создаем errgroup с контекстом
	g, ctx := errgroup.WithContext(context.Background())

	// Запускаем горутину, которая успешно выполнится
	g.Go(func() error {
		select {
		case <-time.After(1 * time.Second):
			fmt.Println("Горутина 1: успех")
			return nil
		case <-ctx.Done():
			fmt.Println("Горутина 1: отменена")
			return ctx.Err()
		}
	})

	// Запускаем горутину, которая завершится ошибкой
	g.Go(func() error {
		select {
		case <-time.After(500 * time.Millisecond):
			fmt.Println("Горутина 2: ошибка!")
			return fmt.Errorf("что-то пошло не так в горутине 2")
		case <-ctx.Done():
			fmt.Println("Горутина 2: отменена")
			return ctx.Err()
		}
	})

	// Запускаем горутину, которая должна быть отменена
	g.Go(func() error {
		select {
		case <-time.After(3 * time.Second):
			fmt.Println("Горутина 3: успех (но этого не случится)")
			return nil
		case <-ctx.Done():
			fmt.Println("Горутина 3: отменена")
			return ctx.Err() // context.Canceled
		}
	})

	// Ждем завершения всех горутин и получаем первую ошибку
	if err := g.Wait(); err != nil {
		fmt.Printf("Ошибка: %v\n", err)
	} else {
		fmt.Println("Все успешно")
	}

	// Даем время увидеть вывод
	time.Sleep(1 * time.Second)
}
