package graph_builder

import (
	"fmt"
	"sort"
	"strings"
)

type Graph struct {
	Nodes map[string]struct{}
	Edges map[string][]string
}

func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string]struct{}),
		Edges: make(map[string][]string),
	}
}

func (g *Graph) AddNode(name string) {
	g.Nodes[name] = struct{}{}
}

func (g *Graph) AddEdge(from, to string) {
	g.Edges[from] = append(g.Edges[from], to)
}

func (g *Graph) ToDOT() string {
	var sb strings.Builder
	sb.WriteString("digraph G {\n")
	sb.WriteString("  node [shape=box];\n")
	sb.WriteString("  rankdir=LR;\n\n")

	// Sorting nodes
	nodes := make([]string, 0, len(g.Nodes))
	for node := range g.Nodes {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)

	// Printing nodes
	for _, node := range nodes {
		sb.WriteString(fmt.Sprintf("  %q;\n", node))
	}

	sb.WriteString("\n")

	// Sorting Edges
	fromNodes := make([]string, 0, len(g.Edges))
	for from := range g.Edges {
		fromNodes = append(fromNodes, from)
	}
	sort.Strings(fromNodes)

	// Printing edges
	for _, from := range fromNodes {
		toList := g.Edges[from]
		sort.Strings(toList)

		for _, to := range toList {
			sb.WriteString(fmt.Sprintf("  %q -> %q;\n", from, to))
		}
	}

	sb.WriteString("}\n")
	return sb.String()
}
