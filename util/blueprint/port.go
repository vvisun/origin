package blueprint

import (
	"fmt"
	"strconv"
)

const (
	Config_PortType_Exec    = "exec"
	Config_PortType_Data    = "data"
	Config_DataType_Int     = "int"
	Config_DataType_Integer = "integer"
	Config_DataType_Float   = "float"
	Config_DataType_Str     = "string"
	Config_DataType_Boolean = "boolean"
	Config_DataType_Bool    = "bool"
	Config_DataType_Array   = "array"
)

type Port[T iPortType] struct {
	PortVal T
}

func (em *Port[T]) Clone() IPort {
	return &Port[T]{
		PortVal: em.PortVal,
	}
}

func (em *Port[T]) Reset() {
	var v T
	em.PortVal = v
}

func (em *Port[T]) GetInt() (Port_Int, bool) {
	if t, ok := any(em.PortVal).(Port_Int); ok {
		return t, true
	}
	return 0, false
}

func (em *Port[T]) GetFloat() (Port_Float, bool) {
	if t, ok := any(em.PortVal).(Port_Float); ok {
		return t, true
	}
	return 0, false
}

func (em *Port[T]) GetStr() (Port_Str, bool) {
	if t, ok := any(em.PortVal).(Port_Str); ok {
		return t, true
	}
	return "", false
}

func (em *Port[T]) GetArrayValInt(idx int) (Port_Int, bool) {
	if t, ok := any(em.PortVal).(Port_Array); ok {
		if idx >= 0 && idx < len(t) {
			return t[idx].IntVal, true
		}
	}

	return 0, false
}

func (em *Port[T]) GetArrayValStr(idx int) (string, bool) {
	if t, ok := any(em.PortVal).(Port_Array); ok {
		if idx >= 0 && idx < len(t) {
			return t[idx].StrVal, true
		}
	}

	return "", false
}

func (em *Port[T]) GetBool() (Port_Bool, bool) {
	if t, ok := any(em.PortVal).(Port_Bool); ok {
		return t, true
	}
	return false, false
}

func (em *Port[T]) GetArray() (Port_Array, bool) {
	if t, ok := any(em.PortVal).(Port_Array); ok {
		return t, true
	}
	return nil, false
}

func (em *Port[T]) SetInt(val Port_Int) bool {
	if t, ok := any(&em.PortVal).(*Port_Int); ok {
		*t = val
		return true
	}
	return false
}

func (em *Port[T]) SetFloat(val Port_Float) bool {
	if t, ok := any(&em.PortVal).(*Port_Float); ok {
		*t = val
		return true
	}
	return false
}

func (em *Port[T]) SetStr(val Port_Str) bool {
	if t, ok := any(&em.PortVal).(*Port_Str); ok {
		*t = val
		return true
	}
	return false
}

func (em *Port[T]) SetBool(val Port_Bool) bool {
	if t, ok := any(&em.PortVal).(*Port_Bool); ok {
		*t = val
		return true
	}
	return false
}

func (em *Port[T]) SetArrayValInt(idx int, val Port_Int) bool {
	if t, ok := any(em.PortVal).(Port_Array); ok {
		if idx >= 0 && idx < len(t) {
			t[idx].IntVal = val
			return true
		}
	}
	return false
}

func (em *Port[T]) SetArrayValStr(idx int, val Port_Str) bool {
	if t, ok := any(em.PortVal).(Port_Array); ok {
		if idx >= 0 && idx < len(t) {
			(t)[idx].StrVal = val
			return true
		}
	}
	return false
}

func (em *Port[T]) AppendArrayValInt(val Port_Int) bool {
	if t, ok := any(&em.PortVal).(*Port_Array); ok {
		*t = append(*t, ArrayData{IntVal: val})
		return true
	}
	return false
}

func (em *Port[T]) AppendArrayValStr(val Port_Str) bool {
	if t, ok := any(&em.PortVal).(*Port_Array); ok {
		*t = append(*t, ArrayData{StrVal: val})
		return true
	}
	return false
}

func (em *Port[T]) GetArrayLen() Port_Int {
	if t, ok := any(&em.PortVal).(*Port_Array); ok {
		return Port_Int(len(*t))
	}

	return 0
}

func (em *Port[T]) IsPortExec() bool {
	_, ok := any(em.PortVal).(Port_Exec)
	return ok
}

func (em *Port[T]) convertInt64(v any) (int64, bool) {
	switch v.(type) {
	case int:
		return int64(v.(int)), true
	case int64:
		return v.(int64), true
	case int32:
		return int64(v.(int32)), true
	case int16:
		return int64(v.(int16)), true
	case int8:
		return int64(v.(int8)), true
	case uint64:
		return int64(v.(uint64)), true
	case uint32:
		return int64(v.(uint32)), true
	case uint16:
		return int64(v.(uint16)), true
	case uint8:
		return int64(v.(uint8)), true
	case uint:
		return int64(v.(uint)), true
	default:
		return 0, false
	}
}

func (em *Port[T]) setAnyVale(v any) error {
	switch v.(type) {
	case int, int64:
		val, ok := em.convertInt64(v)
		if !ok {
			return fmt.Errorf("port type is %T, but value is %v", em.PortVal, v)
		}

		switch any(em.PortVal).(type) {
		case Port_Int:
			em.SetInt(val)
		case Port_Float:
			em.SetFloat(Port_Float(val))
		case Port_Str:
			em.SetStr(fmt.Sprintf("%d", int64(val)))
		case Port_Bool:
			em.SetBool(int64(val) != 0)
		default:
			return fmt.Errorf("port type is %T, but value is %v", em.PortVal, v)
		}
	case float64:
		fV := v.(float64)
		switch any(em.PortVal).(type) {
		case Port_Int:
			em.SetInt(Port_Int(fV))
		case Port_Float:
			em.SetFloat(fV)
		case Port_Str:
			em.SetStr(fmt.Sprintf("%d", int64(fV)))
		case Port_Bool:
			em.SetBool(int64(fV) != 0)
		default:
			return fmt.Errorf("port type is %T, but value is %v", em.PortVal, v)
		}
	case string:
		strV := v.(string)
		switch any(em.PortVal).(type) {
		case Port_Int:
			val, err := strconv.Atoi(strV)
			if err != nil {
				return err
			}
			em.SetInt(Port_Int(val))
		case Port_Float:
			fV, err := strconv.ParseFloat(strV, 64)
			if err != nil {
				return err
			}
			em.SetFloat(fV)
		case Port_Str:
			em.SetStr(strV)
		case Port_Bool:
			val, err := strconv.ParseBool(strV)
			if err != nil {
				return err
			}
			em.SetBool(val)
		default:
			return fmt.Errorf("port type is %T, but value is %v", em.PortVal, v)
		}
	case bool:
		strV := v.(bool)
		switch any(em.PortVal).(type) {
		case Port_Int:
			return fmt.Errorf("port type is int, but value is %v", strV)
		case Port_Float:
			return fmt.Errorf("port type is float, but value is %v", strV)
		case Port_Str:
			return fmt.Errorf("port type is string, but value is %v", strV)
		case Port_Bool:
			em.SetBool(strV)
		default:
			return fmt.Errorf("port type is %T, but value is %v", em.PortVal, v)
		}
	case []int64:
		arr := v.([]int64)
		for _, val := range arr {
			em.AppendArrayValInt(val)
		}
	case []int32:
		arr := v.([]int32)
		for _, val := range arr {
			em.AppendArrayValInt(Port_Int(val))
		}
	case []int16:
		arr := v.([]int16)
		for _, val := range arr {
			em.AppendArrayValInt(Port_Int(val))
		}
	case []int8:
		arr := v.([]int8)
		for _, val := range arr {
			em.AppendArrayValInt(Port_Int(val))
		}
	case []uint64:
		arr := v.([]uint64)
		for _, val := range arr {
			em.AppendArrayValInt(Port_Int(val))
		}
	case []uint32:
		arr := v.([]uint32)
		for _, val := range arr {
			em.AppendArrayValInt(Port_Int(val))
		}
	case []uint16:
		arr := v.([]uint16)
		for _, val := range arr {
			em.AppendArrayValInt(Port_Int(val))
		}
	case []uint8:
		arr := v.([]uint8)
		for _, val := range arr {
			em.AppendArrayValInt(Port_Int(val))
		}
	case []string:
		arr := v.([]string)
		for _, val := range arr {
			em.AppendArrayValStr(val)
		}
	}

	return nil
}

func (em *Port[T]) SetValue(val IPort) bool {
	valT, ok := val.(*Port[T])
	if !ok {
		return false
	}
	em.PortVal = valT.PortVal
	return true
}

type IPort interface {
	GetInt() (Port_Int, bool)
	GetFloat() (Port_Float, bool)
	GetStr() (Port_Str, bool)
	GetArrayValInt(idx int) (Port_Int, bool)
	GetArrayValStr(idx int) (Port_Str, bool)
	GetBool() (Port_Bool, bool)
	GetArray() (Port_Array, bool)

	SetInt(val Port_Int) bool
	SetFloat(val Port_Float) bool
	SetStr(val Port_Str) bool
	SetBool(val Port_Bool) bool
	SetArrayValInt(idx int, val Port_Int) bool
	SetArrayValStr(idx int, val Port_Str) bool
	AppendArrayValInt(val Port_Int) bool
	AppendArrayValStr(val Port_Str) bool
	GetArrayLen() Port_Int
	Clone() IPort
	Reset()

	IsPortExec() bool

	setAnyVale(v any) error
	SetValue(val IPort) bool
}

func NewPortExec() IPort {
	return &Port[Port_Exec]{}
}

func NewPortInt() IPort {
	return &Port[Port_Int]{}
}

func NewPortFloat() IPort {
	return &Port[Port_Float]{}
}

func NewPortStr() IPort {
	return &Port[Port_Str]{}
}

func NewPortBool() IPort {
	return &Port[Port_Bool]{}
}

func NewPortArray() IPort {
	return &Port[Port_Array]{}
}

func NewPortByType(typ string) IPort {
	switch typ {
	case Config_PortType_Exec:
		return NewPortExec()
	case Config_DataType_Int, Config_DataType_Integer:
		return NewPortInt()
	case Config_DataType_Float:
		return NewPortFloat()
	case Config_DataType_Str:
		return NewPortStr()
	case Config_DataType_Bool, Config_DataType_Boolean:
		return NewPortBool()
	case Config_DataType_Array:
		return NewPortArray()
	default:
		return nil
	}
}
