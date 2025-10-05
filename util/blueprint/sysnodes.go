package blueprint

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"math/rand/v2"
)

// 系统入口ID定义，1000以内
const (
	EntranceID_ArrayParam = 2
	EntranceID_IntParam   = 1
)

func init() {
	RegExecNode(&Entrance_ArrayParam{})
	RegExecNode(&Entrance_IntParam{})
	RegExecNode(&Output{})
	RegExecNode(&Sequence{})
	RegExecNode(&Foreach{})
	RegExecNode(&GetArrayInt{})
	RegExecNode(&GetArrayString{})
	RegExecNode(&GetArrayLen{})

	RegExecNode(&BoolIf{})
	RegExecNode(&GreaterThanInteger{})
	RegExecNode(&LessThanInteger{})
	RegExecNode(&EqualInteger{})
	RegExecNode(&RangeCompare{})
	RegExecNode(&Probability{})
}

type Entrance_ArrayParam struct {
	BaseExecNode
}

func (em *Entrance_ArrayParam) GetName() string {
	return "Entrance_ArrayParam"
}

func (em *Entrance_ArrayParam) Exec() (int, error) {
	return 0, nil
}

type Entrance_IntParam struct {
	BaseExecNode
}

func (em *Entrance_IntParam) GetName() string {
	return "Entrance_IntParam"
}

func (em *Entrance_IntParam) Exec() (int, error) {
	return 0, nil
}

type Output struct {
	BaseExecNode
}

func (em *Output) GetName() string {
	return "Output"
}

func (em *Output) Exec() (int, error) {
	val, ok := em.GetInPortInt(1)
	if !ok {
		return 0, fmt.Errorf("Output Exec inParam not found")
	}

	fmt.Printf("Output Exec inParam %d\n", val)
	return 0, nil
}

type Sequence struct {
	BaseExecNode
}

func (em *Sequence) GetName() string {
	return "Sequence"
}

func (em *Sequence) Exec() (int, error) {
	for i := range em.outPort {
		if !em.outPort[i].IsPortExec() {
			break
		}

		err := em.DoNext(i)
		if err != nil {
			return -1, err
		}
	}

	return -1, nil
}

type Foreach struct {
	BaseExecNode
}

func (em *Foreach) GetName() string {
	return "Foreach"
}

func (em *Foreach) Exec() (int, error) {
	startIndex, ok := em.ExecContext.InputPorts[1].GetInt()
	if !ok {
		return 0, fmt.Errorf("foreach Exec inParam not found")
	}
	endIndex, ok := em.ExecContext.InputPorts[2].GetInt()
	if !ok {
		return 0, fmt.Errorf("foreach Exec inParam not found")
	}

	for i := startIndex; i < endIndex; i++ {
		em.ExecContext.OutputPorts[2].SetInt(i)
		err := em.DoNext(0)
		if err != nil {
			return -1, err
		}
	}

	err := em.DoNext(1)
	if err != nil {
		return -1, err
	}

	return -1, nil
}

type GetArrayInt struct {
	BaseExecNode
}

func (em *GetArrayInt) GetName() string {
	return "GetArrayInt"
}

func (em *GetArrayInt) Exec() (int, error) {
	inPort := em.GetInPort(0)
	if inPort == nil {
		return -1, fmt.Errorf("GetArrayInt inParam not found")
	}
	outPort := em.GetOutPort(0)
	if outPort == nil {
		return -1, fmt.Errorf("GetArrayInt outParam not found")
	}

	arrIndexPort := em.GetInPort(1)
	if arrIndexPort == nil {
		return -1, fmt.Errorf("GetArrayInt arrIndexParam not found")
	}
	arrIndex, ok := arrIndexPort.GetInt()
	if !ok {
		return -1, fmt.Errorf("GetArrayInt arrIndexParam not found")
	}

	if arrIndex < 0 || arrIndex >= inPort.GetArrayLen() {
		return -1, fmt.Errorf("GetArrayInt arrIndexParam out of range,index %d", arrIndex)
	}

	val, ok := inPort.GetArrayValInt(int(arrIndex))
	if !ok {
		log.Errorf("GetArrayValInt failed, idx:%d", arrIndex)
		return -1, fmt.Errorf("GetArrayInt inParam not found")
	}

	outPort.SetInt(val)
	return -1, nil
}

type GetArrayString struct {
	BaseExecNode
}

func (em *GetArrayString) GetName() string {
	return "GetArrayString"
}

func (em *GetArrayString) Exec() (int, error) {
	inPort := em.GetInPort(0)
	if inPort == nil {
		return -1, fmt.Errorf("GetArrayInt inParam 0 not found")
	}
	outPort := em.GetOutPort(0)
	if outPort == nil {
		return -1, fmt.Errorf("GetArrayInt outParam 0 not found")
	}

	arrIndexPort := em.GetInPort(1)
	if arrIndexPort == nil {
		return -1, fmt.Errorf("GetArrayInt arrIndexParam 1 not found")
	}
	arrIndex, ok := arrIndexPort.GetInt()
	if !ok {
		return -1, fmt.Errorf("GetArrayInt arrIndexParam not found")
	}

	if arrIndex < 0 || arrIndex >= inPort.GetArrayLen() {
		return -1, fmt.Errorf("GetArrayInt arrIndexParam out of range,index %d", arrIndex)
	}

	val, ok := inPort.GetArrayValStr(int(arrIndex))
	if !ok {
		log.Errorf("GetArrayValStr failed, idx:%d", arrIndex)
		return -1, fmt.Errorf("GetArrayInt inParam not found")
	}

	outPort.SetStr(val)
	return -1, nil
}

type GetArrayLen struct {
	BaseExecNode
}

func (em *GetArrayLen) GetName() string {
	return "GetArrayLen"
}

func (em *GetArrayLen) Exec() (int, error) {
	inPort := em.GetInPort(0)
	if inPort == nil {
		return -1, fmt.Errorf("GetArrayInt inParam 0 not found")
	}
	outPort := em.GetOutPort(0)
	if outPort == nil {
		return -1, fmt.Errorf("GetArrayInt outParam 0 not found")
	}

	outPort.SetInt(inPort.GetArrayLen())
	return -1, nil
}

// BoolIf 布尔判断
type BoolIf struct {
	BaseExecNode
}

func (em *BoolIf) GetName() string {
	return "BoolIf"
}

func (em *BoolIf) Exec() (int, error) {
	inPort := em.GetInPort(1)
	if inPort == nil {
		return -1, fmt.Errorf("GetArrayInt inParam 1 not found")
	}

	ret, ok := inPort.GetBool()
	if !ok {
		return -1, fmt.Errorf("BoolIf inParam error")
	}

	if ret {
		return 1, nil
	}

	return 0, nil
}

// GreaterThanInteger 大于(整型) >
type GreaterThanInteger struct {
	BaseExecNode
}

func (em *GreaterThanInteger) GetName() string {
	return "GreaterThanInteger"
}

func (em *GreaterThanInteger) Exec() (int, error) {
	inPortEqual := em.GetInPort(1)
	if inPortEqual == nil {
		return -1, fmt.Errorf("GreaterThanInteger inParam 1 not found")
	}

	inPortA := em.GetInPort(2)
	if inPortA == nil {
		return -1, fmt.Errorf("GreaterThanInteger inParam 1 not found")
	}

	inPorB := em.GetInPort(3)
	if inPorB == nil {
		return -1, fmt.Errorf("GreaterThanInteger inParam 1 not found")
	}

	ret, ok := inPortEqual.GetBool()
	if !ok {
		return -1, fmt.Errorf("GreaterThanInteger inParam 1 error")
	}

	inA, ok := inPortA.GetInt()
	if !ok {
		return -1, fmt.Errorf("GreaterThanInteger inParam 2 error")
	}
	inB, ok := inPorB.GetInt()
	if !ok {
		return -1, fmt.Errorf("GreaterThanInteger inParam 3 error")
	}
	if ret {
		if inA >= inB {
			return 1, nil
		}
		return 0, nil
	}

	if inA > inB {
		return 1, nil
	}
	return 0, nil
}

// LessThanInteger 小于(整型) <
type LessThanInteger struct {
	BaseExecNode
}

func (em *LessThanInteger) GetName() string {
	return "LessThanInteger"
}

func (em *LessThanInteger) Exec() (int, error) {
	inPortEqual := em.GetInPort(1)
	if inPortEqual == nil {
		return -1, fmt.Errorf("GreaterThanInteger inParam 1 not found")
	}

	inPortA := em.GetInPort(2)
	if inPortA == nil {
		return -1, fmt.Errorf("GreaterThanInteger inParam 1 not found")
	}

	inPorB := em.GetInPort(3)
	if inPorB == nil {
		return -1, fmt.Errorf("GreaterThanInteger inParam 1 not found")
	}

	ret, ok := inPortEqual.GetBool()
	if !ok {
		return -1, fmt.Errorf("GreaterThanInteger inParam 1 error")
	}

	inA, ok := inPortA.GetInt()
	if !ok {
		return -1, fmt.Errorf("GreaterThanInteger inParam 2 error")
	}
	inB, ok := inPorB.GetInt()
	if !ok {
		return -1, fmt.Errorf("GreaterThanInteger inParam 3 error")
	}
	if ret {
		if inA <= inB {
			return 1, nil
		}
		return 0, nil
	}

	if inA < inB {
		return 1, nil
	}
	return 0, nil
}

// EqualInteger 等于(整型)==
type EqualInteger struct {
	BaseExecNode
}

func (em *EqualInteger) GetName() string {
	return "EqualInteger"
}

func (em *EqualInteger) Exec() (int, error) {

	inPortA := em.GetInPort(1)
	if inPortA == nil {
		return -1, fmt.Errorf("GreaterThanInteger inParam 1 not found")
	}

	inPorB := em.GetInPort(2)
	if inPorB == nil {
		return -1, fmt.Errorf("GreaterThanInteger inParam 1 not found")
	}

	inA, ok := inPortA.GetInt()
	if !ok {
		return -1, fmt.Errorf("GreaterThanInteger inParam 2 error")
	}
	inB, ok := inPorB.GetInt()
	if !ok {
		return -1, fmt.Errorf("GreaterThanInteger inParam 3 error")
	}

	if inA == inB {
		return 1, nil
	}
	return 0, nil
}

// RangeCompare 范围比较<=
type RangeCompare struct {
	BaseExecNode
}

func (em *RangeCompare) GetName() string {
	return "RangeCompare"
}

func (em *RangeCompare) Exec() (int, error) {

	inPortA := em.GetInPort(1)
	if inPortA == nil {
		return -1, fmt.Errorf("GreaterThanInteger inParam 1 not found")
	}

	ret, ok := inPortA.GetInt()
	if !ok {
		return -1, fmt.Errorf("GreaterThanInteger inParam 1 error")
	}

	intArray := em.execNode.GetInPortDefaultIntArrayValue(2)
	if intArray == nil {
		return 0, nil
	}

	for i := range intArray {
		if intArray[i] <= ret {
			return i + 2, nil
		}
	}

	return 0, nil
}

// Probability 概率判断(万分比)
type Probability struct {
	BaseExecNode
}

func (em *Probability) GetName() string {
	return "Probability"
}

func (em *Probability) Exec() (int, error) {
	inPortProbability := em.GetInPort(1)
	if inPortProbability == nil {
		return -1, fmt.Errorf("GreaterThanInteger inParam 1 not found")
	}

	inProbability, ok := inPortProbability.GetInt()
	if !ok {
		return -1, fmt.Errorf("GreaterThanInteger inParam 1 error")
	}

	if inProbability > rand.Int64N(10000) {
		return 1, nil
	}

	return 0, nil
}
