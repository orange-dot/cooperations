package main

import (
	"fmt"
	"math"
)

// Define a struct for edges which includes destination and cost
type Edge struct {
	dest, cost int
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
func (g *Graph) AddEdge(source, destination, cost int) {
	g.adj[source] = append(g.adj[source], Edge{dest: destination, cost: cost})
	g.adj[destination] = append(g.adj[destination], Edge{dest: source, cost: cost}) // Add reverse edge for undirected graph
}

// Dijkstra computes the shortest path from a given source vertex to all other vertices
func (g Graph) Dijkstra(start int) ([]int, error) {
	// check for valid start
	if start < 0 || start >= g.vertices {
		return nil, fmt.Errorf("invalid start vertex: %d", start)
	}

	distance := make([]int, g.vertices)
	visited := make([]bool, g.vertices)

	for i := range distance {
		distance[i] = math.MaxInt32
	}

	distance[start] = 0

	for i := 0; i < g.vertices-1; i++ {
		u := minDistance(distance, visited)
		visited[u] = true

		for _, edge := range g.adj[u] {
			if !visited[edge.dest] && distance[u] != math.MaxInt32 && distance[u]+edge.cost < distance[edge.dest] {
				distance[edge.dest] = distance[u] + edge.cost
			}
		}
	}
	return distance, nil
}

// minDistance finds the vertex with the minimum distance value, from
// the set of vertices not yet included in the shortest path tree
func minDistance(dist []int, visited []bool) int {
	min := math.MaxInt32
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
	g.AddEdge(0, 1, 4)
	g.AddEdge(0, 7, 8)
	g.AddEdge(1, 2, 8)
	g.AddEdge(1, 7, 11)
	g.AddEdge(2, 3, 7)
	g.AddEdge(2, 8, 2)
	g.AddEdge(2, 5, 4)
	g.AddEdge(3, 4, 9)
	g.AddEdge(3, 5, 14)
	g.AddEdge(4, 5, 10)
	g.AddEdge(5, 6, 2)
	g.AddEdge(6, 7, 1)
	g.AddEdge(6, 8, 6)
	g.AddEdge(7, 8, 7)

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