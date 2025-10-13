package usync

import "reflect"

// 检查指针是否是空值
func IsNil(comp interface{}) bool {
	v := reflect.ValueOf(comp)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

// 判断是否是双指针
func IsDoublePointer(v interface{}) bool {
	t := reflect.TypeOf(v)
	k := t.Kind()
	if k == reflect.Ptr {
		t = t.Elem()
		k = t.Kind()
		if k == reflect.Ptr {
			return true
		}
	}
	return false
}

// 判断是否是指针
func IsPointer(v interface{}) bool {
	return reflect.ValueOf(v).Kind() == reflect.Ptr
}
