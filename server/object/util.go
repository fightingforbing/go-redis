package object

import (
	"ddvideo/go_redis_server/consts"
	"strconv"
	"strings"
)

func TryEncodingInteger(s string) (int64, bool) {
	v ,err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return v,true
}

func ConvertStringTypeObjectToString(o *Object) string {
	if o.Type == consts.RedisEncodingInt {
		return strconv.FormatInt(o.Data.(int64), 10)
	}
	return o.Data.(string)
}

func CompareStringObject(a, b *Object) int {
	aStr := ConvertStringTypeObjectToString(a)
	bStr := ConvertStringTypeObjectToString(b)
	return strings.Compare(aStr,bStr)
}

func ConvertObjectToFloat64(o *Object) (float64, bool) {
	s := ConvertStringTypeObjectToString(o)
	f64, err := strconv.ParseFloat(s,64)
	if err != nil {
		return 0, false
	}
	return f64, true
}