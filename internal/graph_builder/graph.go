package graph_builder

import (
	"fmt"
	"sort"
	"strings"
)

type EdgeType string

const (
	NormalCall  EdgeType = "normal"
	ClosureCall EdgeType = "closure"
	DeferCall   EdgeType = "defer"
	GoCall      EdgeType = "go"
)

type Graph struct {
	Nodes map[string]struct{}
	Edges []Edge
}

type Edge struct {
	From string
	To   string
	Type EdgeType
}

func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string]struct{}),
		Edges: make([]Edge, 0),
	}
}

func (g *Graph) AddNode(name string) {
	g.Nodes[name] = struct{}{}
}

func (g *Graph) AddEdge(from, to string, edgeType EdgeType) {
	g.Edges = append(g.Edges, Edge{From: from, To: to, Type: edgeType})
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

	// Printing edges with different styles
	for _, edge := range g.Edges {
		switch edge.Type {
		case NormalCall:
			sb.WriteString(fmt.Sprintf("  %q -> %q;\n", edge.From, edge.To))
		case ClosureCall:
			sb.WriteString(fmt.Sprintf("  %q -> %q [style=dashed];\n", edge.From, edge.To))
		case DeferCall:
			sb.WriteString(fmt.Sprintf("  %q -> %q [style=dotted];\n", edge.From, edge.To))
		case GoCall:
			sb.WriteString(fmt.Sprintf("  %q -> %q [style=bold];\n", edge.From, edge.To))
		}
	}

	sb.WriteString("}\n")
	return sb.String()
}
