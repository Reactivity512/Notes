package main

import (
	"fmt"
	"time"
)

func main() {
	// or — функция, объединяющая каналы
	var or func(channels ...<-chan interface{}) <-chan interface{}

	or = func(channels ...<-chan interface{}) <-chan interface{} {
		// Базовые случаи рекурсии
		switch len(channels) {
		case 0:
			return nil
		case 1:
			return channels[0]
		}

		// Создаём объединённый канал для этого уровня
		orDone := make(chan interface{})

		// Запускаем горутину, которая будет ждать первый сигнал
		go func() {
			defer close(orDone)

			switch len(channels) {
			case 2:
				// Для двух каналов — простой select
				select {
				case <-channels[0]:
				case <-channels[1]:
				}
			default:
				// Для трёх и более — слушаем первые три напрямую,
				// а для остальных запускаем рекурсию
				select {
				case <-channels[0]:
				case <-channels[1]:
				case <-channels[2]:
				case <-or(append(channels[3:], orDone)...):
				}
			}
		}()

		return orDone
	}

	// sig создаёт канал, который закроется через заданное время
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	// Засекаем время
	start := time.Now()

	// Объединяем 5 каналов с разными временами ожидания
	<-or(
		sig(2*time.Hour),   // закроется через 2 часа
		sig(5*time.Minute), // через 5 минут
		sig(1*time.Second), // через 1 секунду (самый быстрый!)
		sig(1*time.Hour),   // через 1 час
		sig(1*time.Minute), // через 1 минуту
	)

	// Программа завершится через ~1 секунду, несмотря на остальные каналы
	fmt.Printf("Done after %v\n", time.Since(start))
}
