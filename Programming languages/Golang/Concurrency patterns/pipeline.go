package main

import (
	"bufio"
	"fmt"
	"os"
)

// stage 1: Генератор
func gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		for _, n := range nums {
			out <- n
		}
		close(out)
	}()
	return out
}

// stage 2: Возведение в квадрат
func sq(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * n
		}
		close(out)
	}()
	return out
}

// stage 3: Сохранение в файл
func saveToFile(in <-chan int, filename string) <-chan error {
	errCh := make(chan error, 1) // буферизированный, чтобы избежать блокировки

	go func() {
		defer close(errCh)

		file, err := os.Create(filename)
		if err != nil {
			errCh <- fmt.Errorf("failed to create file: %w", err)
			return
		}
		defer file.Close()

		writer := bufio.NewWriter(file)
		defer writer.Flush()

		for val := range in {
			if _, err := fmt.Fprintf(writer, "%d\n", val); err != nil {
				errCh <- fmt.Errorf("failed to write to file: %w", err)
				return
			}
		}
	}()

	return errCh
}

func main() {
	// Сборка конвейера
	in := gen(2, 3, 4)
	squared := sq(in)
	errCh := saveToFile(squared, "output.txt")

	// Ожидаем завершения записи и проверяем ошибки
	if err := <-errCh; err != nil {
		fmt.Fprintf(os.Stderr, "Pipeline error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Data successfully saved to output.txt")
}
