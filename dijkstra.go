package main

import (
	"fmt"
	"math"
)

// Define a struct for edges which includes destination and cost
type Edge struct {
	dest int
	cost int64
}

// Graph holds the vertices and adjacency list
type Graph struct {
	vertices int
	adj      map[int][]Edge
}

// NewGraph initializes a new Graph with a given number of vertices
func NewGraph(vertices int) *Graph {
	return &Graph{
		vertices: vertices,
		adj:      make(map[int][]Edge),
	}
}

// AddEdge adds an edge to the graph (both directions for undirected graph)
func (g *Graph) AddEdge(source, destination int, cost int64) error {
	if source < 0 || source >= g.vertices {
		return fmt.Errorf("invalid source vertex: %d", source)
	}
	if destination < 0 || destination >= g.vertices {
		return fmt.Errorf("invalid destination vertex: %d", destination)
	}
	if cost < 0 {
		return fmt.Errorf("negative edge cost: %d", cost)
	}
	g.adj[source] = append(g.adj[source], Edge{dest: destination, cost: cost})
	g.adj[destination] = append(g.adj[destination], Edge{dest: source, cost: cost}) // Add reverse edge for undirected graph
	return nil
}

// Dijkstra computes the shortest path from a given source vertex to all other vertices
func (g Graph) Dijkstra(start int) ([]int64, error) {
	// check for valid start
	if start < 0 || start >= g.vertices {
		return nil, fmt.Errorf("invalid start vertex: %d", start)
	}

	distance := make([]int64, g.vertices)
	visited := make([]bool, g.vertices)

	for i := range distance {
		distance[i] = math.MaxInt64
	}

	distance[start] = 0

	for i := 0; i < g.vertices-1; i++ {
		u := minDistance(distance, visited)
		if u == -1 {
			break
		}
		visited[u] = true

		for _, edge := range g.adj[u] {
			if !visited[edge.dest] && distance[u] != math.MaxInt64 && distance[u]+edge.cost < distance[edge.dest] {
				distance[edge.dest] = distance[u] + edge.cost
			}
		}
	}
	return distance, nil
}

// minDistance finds the vertex with the minimum distance value, from
// the set of vertices not yet included in the shortest path tree
func minDistance(dist []int64, visited []bool) int {
	min := int64(math.MaxInt64)
	minIndex := -1

	for i, v := range dist {
		if !visited[i] && v <= min {
			min = v
			minIndex = i
		}
	}
	return minIndex
}

func main() {
	vertices := 9
	g := NewGraph(vertices)
	if err := g.AddEdge(0, 1, 4); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	if err := g.AddEdge(0, 7, 8); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	if err := g.AddEdge(1, 2, 8); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	if err := g.AddEdge(1, 7, 11); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	if err := g.AddEdge(2, 3, 7); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	if err := g.AddEdge(2, 8, 2); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	if err := g.AddEdge(2, 5, 4); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	if err := g.AddEdge(3, 4, 9); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	if err := g.AddEdge(3, 5, 14); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	if err := g.AddEdge(4, 5, 10); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	if err := g.AddEdge(5, 6, 2); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	if err := g.AddEdge(6, 7, 1); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	if err := g.AddEdge(6, 8, 6); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	if err := g.AddEdge(7, 8, 7); err != nil {
		fmt.Println("Error: ", err)
		return
	}

	distance, err := g.Dijkstra(0)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	fmt.Println("Vertex   Distance from Source")
	for i := 0; i < vertices; i++ {
		fmt.Println(i, "\t\t", distance[i])
	}
}

// NEXT: reviewer
