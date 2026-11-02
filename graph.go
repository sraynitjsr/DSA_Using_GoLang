package main

import (
	"container/heap"
	"fmt"
	"math"
)

type Graph struct {
	adj      map[int][]int
	directed bool
}

func NewGraph(directed bool) *Graph {
	return &Graph{adj: make(map[int][]int), directed: directed}
}

func (g *Graph) AddEdge(u, v int) {
	g.adj[u] = append(g.adj[u], v)
	if g.directed {
		if _, ok := g.adj[v]; !ok {
			g.adj[v] = nil
		}
		return
	}
	g.adj[v] = append(g.adj[v], u)
}

func (g *Graph) BFS(start int) []int {
	visited := map[int]bool{start: true}
	queue := []int{start}
	var order []int
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		order = append(order, u)
		for _, v := range g.adj[u] {
			if !visited[v] {
				visited[v] = true
				queue = append(queue, v)
			}
		}
	}
	return order
}

func (g *Graph) DFS(start int) []int {
	visited := make(map[int]bool)
	var order []int
	var visit func(u int)
	visit = func(u int) {
		visited[u] = true
		order = append(order, u)
		for _, v := range g.adj[u] {
			if !visited[v] {
				visit(v)
			}
		}
	}
	visit(start)
	return order
}

type Edge struct{ to, weight int }

type WeightedGraph struct {
	adj map[int][]Edge
}

func NewWeightedGraph() *WeightedGraph {
	return &WeightedGraph{adj: make(map[int][]Edge)}
}

func (g *WeightedGraph) AddEdge(u, v, w int) {
	g.adj[u] = append(g.adj[u], Edge{v, w})
	g.adj[v] = append(g.adj[v], Edge{u, w})
}

type item struct{ node, dist int }
type minHeap []item

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i].dist < h[j].dist }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x any)        { *h = append(*h, x.(item)) }
func (h *minHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func (g *WeightedGraph) Dijkstra(src int) map[int]int {
	dist := make(map[int]int)
	for u := range g.adj {
		dist[u] = math.MaxInt
	}
	dist[src] = 0
	h := &minHeap{{src, 0}}
	for h.Len() > 0 {
		cur := heap.Pop(h).(item)
		if cur.dist > dist[cur.node] {
			continue
		}
		for _, e := range g.adj[cur.node] {
			if nd := cur.dist + e.weight; nd < dist[e.to] {
				dist[e.to] = nd
				heap.Push(h, item{e.to, nd})
			}
		}
	}
	return dist
}

func main() {
	g := NewGraph(false)
	g.AddEdge(1, 2)
	g.AddEdge(1, 3)
	g.AddEdge(2, 4)
	g.AddEdge(3, 4)
	g.AddEdge(4, 5)
	fmt.Println("BFS:", g.BFS(1))
	fmt.Println("DFS:", g.DFS(1))

	wg := NewWeightedGraph()
	wg.AddEdge(1, 2, 7)
	wg.AddEdge(1, 3, 9)
	wg.AddEdge(2, 4, 15)
	wg.AddEdge(3, 4, 2)
	fmt.Println("Dijkstra from 1:", wg.Dijkstra(1))
}
