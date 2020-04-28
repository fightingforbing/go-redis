package object

import (
	"ddvideo/go_redis_server/consts"
	"strconv"
)

type Object struct {
	Type int
	Data interface{}
}

type Lener interface {
	Len() int
}

func CreateObject(s string) *Object {
	i, ok := TryEncodingInteger(s)
	if ok { // 整数
		return CreateIntObject(i)
	}
	return CreateStringObject(s)
}

func CreateIntObject(v int64) *Object {
	return &Object{
		Type:        consts.RedisInt,
		Data:        v,
	}
}

func CreateListObject() *Object {
	return &Object {
		Type: consts.RedisList,
		Data: createList(),
	}
}

func CreateSetObject() *Object {
	return &Object{
		Type: consts.RedisSet,
		Data: createSet(),
	}
}

func CreateZsetObject() *Object {
	return &Object{
		Type: consts.RedisZset,
		Data: createZset(),
	}
}

func CreateStringObject(s string) *Object {
	return &Object{
		Type:        consts.RedisString,
		Data:        s,
	}
}

func CreateFloat64Object(v float64) *Object {
	return &Object{
		Type: consts.RedisFloat,
		Data: v,
	}
}

func CreateFloat64ObjectByString(s string) (*Object, bool) {
	f, err := strconv.ParseFloat(s,64)
	if err != nil {
		return nil, false
	}
	return CreateFloat64Object(f), true
}