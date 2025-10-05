package blueprint

import (
	"fmt"
	"github.com/goccy/go-json"
)

type IGraph interface {
	Do(entranceID int64, args ...any) error
	Release()
}

type baseGraph struct {
	entrance map[int64]*execNode // 入口
}

type Graph struct {
	*baseGraph
	graphContext
}

type graphContext struct {
	context         map[string]*ExecContext // 上下文
	variables       map[string]IPort        // 变量
	globalVariables map[string]IPort        // 全局变量
}
type nodeConfig struct {
	Id     string `json:"id"`
	Class  string `json:"class"`
	Module string `json:"module"`
	//Pos         []float64              `json:"pos"`
	PortDefault map[string]interface{} `json:"port_defaultv"`
}

type edgeConfig struct {
	EdgeID       string `json:"edge_id"`
	SourceNodeID string `json:"source_node_id"`
	DesNodeId    string `json:"des_node_id"`

	SourcePortIndex int `json:"source_port_index"`
	DesPortIndex    int `json:"des_port_index"`
}

type MultiTypeValue struct {
	Value interface{}
}

// 实现json.Unmarshaler接口，自定义解码逻辑
func (v *MultiTypeValue) UnmarshalJSON(data []byte) error {
	// 尝试将数据解析为字符串
	var strVal string
	if err := json.Unmarshal(data, &strVal); err == nil {
		v.Value = strVal
		return nil
	}

	// 如果不是字符串，尝试解析为数字
	var intVal int
	if err := json.Unmarshal(data, &intVal); err == nil {
		v.Value = intVal
		return nil
	}

	// 如果不是字符串，尝试解析为数字
	var boolVal bool
	if err := json.Unmarshal(data, &boolVal); err == nil {
		v.Value = boolVal
		return nil
	}

	// 如果不是字符串，尝试解析为数字
	var float64Val float64
	if err := json.Unmarshal(data, &float64Val); err == nil {
		v.Value = float64Val
		return nil
	}

	// 如果都失败，返回错误
	return fmt.Errorf("cannot unmarshal JSON value: %s", string(data))
}

type variablesConfig struct {
	Name  string         `json:"name"`
	Type  string         `json:"type"`
	Value MultiTypeValue `json:"value"`
}

type graphConfig struct {
	GraphName string `json:"graph_name"`
	Time      string `json:"time"`

	Nodes     []nodeConfig      `json:"nodes"`
	Edges     []edgeConfig      `json:"edges"`
	Variables []variablesConfig `json:"variables"`
}

func (gc *graphConfig) GetVariablesByName(varName string) *variablesConfig {
	for _, varCfg := range gc.Variables {
		if varCfg.Name == varName {
			return &varCfg
		}
	}

	return nil
}

func (gc *graphConfig) GetNodeByID(nodeID string) *nodeConfig {
	for _, node := range gc.Nodes {
		if node.Id == nodeID {
			return &node
		}
	}

	return nil
}

func (gr *Graph) Do(entranceID int64, args ...any) error {
	entranceNode := gr.entrance[entranceID]
	if entranceNode == nil {
		return fmt.Errorf("entranceID:%d not found", entranceID)
	}

	gr.variables = map[string]IPort{}
	if gr.globalVariables == nil {
		gr.globalVariables = map[string]IPort{}
	}

	return entranceNode.Do(gr, args...)
}

func (gr *Graph) GetNodeInPortValue(nodeID string, inPortIndex int) IPort {
	if ctx, ok := gr.context[nodeID]; ok {
		if inPortIndex >= len(ctx.InputPorts) || inPortIndex < 0 {
			return nil
		}

		return ctx.InputPorts[inPortIndex]
	}
	return nil
}

func (gr *Graph) GetNodeOutPortValue(nodeID string, outPortIndex int) IPort {
	if ctx, ok := gr.context[nodeID]; ok {
		if outPortIndex >= len(ctx.OutputPorts) || outPortIndex < 0 {
			return nil
		}
		return ctx.OutputPorts[outPortIndex]
	}
	return nil
}

func (gr *Graph) Release() {
	// 有定时器关闭定时器

	// 清理掉所有数据
	*gr = Graph{}
}
