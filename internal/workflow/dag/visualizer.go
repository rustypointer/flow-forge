package dag

import (
	"fmt"
	"strings"
)

type VisualizationOptions struct {
	Direction  string // LR, TB
	ShowRoots  bool
	ShowLeaves bool
}

func DefaultVisualizationOptions() *VisualizationOptions {
	return &VisualizationOptions{
		Direction:  "LR",
		ShowRoots:  true,
		ShowLeaves: true,
	}
}

func (g *Graph) ToDOT(opts *VisualizationOptions) string {
	if opts == nil {
		opts = DefaultVisualizationOptions()
	}

	var b strings.Builder

	b.WriteString("digraph Workflow {\n")

	b.WriteString(
		fmt.Sprintf(
			`rankdir="%s";`,
			opts.Direction,
		),
	)

	b.WriteString("\n")
	b.WriteString(`node [shape=box];`)
	b.WriteString("\n")

	for _, node := range g.Nodes {

		// isolated node
		if len(node.Dependencies) == 0 {
			b.WriteString(
				fmt.Sprintf(
					`"%s";`,
					node.Name,
				),
			)

			b.WriteString("\n")
		}

		// dependencies
		for _, dep := range node.Dependencies {

			b.WriteString(
				fmt.Sprintf(
					`"%s" -> "%s";`,
					dep.Name,
					node.Name,
				),
			)

			b.WriteString("\n")
		}
	}

	if opts.ShowRoots {

		for _, root := range g.Roots() {

			b.WriteString(
				fmt.Sprintf(
					`"%s"[shape=ellipse];`,
					root.Name,
				),
			)

			b.WriteString("\n")
		}
	}

	if opts.ShowLeaves {

		for _, leaf := range g.Leaves() {

			b.WriteString(
				fmt.Sprintf(
					`"%s"[peripheries=2];`,
					leaf.Name,
				),
			)

			b.WriteString("\n")
		}
	}

	b.WriteString("}")

	return b.String()
}
