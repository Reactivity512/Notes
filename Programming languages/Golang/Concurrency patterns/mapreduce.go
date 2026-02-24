package main

import (
	"fmt"
	"strings"
	"sync"
)

func main() {
	// Входные данные: большой набор документов (строк)
	documents := []string{
		"hello world",
		"hello Go",
		"world of Go",
		"hello again",
	}

	// 1. MAP ФАЗА: Запускаем воркеров для параллельной обработки документов
	type mapResult struct {
		word  string
		count int
	}

	mapChan := make(chan mapResult, 100) // Канал для сбора результатов Map
	var mapWg sync.WaitGroup

	for _, doc := range documents {
		mapWg.Add(1)
		go func(text string) {
			defer mapWg.Done()
			// Map функция: разбивает документ на слова и считает каждое как 1
			words := strings.Fields(text)
			for _, word := range words {
				mapChan <- mapResult{word: word, count: 1}
			}
		}(doc)
	}

	// Закрываем канал mapChan после завершения всех мапперов
	go func() {
		mapWg.Wait()
		close(mapChan)
	}()

	// 2. SHUFFle ФАЗА (упрощенно): Группируем результаты по ключу (слову)
	intermediate := make(map[string][]int)
	for res := range mapChan {
		intermediate[res.word] = append(intermediate[res.word], res.count)
	}

	// 3. REDUCE ФАЗА: Для каждого ключа запускаем редьюсер (может быть параллельно)
	type finalResult struct {
		word  string
		total int
	}
	finalChan := make(chan finalResult, len(intermediate))
	var reduceWg sync.WaitGroup

	for word, counts := range intermediate {
		reduceWg.Add(1)
		go func(w string, c []int) {
			defer reduceWg.Done()
			// Reduce функция: суммирует все значения (counts) для данного ключа (word)
			sum := 0
			for _, val := range c {
				sum += val
			}
			finalChan <- finalResult{word: w, total: sum}
		}(word, counts)
	}

	// Закрываем финальный канал
	go func() {
		reduceWg.Wait()
		close(finalChan)
	}()

	// Собираем и выводим результат
	for res := range finalChan {
		fmt.Printf("Слово '%s' встречается %d раз(а)\n", res.word, res.total)
	}
}
