package dag

import "sort"

type Graph struct {
	Nodes map[string]*Node
}

func NewGraph(nodesList []*Node) *Graph {
	adjMap := make(map[string]*Node)

	for _, node := range nodesList {
		adjMap[node.Name] = node
	}

	return &Graph{
		Nodes: adjMap,
	}
}

func (g *Graph) InDegrees() map[string]int {
	inDegrees := make(map[string]int)

	for stepName, node := range g.Nodes {
		inDegrees[stepName] = len(node.Dependencies)
	}

	return inDegrees
}

func (g *Graph) Roots() []*Node {
	roots := make([]*Node, 0)

	for _, node := range g.Nodes {
		if len(node.Dependencies) == 0 {
			roots = append(roots, node)
		}
	}

	sort.Slice(roots, func(i, j int) bool {
		return roots[i].Name < roots[j].Name
	})

	return roots
}

func (g *Graph) Leaves() []*Node {
	leaves := make([]*Node, 0)

	for _, node := range g.Nodes {
		if len(node.Dependents) == 0 {
			leaves = append(leaves, node)
		}
	}

	sort.Slice(leaves, func(i, j int) bool {
		return leaves[i].Name < leaves[j].Name
	})

	return leaves
}
