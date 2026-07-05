package benchmark

import (
	"testing"

	"workflow_engine/internal/workflow/compiler"
)

func BenchmarkCompile100(b *testing.B) {
	wf := buildLinearWorkflow(100)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = compiler.Compile(wf)
	}
}

func BenchmarkCompile1000(b *testing.B) {
	wf := buildLinearWorkflow(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = compiler.Compile(wf)
	}
}
