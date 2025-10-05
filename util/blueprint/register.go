package blueprint

var execNodes []IExecNode

func RegExecNode(exec IExecNode) {
	execNodes = append(execNodes, exec)
}
