package object

import (
	"ddvideo/go_redis_server/consts"
	"strings"
)

// 测试push
var pushData = []struct{
	V string
	Where int
	IsInteger bool
}{
	{"2", consts.RedisHead, true}, // 0-12 内的数字
	{"15", consts.RedisTail, true}, // 8bit有符号数字
	{"-15", consts.RedisTail, true}, // 8bit有符号数字
	{"12352", consts.RedisTail, true}, // int16类型
	{"-12352", consts.RedisTail, true}, // int16类型
	{"3158080", consts.RedisTail, true}, // 24bit有符号整数
	{"-3158080", consts.RedisTail, true}, // 24bit有符号整数
	{"338702400", consts.RedisTail, true}, // int32类型
	{"-338702400", consts.RedisTail, true}, // int32类型
	{"1454715731415412700", consts.RedisTail, true}, // int64类型
	{"-1454715731415412700", consts.RedisTail, true}, // int64类型
	{strings.Repeat("a",255),consts.RedisHead, false}, // 触发级联更新
	{"abc", consts.RedisHead, false},
	{"abcd", consts.RedisHead, false},
	{strings.Repeat("a",255), consts.RedisHead, false}, // 触发级联更新
}

