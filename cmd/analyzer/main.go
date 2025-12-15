package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"strings"

	gb "github.com/coIorbIind/go_code_analyzer/internal/graph_builder"
)

func main() {
	include := flag.String("include", "", "Comma-separated list of packages to include")
	exclude := flag.String("exclude", "", "Comma-separated list of packages to exclude")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Usage: callgraph [flags] <project-dir>")
		fmt.Println("Flags:")
		fmt.Println("  -include: Comma-separated list of packages to include")
		fmt.Println("  -exclude: Comma-separated list of packages to exclude")
		os.Exit(1)
	}

	projectDir, err := filepath.Abs(args[0])
	if err != nil {
		fmt.Printf("Invalid path: %v\n", err)
		os.Exit(1)
	}

	cfg := gb.Config{
		IncludePkgs: make(map[string]bool),
		ExcludePkgs: make(map[string]bool),
	}

	if *include != "" {
		for _, pkg := range strings.Split(*include, ",") {
			cfg.IncludePkgs[pkg] = true
		}
	}

	if *exclude != "" {
		for _, pkg := range strings.Split(*exclude, ",") {
			cfg.ExcludePkgs[pkg] = true
		}
	}

	graph := gb.NewGraph()

	if err := gb.AnalyzeProject(projectDir, graph, &cfg); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(graph.ToDOT())
}
