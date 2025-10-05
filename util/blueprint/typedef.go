package blueprint

type ArrayElement struct {
	IntVal int64
	StrVal string
}

type PortExec struct {
}
type ArrayData struct {
	IntVal int64
	StrVal string
}

type Port_Exec = PortExec
type Port_Int = int64
type Port_Float = float64
type Port_Str = string
type Port_Bool = bool

type Port_Array []ArrayData

type iPortType interface {
	Port_Exec | Port_Int | Port_Float | Port_Str | Port_Bool | Port_Array
}
