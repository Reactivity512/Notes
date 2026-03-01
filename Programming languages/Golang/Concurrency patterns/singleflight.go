package main

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

func main() {
	var group singleflight.Group
	var wg sync.WaitGroup

	key := "user:123"
	requests := 5

	// Симулируем тяжелую операцию (поход в БД или внешнее API)
	expensiveOp := func() (interface{}, error) {
		fmt.Println("Тяжелая операция началась (реально один раз)")
		time.Sleep(2 * time.Second) // имитация долгой работы
		fmt.Println("Тяжелая операция завершилась")
		return "data_for_" + key, nil
	}

	// Запускаем 5 конкурентных запросов
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func(reqID int) {
			defer wg.Done()

			// Выполняем через singleflight
			result, err, shared := group.Do(key, expensiveOp)

			if err != nil {
				fmt.Printf("Запрос %d: ошибка %v\n", reqID, err)
				return
			}
			fmt.Printf("Запрос %d: результат = %v (shared = %v)\n",
				reqID, result, shared)
		}(i)
	}

	wg.Wait()
}
