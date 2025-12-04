package blueprint

import (
	"fmt"
	"strings"
)

const GetVariables = "GetVar"
const SetVariables = "SetVar"
const globalVariablesPrefix = "g_"

type GetVariablesNode struct {
	BaseExecNode
	nodeName string
}

type SetVariablesNode struct {
	BaseExecNode
	nodeName string
}

func (g *GetVariablesNode) GetName() string {
	return g.nodeName
}

func (g *GetVariablesNode) Exec() (int, error) {
	var port IPort
	if strings.HasPrefix(g.GetVariableName(), globalVariablesPrefix) {
		port = g.gr.globalVariables[g.GetVariableName()]
	} else {
		port = g.gr.variables[g.GetVariableName()]
	}

	if port == nil {
		return -1, fmt.Errorf("variable %s not found,node name %s", g.GetVariableName(), g.nodeName)
	}

	if !g.SetOutPort(0, port) {
		return -1, fmt.Errorf("set out port failed,node name %s", g.nodeName)
	}

	return 0, nil
}



func (g *SetVariablesNode) GetName() string {
	return g.nodeName
}

func (g *SetVariablesNode) Exec() (int, error) {
	port := g.GetInPort(1)
	if port == nil {
		return -1, fmt.Errorf("get in port failed,node name %s", g.nodeName)
	}

	varPort := port.Clone()
	if strings.HasPrefix(g.GetVariableName(), globalVariablesPrefix) {
		g.gr.globalVariables[g.GetVariableName()] = varPort
	} else {
		g.gr.variables[g.GetVariableName()] = varPort
	}

	if !g.SetOutPort(1, varPort) {
		return -1, fmt.Errorf("set out port failed,node name %s", g.nodeName)
	}

	return 0, nil
}

