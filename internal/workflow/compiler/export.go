package compiler

import "workflow_engine/internal/workflow/dag"

func (c *CompiledWorkflow) ExportVisualization(
	path string,
	opts *dag.VisualizationOptions,
) error {

	return dag.ExportDOT(
		c.Graph,
		path,
		opts,
	)
}
