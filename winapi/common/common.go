package common

// Boo2Ptr 将bool值转换为uintptr值
func Boo2Ptr(b bool) uintptr {
	var i uintptr
	if b {
		i = 1
	}
	return i
}
