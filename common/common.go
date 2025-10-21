package comm

import (
	"log"
	"regexp"
)

// Boo2Ptr 将bool值转换为uintptr值
func Boo2Ptr(b bool) uintptr {
	var i uintptr
	if b {
		i = 1
	}
	return i
}

// InArray 判断是否存在列表中，如果regex为true，则进行正则匹配
func InArray(list []string, value string, regex bool) bool {
	for _, v := range list {
		if regex {
			if ok, err := regexp.Match(v, []byte(value)); ok {
				return true
			} else if err != nil {
				log.Println(err.Error())
			}
		} else {
			if value == v {
				return true
			}
		}
	}
	return false
}
