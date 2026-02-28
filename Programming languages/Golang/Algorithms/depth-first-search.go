package main

import "fmt"

// Граф представлен как map узла к списку соседей
type Graph map[int][]int

func DFSRecursive(graph Graph, start int, visited map[int]bool) {
	// Отмечаем текущий узел как посещенный
	visited[start] = true
	fmt.Printf("Посетили узел: %d\n", start)

	// Рекурсивно обходим всех непосещенных соседей
	for _, neighbor := range graph[start] {
		if !visited[neighbor] {
			DFSRecursive(graph, neighbor, visited)
		}
	}
}

func DFSIterative(graph Graph, start int) {
	visited := make(map[int]bool)
	stack := []int{start} // Стек для DFS

	for len(stack) > 0 {
		// Извлекаем последний элемент (LIFO)
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited[node] {
			continue
		}

		visited[node] = true
		fmt.Printf("Посетили узел: %d\n", node)

		// Добавляем соседей в стек
		for _, neighbor := range graph[node] {
			if !visited[neighbor] {
				stack = append(stack, neighbor)
			}
		}
	}
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

	fmt.Println("DFS рекурсивный обход:")
	DFSIterative(graph, 1)
}
