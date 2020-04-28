package server

import (
	"ddvideo/go_redis_server/consts"
	"ddvideo/go_redis_server/server/object"
)

func zaddCommand(c *redisClient) {
	zaddGenericCommand(c, false)
}

func zaddGenericCommand(c *redisClient, incr bool) {
	key := c.Argv[1]
	added := int64(0)
	// 输入的score - member 参数必须是成对出现的
	if c.Argc % 2 != 0 {
		c.addReplySyntaxError()
		return
	}
	// 取出所有输入的score分值
	elements := (c.Argc - 2) / 2
	scores := make([]float64, elements)
	for j := 0; j < elements; j++ {
		f,ok := object.ConvertObjectToFloat64(c.Argv[2 + j * 2])
		if !ok {
			c.addReplyTypeError()
			return
		}
		scores[j] = f
	}
	// 取出有序集合对象
	zobj := c.lookupKeyWrite(key)
	if zobj == nil {
		zobj = object.CreateZsetObject()
		c.setKey(key, zobj)
	} else {
		if c.checkTypeOrReply(zobj, consts.RedisZset) {
			return
		}
	}
	zset := zobj.Data.(*object.Zset)
	// 处理所有元素
	for j := 0; j < elements; j++ {
		score := scores[j]
		ele := c.Argv[3 + j * 2]
		// 查找存不存在
		if zset.Insert(ele, score) {
			added++
		}
	}
	c.addReplyInt64(added)
}

func zremCommand(c *redisClient) {
	key := c.Argv[1]
	zobj := c.lookupKeyWrite(key)
	if zobj == nil {
		c.addReplyInt64(0)
		return
	}
	if c.checkTypeOrReply(zobj,consts.RedisZset) {
		return
	}
	zset := zobj.Data.(*object.Zset)
	// 遍历所有输入元素
	for j := 2; j < c.Argc; j++ {
		zset.Delete(c.Argv[j])
	}
}

func zrangeCommand(c *redisClient) {
	zrangeGenericCommand(c, false)
}

func zrevrangeCommand(c *redisClient) {
	zrangeGenericCommand(c, true)
}

func zrangeGenericCommand(c *redisClient, reverse bool) {
	var (
		start int64
		end int64
		withscores bool
		key = c.Argv[1]
	)
	start, ok := c.Argv[2].Data.(int64)
	if !ok {
		c.addReplyTypeError()
		return
	}
	end, ok := c.Argv[3].Data.(int64)
	if !ok {
		c.addReplyTypeError()
		return
	}
	// 确定是否显示分值
	if c.Argc == 5 && StrcmpObjectAndStringIngoreCase(c.Argv[4], "withscores") {
		withscores = true
	} else if c.Argc >= 5 {
		c.addReplySyntaxError()
		return
	}

	zobj := c.lookupKeyRead(key)
	if zobj == nil {
		c.addReplyNulMultilBulk()
		return
	}
	if c.checkTypeOrReply(zobj, consts.RedisZset) {
		return
	}
	zset := zobj.Data.(*object.Zset)
	// 将负数索引转化为正数索引
	l := zset.Len()
	if start < 0 { start = int64(l) + start }
	if end < 0 { end = int64(l) + end }
	if start < 0 { start = 0 }
	if start > end || start > int64(l) {
		c.addReplyNulMultilBulk()
		return
	}
	if end >= int64(l) { end = int64(l) - 1 }
	rangelen := (end - start) + 1
	for j := rangelen; j >= 0; j--{

	}
}