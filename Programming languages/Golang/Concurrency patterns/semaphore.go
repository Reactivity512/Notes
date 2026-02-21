package main

import (
	"fmt"
	"sync"
	"time"
)

// Semaphore - ограничитель конкурентности
type Semaphore struct {
	ch chan struct{}
}

func NewSemaphore(maxConcurrent int) *Semaphore {
	return &Semaphore{
		ch: make(chan struct{}, maxConcurrent),
	}
}

// Acquire - получаем разрешение (блокирует если лимит исчерпан)
func (s *Semaphore) Acquire() {
	s.ch <- struct{}{} // Отправка блокируется при заполнении канала
}

// Release - возвращаем разрешение в пул
func (s *Semaphore) Release() {
	<-s.ch // Освобождаем слот
}

func worker(id int, sem *Semaphore, wg *sync.WaitGroup) {
	defer wg.Done()

	sem.Acquire()       // Ждём свободный слот
	defer sem.Release() // Освобождаем после работы

	fmt.Printf("[%s] Worker %d: начал работу\n", time.Now().Format("15:04:05"), id)
	time.Sleep(2 * time.Second) // Имитация работы
	fmt.Printf("[%s] Worker %d: завершил\n", time.Now().Format("15:04:05"), id)
}

func main() {
	sem := NewSemaphore(3) // Максимум 3 одновременных задачи
	var wg sync.WaitGroup

	// Запускаем 10 горутин, но активными будут только 3
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go worker(i, sem, &wg)
		time.Sleep(200 * time.Millisecond) // Небольшая задержка между запусками
	}

	wg.Wait()
	fmt.Println("Все задачи выполнены")
}
