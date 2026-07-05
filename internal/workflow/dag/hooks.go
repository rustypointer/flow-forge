package dag

type ExecutionHooks struct {
	OnStart  func(*Node)
	OnFinish func(result *TaskResult)
}
