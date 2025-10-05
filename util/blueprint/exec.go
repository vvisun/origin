package blueprint

import "fmt"

type IBaseExecNode interface {
	initInnerExecNode(innerNode *innerExecNode)
	initExecNode(gr *Graph, en *execNode) error
	GetPorts() ([]IPort, []IPort)
	getExecNodeInfo() (*ExecContext, *execNode)
	setExecNodeInfo(gr *ExecContext, en *execNode)
}

type IInnerExecNode interface {
	GetName() string
	SetExec(exec IExecNode)
	IsInPortExec(index int) bool
	IsOutPortExec(index int) bool
	GetInPortCount() int
	GetOutPortCount() int
	CloneInOutPort() ([]IPort, []IPort)

	GetInPort(index int) IPort
	GetOutPort(index int) IPort

	GetInPortParamStartIndex() int
	GetOutPortParamStartIndex() int
}

type IExecNode interface {
	GetName() string
	DoNext(index int) error
	Exec() (int, error) // 返回后续执行的Node的Index
	GetNextExecLen() int
	getInnerExecNode() IInnerExecNode

	setVariableName(name string) bool
}

type innerExecNode struct {
	Name        string
	Title       string
	Package     string
	Description string

	inPort  []IPort
	outPort []IPort

	inPortParamStartIndex  int // 输入参数的起始索引,用于排除执行入口
	outPortParamStartIndex int // 输出参数的起始索引,用于排除执行出口

	IExecNode
}

type BaseExecNode struct {
	*innerExecNode

	// 执行时初始化的数据
	*ExecContext
	gr       *Graph
	execNode *execNode
}

type InputConfig struct {
	Name      string `json:"name"`
	PortType  string `json:"type"`
	DataType  string `json:"data_type"`
	HasInput  bool   `json:"has_input"`
	PinWidget string `json:"pin_widget"`
}

type OutInputConfig struct {
	Name     string `json:"name"`
	PortType string `json:"type"`
	DataType string `json:"data_type"`
	HasInput bool   `json:"has_input"`
}

type BaseExecConfig struct {
	Name        string           `json:"name"`
	Title       string           `json:"title"`
	Package     string           `json:"package"`
	Description string           `json:"description"`
	IsPure      bool             `json:"is_pure"`
	Inputs      []InputConfig    `json:"inputs"`
	Outputs     []OutInputConfig `json:"outputs"`
}

func (em *innerExecNode) AppendInPort(port ...IPort) {
	if len(em.inPort) == 0 {
		em.inPortParamStartIndex = -1
	}

	for i := 0; i < len(port); i++ {
		if !port[i].IsPortExec() && em.inPortParamStartIndex < 0 {
			em.inPortParamStartIndex = len(em.inPort)
		}

		em.inPort = append(em.inPort, port[i])
	}
}

func (em *innerExecNode) AppendOutPort(port ...IPort) {
	if len(em.outPort) == 0 {
		em.outPortParamStartIndex = -1
	}
	for i := 0; i < len(port); i++ {
		if !port[i].IsPortExec() && em.outPortParamStartIndex < 0 {
			em.outPortParamStartIndex = len(em.outPort)
		}
		em.outPort = append(em.outPort, port[i])
	}
}

func (em *innerExecNode) GetName() string {
	return em.Name
}

func (em *innerExecNode) SetExec(exec IExecNode) {
	em.IExecNode = exec
}

func (em *innerExecNode) CloneInOutPort() ([]IPort, []IPort) {
	inPorts := make([]IPort, 0, 2)
	for _, port := range em.inPort {
		if port.IsPortExec() {
			// 执行入口, 不需要克隆,占位处理
			inPorts = append(inPorts, nil)
			continue
		}

		inPorts = append(inPorts, port.Clone())
	}
	outPorts := make([]IPort, 0, 2)

	for _, port := range em.outPort {
		if port.IsPortExec() {
			outPorts = append(outPorts, nil)
			continue
		}
		outPorts = append(outPorts, port.Clone())
	}

	return inPorts, outPorts
}

func (em *innerExecNode) IsInPortExec(index int) bool {
	if index >= len(em.inPort) || index < 0 {
		return false
	}

	return em.inPort[index].IsPortExec()
}

func (em *innerExecNode) IsOutPortExec(index int) bool {
	if index >= len(em.outPort) || index < 0 {
		return false
	}

	return em.outPort[index].IsPortExec()
}

func (em *innerExecNode) GetInPortCount() int {
	return len(em.inPort)
}

func (em *innerExecNode) GetOutPortCount() int {
	return len(em.outPort)
}

func (em *innerExecNode) GetInPort(index int) IPort {
	if index >= len(em.inPort) || index < 0 {
		return nil
	}
	return em.inPort[index]
}

func (em *innerExecNode) GetOutPort(index int) IPort {
	if index >= len(em.outPort) || index < 0 {
		return nil
	}
	return em.outPort[index]
}

func (em *innerExecNode) GetInPortParamStartIndex() int {
	return em.inPortParamStartIndex
}

func (em *innerExecNode) GetOutPortParamStartIndex() int {
	return em.outPortParamStartIndex
}

func (en *BaseExecNode) initInnerExecNode(innerNode *innerExecNode) {
	en.innerExecNode = innerNode
}

func (en *BaseExecNode) getExecNodeInfo() (*ExecContext, *execNode) {
	return en.ExecContext, en.execNode
}

func (en *BaseExecNode) setExecNodeInfo(c *ExecContext, e *execNode) {
	en.ExecContext = c
	en.execNode = e
}

func (en *BaseExecNode) initExecNode(gr *Graph, node *execNode) error {
	ctx, ok := gr.context[node.Id]
	if !ok {
		return fmt.Errorf("node %s not found", node.Id)
	}

	en.ExecContext = ctx
	en.gr = gr
	en.execNode = node
	return nil
}

func (en *BaseExecNode) GetPorts() ([]IPort, []IPort) {
	return en.InputPorts, en.OutputPorts
}

func (en *BaseExecNode) GetInPort(index int) IPort {
	if en.InputPorts == nil {
		return nil
	}

	if index >= len(en.InputPorts) || index < 0 {
		return nil
	}
	return en.InputPorts[index]
}

func (en *BaseExecNode) GetOutPort(index int) IPort {
	if en.OutputPorts == nil {
		return nil
	}
	if index >= len(en.OutputPorts) || index < 0 {
		return nil
	}
	return en.OutputPorts[index]
}

func (en *BaseExecNode) SetOutPort(index int, val IPort) bool {
	if index >= len(en.OutputPorts) || index < 0 {
		return false
	}
	en.OutputPorts[index].SetValue(val)
	return true
}

func (en *BaseExecNode) GetInPortInt(index int) (Port_Int, bool) {
	port := en.GetInPort(index)
	if port == nil {
		return 0, false
	}

	return port.GetInt()
}

func (en *BaseExecNode) GetInPortFloat(index int) (Port_Float, bool) {
	port := en.GetInPort(index)
	if port == nil {
		return 0, false
	}
	return port.GetFloat()
}

func (en *BaseExecNode) GetInPortStr(index int) (Port_Str, bool) {
	port := en.GetInPort(index)
	if port == nil {
		return "", false
	}
	return port.GetStr()
}

func (en *BaseExecNode) GetInPortArrayValInt(index int, idx int) (Port_Int, bool) {
	port := en.GetInPort(index)
	if port == nil {
		return 0, false
	}
	return port.GetArrayValInt(idx)
}

func (en *BaseExecNode) GetInPortArrayValStr(idx int) (Port_Str, bool) {
	port := en.GetInPort(idx)
	if port == nil {
		return "", false
	}
	return port.GetArrayValStr(idx)
}

func (en *BaseExecNode) GetInPortBool(index int) (Port_Bool, bool) {
	port := en.GetInPort(index)
	if port == nil {
		return false, false
	}
	return port.GetBool()
}

func (en *BaseExecNode) GetOutPortInt(index int) (Port_Int, bool) {
	port := en.GetOutPort(index)
	if port == nil {
		return 0, false
	}
	return port.GetInt()
}

func (en *BaseExecNode) GetOutPortFloat(index int) (Port_Float, bool) {
	port := en.GetOutPort(index)
	if port == nil {
		return 0, false
	}
	return port.GetFloat()
}

func (en *BaseExecNode) GetOutPortStr(index int) (Port_Str, bool) {
	port := en.GetOutPort(index)
	if port == nil {
		return "", false
	}
	return port.GetStr()
}

func (en *BaseExecNode) GetOutPortArrayValInt(index int, idx int) (Port_Int, bool) {
	port := en.GetOutPort(index)
	if port == nil {
		return 0, false
	}
	return port.GetArrayValInt(idx)
}

func (en *BaseExecNode) GetOutPortArrayValStr(index int, idx int) (Port_Str, bool) {
	port := en.GetOutPort(index)
	if port == nil {
		return "", false
	}
	return port.GetArrayValStr(idx)
}

func (en *BaseExecNode) GetOutPortBool(index int) (Port_Bool, bool) {
	port := en.GetInPort(index)
	if port == nil {
		return false, false
	}
	return port.GetBool()
}

func (en *BaseExecNode) SetInPortInt(index int, val Port_Int) bool {
	port := en.GetInPort(index)
	if port == nil {
		return false
	}
	return port.SetInt(val)
}

func (en *BaseExecNode) SetInPortFloat(index int, val Port_Float) bool {
	port := en.GetInPort(index)
	if port == nil {
		return false
	}
	return port.SetFloat(val)
}

func (en *BaseExecNode) SetInPortStr(index int, val Port_Str) bool {
	port := en.GetInPort(index)
	if port == nil {
		return false
	}
	return port.SetStr(val)
}

func (en *BaseExecNode) SetInBool(index int, val Port_Bool) bool {
	port := en.GetInPort(index)
	if port == nil {
		return false
	}
	return port.SetBool(val)
}

func (en *BaseExecNode) SetInPortArrayValInt(index int, idx int, val Port_Int) bool {
	port := en.GetInPort(index)
	if port == nil {
		return false
	}
	return port.SetArrayValInt(idx, val)
}

func (en *BaseExecNode) SetInPortArrayValStr(index int, idx int, val Port_Str) bool {
	port := en.GetInPort(index)
	if port == nil {
		return false
	}
	return port.SetArrayValStr(idx, val)
}

func (en *BaseExecNode) AppendInPortArrayValInt(index int, val Port_Int) bool {
	port := en.GetInPort(index)
	if port == nil {
		return false
	}
	return port.AppendArrayValInt(val)
}

func (en *BaseExecNode) AppendInPortArrayValStr(index int, val Port_Str) bool {
	port := en.GetInPort(index)
	if port == nil {
		return false
	}
	return port.AppendArrayValStr(val)
}

func (en *BaseExecNode) GetInPortArrayLen(index int) Port_Int {
	port := en.GetInPort(index)
	if port == nil {
		return 0
	}
	return port.GetArrayLen()
}

func (en *BaseExecNode) SetOutPortInt(index int, val Port_Int) bool {
	port := en.GetOutPort(index)
	if port == nil {
		return false
	}
	return port.SetInt(val)
}

func (en *BaseExecNode) SetOutPortFloat(index int, val Port_Float) bool {
	port := en.GetOutPort(index)
	if port == nil {
		return false
	}
	return port.SetFloat(val)
}

func (en *BaseExecNode) SetOutPortStr(index int, val Port_Str) bool {
	port := en.GetOutPort(index)
	if port == nil {
		return false
	}
	return port.SetStr(val)
}

func (en *BaseExecNode) SetOutPortBool(index int, val Port_Bool) bool {
	port := en.GetOutPort(index)
	if port == nil {
		return false
	}
	return port.SetBool(val)
}

func (en *BaseExecNode) SetOutPortArrayValInt(index int, idx int, val Port_Int) bool {
	port := en.GetOutPort(index)
	if port == nil {
		return false
	}
	return port.SetArrayValInt(idx, val)
}

func (en *BaseExecNode) SetOutPortArrayValStr(index int, idx int, val Port_Str) bool {
	port := en.GetOutPort(index)
	if port == nil {
		return false
	}
	return port.SetArrayValStr(idx, val)
}

func (en *BaseExecNode) AppendOutPortArrayValInt(index int, val Port_Int) bool {
	port := en.GetOutPort(index)
	if port == nil {
		return false
	}
	return port.AppendArrayValInt(val)
}

func (en *BaseExecNode) AppendOutPortArrayValStr(index int, val Port_Str) bool {
	port := en.GetOutPort(index)
	if port == nil {
		return false
	}
	return port.AppendArrayValStr(val)
}

func (en *BaseExecNode) GetOutPortArrayLen(index int) Port_Int {
	port := en.GetOutPort(index)
	if port == nil {
		return 0
	}
	return port.GetArrayLen()
}

func (en *BaseExecNode) DoNext(index int) error {
	// -1 表示中断运行
	if index == -1 {
		return nil
	}

	if index < 0 || index >= len(en.execNode.nextNode) {
		return fmt.Errorf("next index %d not found", index)
	}

	if en.execNode.nextNode[index] == nil {
		return nil
	}

	return en.execNode.nextNode[index].Do(en.gr)
}

func (en *BaseExecNode) GetNextExecLen() int {
	return len(en.execNode.nextNode)
}

func (en *BaseExecNode) getInnerExecNode() IInnerExecNode {
	return en.innerExecNode.IExecNode.(IInnerExecNode)
}

func (en *BaseExecNode) setVariableName(name string) bool {
	return false
}
