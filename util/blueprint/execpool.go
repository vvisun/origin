package blueprint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Entrance 格式:Entrance_XXXX_ID
const (
	Entrance = "Entrance_"
)

type ExecPool struct {
	innerExecNodeMap map[string]IInnerExecNode // 所有配置对应的结点信息
	execNodeMap      map[string]IExecNode      // 实际注册的执行结点
}

func (em *ExecPool) Load(execDefFilePath string) error {
	em.innerExecNodeMap = make(map[string]IInnerExecNode, 512)
	em.execNodeMap = make(map[string]IExecNode, 512)

	// 检查路径是否存在
	stat, err := os.Stat(execDefFilePath)
	if err != nil {
		return fmt.Errorf("failed to access path %s: %v", execDefFilePath, err)
	}

	// 如果是单个文件，直接处理
	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", execDefFilePath)
	}

	// 遍历目录及其子目录
	err = filepath.Walk(execDefFilePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("访问路径出错 %s: %v\n", path, err)
			return nil // 继续遍历其他文件
		}

		// 如果是目录，继续遍历
		if info.IsDir() {
			return nil
		}

		// 只处理JSON文件
		if filepath.Ext(path) == ".json" {
			// 将配置的结点初始化为innerExecNode将加入到innerExecNodeMap中
			return em.processJSONFile(path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk path %s: %v", execDefFilePath, err)
	}

	// 生成变量配置对应的配置结点GetVar_类型、SetVar_类型
	return em.regVariablesNode()
}

// 处理单个JSON文件
func (em *ExecPool) processJSONFile(filePath string) error {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", filePath, err)
	}

	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			fmt.Printf("failed to close file %s: %v\n", filePath, err)
			return
		}
	}(file)

	var baseExecConfig []BaseExecConfig
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&baseExecConfig); err != nil {
		return fmt.Errorf("failed to decode JSON from file %s: %v", filePath, err)
	}

	for i := range baseExecConfig {
		// 对PortId进行排序
		sort.Slice(baseExecConfig[i].Inputs, func(left, right int) bool {
			return baseExecConfig[i].Inputs[left].PortId < baseExecConfig[i].Inputs[right].PortId
		})
		// 对PortId进行排序
		sort.Slice(baseExecConfig[i].Outputs, func(left, right int) bool {
			return baseExecConfig[i].Outputs[left].PortId < baseExecConfig[i].Outputs[right].PortId
		})

		// 根据配置的结点信息，创建innerExecNode
		var execError error
		exec, execError := em.createExecFromJSON(baseExecConfig[i])
		if execError != nil {
			return execError
		}

		// 加载到innerExecNodeMap中
		if !em.addInnerExec(exec) {
			return fmt.Errorf("exec %s already registered", exec.GetName())
		}
	}

	return nil
}

func (em *ExecPool) createPortByDataType(nodeName, portName, dataType string) (IPort, error) {
	switch strings.ToLower(dataType) {
	case Config_DataType_Int, Config_DataType_Integer:
		return NewPortInt(), nil
	case Config_DataType_Float:
		return NewPortFloat(), nil
	case Config_DataType_Str:
		return NewPortStr(), nil
	case Config_DataType_Boolean, Config_DataType_Bool:
		return NewPortBool(), nil
	case Config_DataType_Array:
		return NewPortArray(), nil
	}

	return nil, fmt.Errorf("invalid data type %s,node %s port %s", dataType, nodeName, portName)
}

func (em *ExecPool) createExecFromJSON(baseExecConfig BaseExecConfig) (IInnerExecNode, error) {
	var baseExec innerExecNode

	// 如果是入口名，则按入口名Entrance_ArrayParam_000002生成结点名:Entrance_ArrayParam
	entranceName, _, ok := getEntranceNodeNameAndID(baseExecConfig.Name)
	if ok {
		baseExec.Name = entranceName
	} else {
		baseExec.Name = baseExecConfig.Name
	}
	baseExec.Title = baseExecConfig.Title
	baseExec.Package = baseExecConfig.Package
	baseExec.Description = baseExecConfig.Description
	baseExec.PrepareMaxInPortId(baseExecConfig.GetMaxInPortId())
	baseExec.PrepareMaxOutPortId(baseExecConfig.GetMaxOutPortId())

	// 初始化所有的输入端口
	inExecNum := 0
	for index, input := range baseExecConfig.Inputs {
		portType := strings.ToLower(input.PortType)
		if portType != Config_PortType_Exec && portType != Config_PortType_Data {
			return nil, fmt.Errorf("input %s data type %s not support", input.Name, input.DataType)
		}

		// 输入执行结点只能有一个,且只能放在第一个
		if portType == Config_PortType_Exec {
			if inExecNum > 0 {
				return nil, fmt.Errorf("inPort only allows one Execute,node name %s", baseExec.Name)
			}
			if index > 0 {
				return nil, fmt.Errorf("the exec port is only allowed to be placed on the first one,node name %s", baseExec.Name)
			}

			inExecNum++
			// 设置执行端口
			baseExec.SetInPortById(input.PortId, NewPortExec())
			continue
		}

		// 根据类型设置对应的端口
		port, err := em.createPortByDataType(baseExec.Name, input.Name, input.DataType)
		if err != nil {
			return nil, err
		}

		// 根据PortId设置端口
		baseExec.SetInPortById(input.PortId, port)
	}

	// 初始化所有的输出端口
	hasData := false
	for _, output := range baseExecConfig.Outputs {
		portType := strings.ToLower(output.PortType)
		if portType != Config_PortType_Exec && portType != Config_PortType_Data {
			return nil, fmt.Errorf("output %s data type %s not support,node name %s", output.Name, output.DataType, baseExec.Name)
		}

		// Exec出口只能先Exec，再Data，不能穿插，如果是Data类型，但遇到Exec入口则不允许
		if hasData && portType == Config_PortType_Exec {
			return nil, fmt.Errorf("the exec port can only be placed at the front,node name %s", baseExec.Name)
		}

		// 设置执行端口
		if portType == Config_PortType_Exec {
			baseExec.SetOutPortById(output.PortId, NewPortExec())
			continue
		}

		// 根据类型设置数据端口
		hasData = true
		port, err := em.createPortByDataType(baseExec.Name, output.Name, output.DataType)
		if err != nil {
			return nil, err
		}

		baseExec.SetOutPortById(output.PortId, port)
	}
	return &baseExec, nil
}

func (em *ExecPool) addInnerExec(exec IInnerExecNode) bool {
	if _, ok := em.innerExecNodeMap[exec.GetName()]; ok {
		return false
	}
	em.innerExecNodeMap[exec.GetName()] = exec
	return true
}

func (em *ExecPool) Register(exec IExecNode) bool {
	baseExec, ok := exec.(IExecNode)
	if !ok {
		return false
	}

	innerNode, ok := em.innerExecNodeMap[baseExec.GetName()]
	if !ok {
		return false
	}

	if _, ok = em.execNodeMap[innerNode.GetName()]; ok {
		return false
	}

	baseExecNode, ok := exec.(IBaseExecNode)
	if !ok {
		return false
	}

	// 设置实际执行结点中innerExecNode变量,BaseExecNode.innerExecNode = innerNode
	baseExecNode.initInnerExecNode(innerNode.(*innerExecNode))

	// innerNode设置实际的exec变量,innerExecNode.IExecNode = exec
	innerNode.SetExec(exec)

	// 将实际的执行结点保存到execNodeMap中
	em.execNodeMap[baseExec.GetName()] = exec
	return true
}

func (em *ExecPool) GetExec(name string) IInnerExecNode {
	if exec, ok := em.execNodeMap[name]; ok {
		return exec.getInnerExecNode()
	}
	return nil
}

// regVariablesNode 注册变量结点GetVar_类型、SetVar_类型
func (em *ExecPool) regVariablesNode() error {
	var err error
	if err = em.regGetVariables(Config_DataType_Int); err != nil {
		return err
	}
	if err = em.regGetVariables(Config_DataType_Integer); err != nil {
		return err
	}
	if err = em.regGetVariables(Config_DataType_Float); err != nil {
		return err
	}
	if err = em.regGetVariables(Config_DataType_Str); err != nil {
		return err
	}
	if err = em.regGetVariables(Config_DataType_Boolean); err != nil {
		return err
	}
	if err = em.regGetVariables(Config_DataType_Bool); err != nil {
		return err
	}
	if err = em.regGetVariables(Config_DataType_Array); err != nil {
		return err
	}

	if err = em.regSetVariables(Config_DataType_Int); err != nil {
		return err
	}
	if err = em.regSetVariables(Config_DataType_Integer); err != nil {
		return err
	}
	if err = em.regSetVariables(Config_DataType_Float); err != nil {
		return err
	}
	if err = em.regSetVariables(Config_DataType_Str); err != nil {
		return err
	}
	if err = em.regSetVariables(Config_DataType_Boolean); err != nil {
		return err
	}
	if err = em.regSetVariables(Config_DataType_Bool); err != nil {
		return err
	}
	if err = em.regSetVariables(Config_DataType_Array); err != nil {
		return err
	}
	return nil
}

func (em *ExecPool) regGetVariables(typ string) error {
	var baseExec innerExecNode
	baseExec.Name = genGetVariablesNodeName(typ)
	baseExec.PrepareMaxOutPortId(0)

	outPort := NewPortByType(typ)
	if outPort == nil {
		return fmt.Errorf("invalid type %s", typ)
	}
	baseExec.SetOutPortById(0, outPort)

	var getVariablesNode GetVariablesNode
	getVariablesNode.nodeName = baseExec.GetName()

	if !em.addInnerExec(&baseExec) {
		return fmt.Errorf("exec %s already registered", baseExec.GetName())
	}
	if !em.Register(&getVariablesNode) {
		return fmt.Errorf("exec %s already registered", baseExec.GetName())
	}

	return nil
}

func genSetVariablesNodeName(typ string) string {
	return fmt.Sprintf("%s_%s", SetVariables, typ)
}

func genGetVariablesNodeName(typ string) string {
	return fmt.Sprintf("%s_%s", GetVariables, typ)
}

func (em *ExecPool) regSetVariables(typ string) error {
	var baseExec innerExecNode
	baseExec.Name = genSetVariablesNodeName(typ)

	inExecPort := NewPortByType(Config_PortType_Exec)
	inPort := NewPortByType(typ)
	outExecPort := NewPortByType(Config_PortType_Exec)
	outPort := NewPortByType(typ)
	baseExec.PrepareMaxInPortId(1)
	baseExec.PrepareMaxOutPortId(1)

	baseExec.SetInPortById(0, inExecPort)
	baseExec.SetInPortById(1, inPort)

	baseExec.SetOutPortById(0, outExecPort)
	baseExec.SetOutPortById(1, outPort)

	baseExec.IExecNode = &SetVariablesNode{nodeName: baseExec.GetName()}
	if !em.addInnerExec(&baseExec) {
		return fmt.Errorf("exec %s already registered", baseExec.GetName())
	}
	if !em.Register(baseExec.IExecNode) {
		return fmt.Errorf("exec %s already registered", baseExec.GetName())
	}

	return nil
}

func getEntranceNodeNameAndID(className string) (string, int64, bool) {
	if !strings.HasPrefix(className, Entrance) {
		return "", 0, false
	}

	parts := strings.Split(className, "_")
	if len(parts) != 3 {
		return "", 0, false
	}

	entranceID, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", 0, false
	}

	return parts[0] + "_" + parts[1], int64(entranceID), true
}
