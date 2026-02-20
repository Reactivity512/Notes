package main

import (
	"fmt"
	"sync"
	"time"
)

type Job struct {
	ID   int
	Data string
}

type Result struct {
	JobID    int
	Output   string
	Duration time.Duration
}

func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		start := time.Now()

		// Имитация работы
		time.Sleep(100 * time.Millisecond)

		results <- Result{
			JobID:    job.ID,
			Output:   fmt.Sprintf("Processed by worker-%d", id),
			Duration: time.Since(start),
		}
	}
}

func main() {
	const numWorkers = 3
	const numJobs = 10

	jobs := make(chan Job, numJobs)
	results := make(chan Result, numJobs)
	var wg sync.WaitGroup

	// Запускаем пул воркеров
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i, jobs, results, &wg)
	}

	// Отправляем задачи
	go func() {
		for i := 1; i <= numJobs; i++ {
			jobs <- Job{ID: i, Data: fmt.Sprintf("task-%d", i)}
		}
		close(jobs)
	}()

	// Собираем результаты в отдельной горутине
	go func() {
		wg.Wait()
		close(results)
	}()

	// Читаем результаты
	for res := range results {
		fmt.Printf("Job %d: %s (took %v)\n", res.JobID, res.Output, res.Duration)
	}
}
