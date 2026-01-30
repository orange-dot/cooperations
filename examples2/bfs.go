package main

import (
	"errors"
	"fmt"
	"sort"
)

type Graph struct {
	adj    map[string][]string
	sorted bool
}

func NewGraph() *Graph {
	return &Graph{adj: make(map[string][]string)}
}

func (g *Graph) AddNode(id string) {
	if id == "" {
		return
	}
	if _, ok := g.adj[id]; !ok {
		g.adj[id] = nil
	}
}

func (g *Graph) AddEdge(from, to string) {
	if from == "" || to == "" {
		return
	}
	g.AddNode(from)
	g.AddNode(to)
	g.adj[from] = append(g.adj[from], to)
	g.sorted = false
}

func (g *Graph) HasNode(id string) bool {
	_, ok := g.adj[id]
	return ok
}

func (g *Graph) sortAdjacency() {
	if g.sorted {
		return
	}
	for k := range g.adj {
		sort.Strings(g.adj[k])
	}
	g.sorted = true
}

// BFS traverses the graph from start, visiting each reachable node at most once.
// It returns nodes in the order they are visited.
// If visit is non-nil, it is called for each visited node in order.
// Neighbors are visited in deterministic order (lexicographically sorted).
func (g *Graph) BFS(start string, visit func(node string) error) ([]string, error) {
	if start == "" {
		return nil, errors.New("start node is required")
	}
	if !g.HasNode(start) {
		return nil, fmt.Errorf("start node %q not found", start)
	}

	g.sortAdjacency()

	visited := make(map[string]bool, len(g.adj))
	order := make([]string, 0, len(g.adj))

	queue := make([]string, 0, 16)
	head := 0

	visited[start] = true
	queue = append(queue, start)

	for head < len(queue) {
		u := queue[head]
		head++

		order = append(order, u)
		if visit != nil {
			if err := visit(u); err != nil {
				return order, err
			}
		}

		for _, v := range g.adj[u] {
			if v == "" {
				continue
			}
			if !visited[v] {
				visited[v] = true
				queue = append(queue, v)
			}
		}
	}

	return order, nil
}

func demoBFS() {
	g := NewGraph()
	g.AddEdge("A", "B")
	g.AddEdge("A", "C")
	g.AddEdge("B", "D")
	g.AddEdge("C", "E")
	g.AddEdge("E", "B") // cycle

	order, err := g.BFS("A", func(node string) error {
		fmt.Println("visit:", node)
		return nil
	})
	if err != nil {
		fmt.Println("BFS error:", err)
		return
	}

	fmt.Println("order:", order)
}
