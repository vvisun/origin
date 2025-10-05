package blueprint

import (
	"fmt"
	"math/rand"
)

func init() {
	RegExecNode(&AddInt{})
	RegExecNode(&SubInt{})
	RegExecNode(&MulInt{})
	RegExecNode(&DivInt{})
	RegExecNode(&ModInt{})
	RegExecNode(&RandNumber{})
}

// AddInt 加(int)
type AddInt struct {
	BaseExecNode
}

func (em *AddInt) GetName() string {
	return "AddInt"
}

func (em *AddInt) Exec() (int, error) {
	inPortA := em.GetInPort(0)
	if inPortA == nil {
		return -1, fmt.Errorf("AddInt inParam 1 not found")
	}

	inPortB := em.GetInPort(1)
	if inPortB == nil {
		return -1, fmt.Errorf("AddInt inParam 2 not found")
	}

	outPortRet := em.GetOutPort(0)
	if outPortRet == nil {
		return -1, fmt.Errorf("AddInt outParam not found")
	}

	inA, ok := inPortA.GetInt()
	if !ok {
		return -1, fmt.Errorf("AddInt inParam 1 not found")
	}
	inB, ok := inPortB.GetInt()
	if !ok {
		return -1, fmt.Errorf("AddInt inParam 2 not found")
	}
	ret := inA + inB
	outPortRet.SetInt(ret)

	return -1, nil
}

// SubInt 减(int)
type SubInt struct {
	BaseExecNode
}

func (em *SubInt) GetName() string {
	return "SubInt"
}

func (em *SubInt) Exec() (int, error) {
	inPortA := em.GetInPort(0)
	if inPortA == nil {
		return -1, fmt.Errorf("AddInt inParam 1 not found")
	}

	inPortB := em.GetInPort(1)
	if inPortB == nil {
		return -1, fmt.Errorf("AddInt inParam 2 not found")
	}

	inPortAbs := em.GetInPort(2)
	if inPortAbs == nil {
		return -1, fmt.Errorf("AddInt inParam 2 not found")
	}

	outPortRet := em.GetOutPort(0)
	if outPortRet == nil {
		return -1, fmt.Errorf("AddInt outParam not found")
	}

	inA, ok := inPortA.GetInt()
	if !ok {
		return -1, fmt.Errorf("AddInt inParam 1 not found")
	}
	inB, ok := inPortB.GetInt()
	if !ok {
		return -1, fmt.Errorf("AddInt inParam 2 not found")
	}
	isAbs, ok := inPortAbs.GetBool()
	if !ok {
		return -1, fmt.Errorf("AddInt inParam 2 not found")
	}
	ret := inA - inB
	if isAbs && ret < 0 {
		ret *= -1
	}

	outPortRet.SetInt(ret)

	return -1, nil
}

// MulInt 乘(int)
type MulInt struct {
	BaseExecNode
}

func (em *MulInt) GetName() string {
	return "MulInt"
}

func (em *MulInt) Exec() (int, error) {
	inPortA := em.GetInPort(0)
	if inPortA == nil {
		return -1, fmt.Errorf("AddInt inParam 1 not found")
	}

	inPortB := em.GetInPort(1)
	if inPortB == nil {
		return -1, fmt.Errorf("AddInt inParam 2 not found")
	}

	outPortRet := em.GetOutPort(0)
	if outPortRet == nil {
		return -1, fmt.Errorf("AddInt outParam not found")
	}

	inA, ok := inPortA.GetInt()
	if !ok {
		return -1, fmt.Errorf("AddInt inParam 1 not found")
	}
	inB, ok := inPortB.GetInt()
	if !ok {
		return -1, fmt.Errorf("AddInt inParam 2 not found")
	}

	outPortRet.SetInt(inA * inB)

	return -1, nil
}

// DivInt 除(int)
type DivInt struct {
	BaseExecNode
}

func (em *DivInt) GetName() string {
	return "DivInt"
}

func (em *DivInt) Exec() (int, error) {
	inPortA := em.GetInPort(0)
	if inPortA == nil {
		return -1, fmt.Errorf("AddInt inParam 1 not found")
	}

	inPortB := em.GetInPort(1)
	if inPortB == nil {
		return -1, fmt.Errorf("AddInt inParam 2 not found")
	}

	inPortRound := em.GetInPort(2)
	if inPortRound == nil {
		return -1, fmt.Errorf("AddInt inParam 2 not found")
	}

	outPortRet := em.GetOutPort(0)
	if outPortRet == nil {
		return -1, fmt.Errorf("AddInt outParam not found")
	}

	inA, ok := inPortA.GetInt()
	if !ok {
		return -1, fmt.Errorf("AddInt inParam 1 not found")
	}
	inB, ok := inPortB.GetInt()
	if !ok {
		return -1, fmt.Errorf("AddInt inParam 2 not found")
	}
	isRound, ok := inPortRound.GetBool()
	if !ok {
		return -1, fmt.Errorf("AddInt inParam 2 not found")
	}

	if inB == 0 {
		return -1, fmt.Errorf("div zero error")
	}

	var ret int64
	if isRound {
		ret = (inA + inB/2) / inB
	} else {
		ret = inA / inB
	}

	outPortRet.SetInt(ret)

	return -1, nil
}

// ModInt 取模(int)
type ModInt struct {
	BaseExecNode
}

func (em *ModInt) GetName() string {
	return "ModInt"
}

func (em *ModInt) Exec() (int, error) {
	inPortA := em.GetInPort(0)
	if inPortA == nil {
		return -1, fmt.Errorf("AddInt inParam 1 not found")
	}

	inPortB := em.GetInPort(1)
	if inPortB == nil {
		return -1, fmt.Errorf("AddInt inParam 2 not found")
	}

	outPortRet := em.GetOutPort(0)
	if outPortRet == nil {
		return -1, fmt.Errorf("AddInt outParam not found")
	}

	inA, ok := inPortA.GetInt()
	if !ok {
		return -1, fmt.Errorf("AddInt inParam 1 not found")
	}
	inB, ok := inPortB.GetInt()
	if !ok {
		return -1, fmt.Errorf("AddInt inParam 2 not found")
	}
	if inB == 0 {
		return -1, fmt.Errorf("mod zero error")
	}

	outPortRet.SetInt(inA % inB)

	return -1, nil
}

// RandNumber 范围随机[0,99]
type RandNumber struct {
	BaseExecNode
}

func (em *RandNumber) GetName() string {
	return "RandNumber"
}

func (em *RandNumber) Exec() (int, error) {
	inPortSeed := em.GetInPort(0)
	if inPortSeed == nil {
		return -1, fmt.Errorf("AddInt inParam 1 not found")
	}

	inPortMin := em.GetInPort(1)
	if inPortMin == nil {
		return -1, fmt.Errorf("AddInt inParam 2 not found")
	}

	inPortMax := em.GetInPort(2)
	if inPortMax == nil {
		return -1, fmt.Errorf("AddInt inParam 2 not found")
	}

	outPortRet := em.GetOutPort(0)
	if outPortRet == nil {
		return -1, fmt.Errorf("AddInt outParam not found")
	}

	inSeed, ok := inPortSeed.GetInt()
	if !ok {
		return -1, fmt.Errorf("AddInt inParam 1 not found")
	}
	inMin, ok := inPortMin.GetInt()
	if !ok {
		return -1, fmt.Errorf("AddInt inParam 2 not found")
	}
	inMax, ok := inPortMax.GetInt()
	if !ok {
		return -1, fmt.Errorf("AddInt inParam 2 not found")
	}

	var ret int64
	if inSeed > 0 {
		r := rand.New(rand.NewSource(inSeed))
		if r == nil {
			return -1, fmt.Errorf("RandNumber fail")
		}
		ret = int64(r.Intn(int(inMax-inMin+1))) + inMin
	} else {
		ret = int64(rand.Intn(int(inMax-inMin+1))) + inMin
	}

	outPortRet.SetInt(ret)
	return -1, nil
}
