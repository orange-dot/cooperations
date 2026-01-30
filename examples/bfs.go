package main

import (
	"fmt"
	"container/list"
)

// Graph represents a graph structure with nodes and edges
type Graph struct {
	vertices map[int]*Vertex
}

// Vertex represents a single node in the graph with its connections (edges)
type Vertex struct {
	key      int
	adjacent []*Vertex
}

// NewGraph initializes a new instance of a Graph
func NewGraph() *Graph {
	return &Graph{vertices: make(map[int]*Vertex)}
}

// AddVertex adds a new vertex to the graph
func (g *Graph) AddVertex(k int) error {
	if _, ok := g.vertices[k]; ok {
		return fmt.Errorf("vertex %d already exists", k)
	}
	g.vertices[k] = &Vertex{key: k}
	return nil
}

// AddEdge adds an undirected edge between two vertices, handling self-loops correctly
func (g *Graph) AddEdge(from, to int) error {
	fromVertex, fromExists := g.vertices[from]
	toVertex, toExists := g.vertices[to]
	
	if !fromExists || !toExists {
		return fmt.Errorf("failed to add edge: vertex %d or %d not found", from, to)
	}

	if !contains(fromVertex.adjacent, to) {
		fromVertex.adjacent = append(fromVertex.adjacent, toVertex)
	}

	if from != to && !contains(toVertex.adjacent, from) { // Prevents adding the origin vertex into the adjacent slice twice in case of self-loops
		toVertex.adjacent = append(toVertex.adjacent, fromVertex)
	}
	
	return nil
}

// contains checks if a vertex with the given key exists in a slice of vertices
func contains(s []*Vertex, k int) bool {
	for _, v := range s {
		if k == v.key {
			return true
		}
	}
	return false
}

// BreadthFirstSearch traverses the graph using BFS from a start vertex and returns the traversal order
func (g *Graph) BreadthFirstSearch(startKey int) ([]int, error) {
	startVertex, ok := g.vertices[startKey]
	if !ok {
		return nil, fmt.Errorf("start vertex %d not found in the graph", startKey)
	}

	var result []int
	visited := make(map[int]bool)
	queue := list.New()
	queue.PushBack(startVertex)

	for queue.Len() > 0 {
		current := queue.Remove(queue.Front()).(*Vertex)
		if visited[current.key] {
			continue
		}
		
		visited[current.key] = true
		result = append(result, current.key)

		for _, vertex := range current.adjacent {
			if !visited[vertex.key] {
				queue.PushBack(vertex)
			}
		}
	}

	return result, nil
}

func main() {
	graph := NewGraph()
	for i := 0; i < 5; i++ {
		err := graph.AddVertex(i)
		if err != nil {
			fmt.Println("Error adding vertex:", err)
		}
	}
	
	edges := []struct{ from, to int }{
		{0, 1},
		{0, 2},
		{1, 2},
		{2, 3},
		{3, 3}, // Self-loop
	}
	
	for _, edge := range edges {
		if err := graph.AddEdge(edge.from, edge.to); err != nil {
			fmt.Println(err)
		}
	}
	
	fmt.Println("Starting BFS traversal from vertex 2")
	bfsResult, err := graph.BreadthFirstSearch(2)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, v := range bfsResult {
		fmt.Printf("%d ", v)
	}
}

