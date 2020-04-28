package server

import (
	"ddvideo/go_redis_server/consts"
	"ddvideo/go_redis_server/server/object"
	"strings"
)

// 不区分大小写字符串判断
func EqualIngoreCase(src string, dest string) bool {
	if len(src) != len(dest) {
		return false
	}
	srcLower := strings.ToLower(src)
	destLower := strings.ToLower(dest)
	if srcLower != destLower {
		return false
	}
	return true
}

func StrcmpObjectAndStringIngoreCase(o *object.Object, s string) bool {
	if o.Type != consts.RedisEncodingRaw {
		return false
	}
	so := object.ConvertStringTypeObjectToString(o)
	return EqualIngoreCase(so, s)
}

