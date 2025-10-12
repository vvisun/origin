package typ

import "errors"

type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Float interface {
	~float32 | ~float64
}

type Integer interface {
	Signed | Unsigned
}

type Number interface {
	Signed | Unsigned | Float
}

type Ordered interface {
	Number | Float | ~string
}

func ConvertToNumber[DType Number](val interface{}) (DType, error) {
	switch v := val.(type) {
	case int64:
		return DType(v), nil
	case int:
		return DType(v), nil
	case uint:
		return DType(v), nil
	case uint64:
		return DType(v), nil
	case float32:
		return DType(v), nil
	case float64:
		return DType(v), nil
	case int32:
		return DType(v), nil
	case uint32:
		return DType(v), nil
	case int16:
		return DType(v), nil
	case uint16:
		return DType(v), nil
	}

	return 0, errors.New("unsupported type")
}
