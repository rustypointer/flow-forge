package dag

import "os"

func ExportDOT(
	graph *Graph,
	path string,
	opts *VisualizationOptions,
) error {

	data := graph.ToDOT(opts)

	return os.WriteFile(
		path,
		[]byte(data),
		0644,
	)
}
