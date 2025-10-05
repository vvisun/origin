package stack

// Stack 是一个通用类型的栈结构
type Stack[T any] struct {
	items []*T
}

// Push 将元素压入栈顶
func (s *Stack[T]) Push(item *T) {
	s.items = append(s.items, item)
}

// Pop 弹出并返回栈顶元素，如果栈为空则返回错误
func (s *Stack[T]) Pop() *T {
	if len(s.items) == 0 {
		return nil
	}

	index := len(s.items) - 1
	item := s.items[index]
	s.items = s.items[:index]
	return item
}

// Peek 返回栈顶元素但不移除，如果栈为空则返回错误
func (s *Stack[T]) Peek() *T {
	if len(s.items) == 0 {
		return nil
	}
	return s.items[len(s.items)-1]
}

// IsEmpty 检查栈是否为空
func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

// Size 返回栈中元素的数量
func (s *Stack[T]) Size() int {
	return len(s.items)
}
