package blueprint

type ExecContext struct {
	InputPorts  []IPort
	OutputPorts []IPort
}

func (ec *ExecContext) Reset() {
	*ec = ExecContext{}
}
