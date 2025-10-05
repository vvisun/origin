package blueprint

import (
	"fmt"
	"github.com/goccy/go-json"
	"os"
	"path/filepath"
	"strings"
)

type GraphPool struct {
	mapGraphs map[string]*baseGraph
	execPool  *ExecPool
}

func (gp *GraphPool) Load(execPool *ExecPool, graphFilePath string) error {
	gp.execPool = execPool
	gp.mapGraphs = make(map[string]*baseGraph, 1024)

	// 检查路径是否存在
	stat, err := os.Stat(graphFilePath)
	if err != nil {
		return fmt.Errorf("failed to access path %s: %v", graphFilePath, err)
	}

	// 如果是单个文件，直接处理
	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", graphFilePath)
	}

	// 遍历目录及其子目录
	return filepath.Walk(graphFilePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("访问路径出错 %s: %v\n", path, err)
			return nil // 继续遍历其他文件
		}

		// 如果是目录，继续遍历
		if info.IsDir() {
			return nil
		}

		// 只处理JSON文件
		if filepath.Ext(path) == ".vgf" {
			return gp.processJSONFile(path)
		}

		return nil
	})
}

func (gp *GraphPool) Create(graphName string) IGraph {
	gr, ok := gp.mapGraphs[graphName]
	if !ok {
		return nil
	}

	var graph Graph
	graph.baseGraph = gr
	graph.context = make(map[string]*ExecContext, 4)

	return &graph
}

func (gp *GraphPool) processJSONFile(filePath string) error {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", filePath, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("关闭文件 %s 时出错: %v\n", filePath, err)
		}
	}()

	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)             // 获取".html"
	name := strings.TrimSuffix(fileName, ext) // 获取"name"
	var gConfig graphConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&gConfig); err != nil {
		return fmt.Errorf("failed to decode JSON from file %s: %v", filePath, err)
	}

	return gp.prepareGraph(name, &gConfig)
}

func (gp *GraphPool) prepareGraph(graphName string, graphConfig *graphConfig) error {
	// 找到所有的入口
	for _, node := range graphConfig.Nodes {
		_, entranceID, ok := getEntranceNodeNameAndID(node.Class)
		if !ok {
			continue
		}

		// 对入口进行预处理
		err := gp.prepareOneEntrance(graphName, entranceID, &node, graphConfig)
		if err != nil {
			return err
		}

	}

	return nil
}

func (gp *GraphPool) genVarExec(nodeCfg *nodeConfig, graphConfig *graphConfig) (IInnerExecNode, string) {
	// 是否为Get_或Set_开头
	if !strings.HasPrefix(nodeCfg.Class, "Get_") && !strings.HasPrefix(nodeCfg.Class, "Set_") {
		return gp.execPool.GetExec(nodeCfg.Class), ""
	}

	// 获取Get_或Set_结尾字符串
	var nodeName string
	var varName string
	if strings.HasPrefix(nodeCfg.Class, "Get_") {
		var typ string
		varName = strings.TrimPrefix(nodeCfg.Class, "Get_")
		varCfg := graphConfig.GetVariablesByName(varName)
		if varCfg != nil {
			typ = varCfg.Type
		}
		nodeName = genGetVariablesNodeName(typ)
	} else if strings.HasPrefix(nodeCfg.Class, "Set_") {
		var typ string
		varName = strings.TrimPrefix(nodeCfg.Class, "Set_")
		varCfg := graphConfig.GetVariablesByName(varName)
		if varCfg != nil {
			typ = varCfg.Type
		}
		nodeName = genSetVariablesNodeName(typ)
	}

	e := gp.execPool.GetExec(nodeName)
	e.(IExecNode).setVariableName(varName)

	return e, varName
}

func (gp *GraphPool) genAllNode(graphConfig *graphConfig) (map[string]*execNode, error) {
	nodes := make(map[string]*execNode)
	for _, node := range graphConfig.Nodes {
		var varName string
		className := node.Class
		if name, _, ok := getEntranceNodeNameAndID(className); ok {
			className = name
		}

		// 获取不到node，则获取变量node
		exec := gp.execPool.GetExec(className)
		if exec == nil {
			exec, varName = gp.genVarExec(&node, graphConfig)
			if exec == nil {
				return nil, fmt.Errorf("%s node has not been registered", node.Class)
			}
		}

		nodes[node.Id] = &execNode{
			Id:                 node.Id,
			execNode:           exec,
			preInPort:          make([]*prePortNode, exec.GetInPortCount()),
			inPortDefaultValue: node.PortDefault,
			variableName:       varName,
		}
	}

	return nodes, nil
}

func (gp *GraphPool) prepareOneNode(mapNodeExec map[string]*execNode, nodeExec *execNode, graphConfig *graphConfig, recursion *int) error {
	*recursion++
	if *recursion > 100 {
		return fmt.Errorf("recursion too deep")
	}

	// 找到所有出口
	var idx int
	for ; nodeExec.execNode.IsOutPortExec(idx) && idx < nodeExec.execNode.GetOutPortCount(); idx++ {
		// 找到出口结点
		nextExecNode := gp.findOutNextNode(graphConfig, mapNodeExec, nodeExec.Id, idx)
		nodeExec.nextNode = append(nodeExec.nextNode, nextExecNode)
		if nextExecNode != nil {
			nextExecNode.beConnect = true
		}
	}

	// 将所有的next填充next
	for _, nextOne := range nodeExec.nextNode {
		if nextOne == nil {
			continue
		}
		// 对出口进行预处理
		err := gp.prepareOneNode(mapNodeExec, nextOne, graphConfig, recursion)
		if err != nil {
			return err
		}
	}

	return nil
}

func (gp *GraphPool) findOutNextNode(graphConfig *graphConfig, mapNodeExec map[string]*execNode, sourceNodeID string, sourcePortIdx int) *execNode {
	// 找到出口的NodeID
	for _, edge := range graphConfig.Edges {
		if edge.SourceNodeID == sourceNodeID && edge.SourcePortIndex == sourcePortIdx {
			return mapNodeExec[edge.DesNodeId]
		}
	}

	return nil
}

// prepareOneEntrance 先处理执行Exec入出口连线
func (gp *GraphPool) prepareOneEntrance(graphName string, entranceID int64, nodeCfg *nodeConfig, graphConfig *graphConfig) error {
	// 将所有的Node执行结点生成出来
	mapNodes, err := gp.genAllNode(graphConfig)
	if err != nil {
		return err
	}

	// 从入口结点开始做预处理，将next结点都统一生成
	nodeExec := mapNodes[nodeCfg.Id]
	if nodeExec == nil {
		return fmt.Errorf("entrance node %s not found", nodeCfg.Id)
	}

	nodeExec.isEntrance = true
	err = gp.prepareOneNode(mapNodes, nodeExec, graphConfig, new(int))
	if err != nil {
		return err
	}

	// 处理inPort前置结点
	err = gp.prepareInPort(mapNodes, graphConfig)
	if err != nil {
		return err
	}

	var gr baseGraph
	gr.entrance = make(map[int64]*execNode, 16)
	gr.entrance[entranceID] = nodeExec

	if _, ok := gp.mapGraphs[graphName]; ok {
		return fmt.Errorf("baseGraph %s already exists", graphName)
	}

	gp.mapGraphs[graphName] = &gr

	return nil
}

func (gp *GraphPool) findPreInPortNode(mapNodes map[string]*execNode, nodeExec *execNode, graphConfig *graphConfig, portIdx int) *prePortNode {
	for _, edge := range graphConfig.Edges {
		if edge.DesNodeId == nodeExec.Id && edge.DesPortIndex == portIdx {
			srcNode := mapNodes[edge.SourceNodeID]
			if srcNode == nil {
				return nil
			}

			var preNode prePortNode
			preNode.node = srcNode
			preNode.outPortIndex = edge.SourcePortIndex

			return &preNode
		}
	}

	return nil
}

func (gp *GraphPool) preparePreInPortNode(mapNodes map[string]*execNode, nodeExec *execNode, graphConfig *graphConfig) error {
	// 找到当前结点的所有inPort的前一个端口
	for i := 0; i < nodeExec.execNode.GetInPortCount(); i++ {
		// 如果是执行结点，则跳过
		if nodeExec.execNode.IsInPortExec(i) {
			continue
		}

		// 找到入口的上一个结点
		preNode := gp.findPreInPortNode(mapNodes, nodeExec, graphConfig, i)
		if preNode == nil {
			continue
		}
		nodeExec.preInPort[i] = preNode
	}
	return nil
}

func (gp *GraphPool) prepareInPort(mapNodeExec map[string]*execNode, graphConfig *graphConfig) error {
	for _, e := range mapNodeExec {
		// 对当前结点的入口进行预处理
		err := gp.preparePreInPortNode(mapNodeExec, e, graphConfig)
		if err != nil {
			return err
		}
	}

	return nil
}
