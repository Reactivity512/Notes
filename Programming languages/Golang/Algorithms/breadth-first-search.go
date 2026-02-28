package main

import (
	"fmt"
)

// Граф представлен как map узла к списку соседей
type Graph map[int][]int

func BFS(graph Graph, start int) {
	visited := make(map[int]bool)
	queue := []int{start} // Очередь для BFS

	visited[start] = true

	for len(queue) > 0 {
		// Извлекаем первый элемент (FIFO)
		node := queue[0]
		queue = queue[1:]

		fmt.Printf("Посетили узел: %d\n", node)

		// Добавляем всех непосещенных соседей в очередь
		for _, neighbor := range graph[node] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}
}

type Node struct {
	ID   int
	Dist int
	Path []int
}

func BFSShortestPath(graph Graph, start, target int) ([]int, int) {
	visited := make(map[int]bool)
	queue := []Node{{ID: start, Dist: 0, Path: []int{start}}}
	visited[start] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Нашли цель
		if current.ID == target {
			return current.Path, current.Dist
		}

		// Добавляем соседей
		for _, neighbor := range graph[current.ID] {
			if !visited[neighbor] {
				visited[neighbor] = true
				newPath := make([]int, len(current.Path))
				copy(newPath, current.Path)
				newPath = append(newPath, neighbor)

				queue = append(queue, Node{
					ID:   neighbor,
					Dist: current.Dist + 1,
					Path: newPath,
				})
			}
		}
	}

	return nil, -1 // Путь не найден
}

func main() {
	graph := Graph{
		1: {2, 3, 4},
		2: {5, 6},
		3: {7},
		4: {8},
		5: {},
		6: {},
		7: {},
		8: {},
	}

	/*fmt.Println("BFS обход:")
	  BFS(graph, 1)*/

	fmt.Println("BFS поиска кратчайшего пути:")
	path, dist := BFSShortestPath(graph, 1, 8)
	fmt.Printf("Path: %d, Dist:%d", path, dist)
}
