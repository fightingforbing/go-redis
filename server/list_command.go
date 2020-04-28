package server

import (
	"ddvideo/go_redis_server/consts"
	"ddvideo/go_redis_server/server/object"
	"time"
)

func lindexCommand(c *redisClient) {
	obj := c.lookupKeyRead(c.Argv[1])
	if obj == nil {
		c.addReplyNull()
		return
	}
	if obj.Type != consts.RedisList {
		c.addReplyTypeError()
		return
	}
	index,ok := c.Argv[2].Data.(int64)
	if !ok {
		c.addReplyTypeError()
		return
	}
	l := obj.Data.(*object.List)
	v := l.Index(int(index))
	if v == nil {
		c.addReplyNull()
		return
	}
	c.addReplyBulk(v)
}

func lpushCommand(c *redisClient) {
	pushGenericCommand(c, consts.RedisHead)
}

func rpushCommand(c *redisClient) {
	pushGenericCommand(c, consts.RedisTail)
}

func pushGenericCommand(c *redisClient, where int) {
	var (
		pushed int64
	)
	robj := c.lookupKey(c.Argv[0])
	if robj != nil && robj.Type != consts.RedisList {
		c.addReplyTypeError()
		return
	}
	if robj == nil {
		robj = object.CreateListObject()
		c.addKeyValue(c.Argv[1], robj)
	}
	l := robj.Data.(*object.List)
	for j := 2; j < c.Argc; j++ {
		l.Push(c.Argv[j], where)
		pushed++
	}
	c.addReplyInt64(pushed)
}
func lpopCommand(c *redisClient) {
	popGenericCommand(c, consts.RedisHead)
}

func rpopCommmand(c *redisClient) {
	popGenericCommand(c, consts.RedisTail)
}

func popGenericCommand(c *redisClient, where int) {
	obj := c.lookupKey(c.Argv[1])
	if obj == nil {
		c.addReplyNull()
	}
	if c.checkTypeOrReply(obj, consts.RedisList) {
		return
	}
	vo := obj.Data.(*object.List).Pop(where)
	if vo == nil {
		c.addReplyNull()
		return
	}
	c.addReply(vo)
}

func blpopCommand(c *redisClient) {
	blockingPopGenericCommand(c, consts.RedisHead)
}

func brpopCommand(c *redisClient) {
	blockingPopGenericCommand(c, consts.RedisTail)
}


func blockingPopGenericCommand(c *redisClient, where int) {
	var timeout int64
	timeout, ok := c.Argv[c.Argc-1].Data.(int64);
	if !ok {
		c.addReplyTimeoutError()
		return
	}
	// 取出timeout参数
	repy := make([]*object.Object,0)
	// 遍历所有列表键
	for j := 1; j < c.Argc-1; j++ {
		// 取出列表键
		o := c.lookupKey(c.Argv[j])

		if o != nil { // 有非空列表
			if o.Type != consts.RedisList {
				c.addReplyTypeError()
				return
			}
			list := o.Data.(*object.List)
			if list.Len() != 0 { // 非空列表
				po := list.Pop(where)
				if po == nil {
					panic("list pop nil")
				}
				repy := append(repy, c.Argv[j], po)
				if list.Len() == 0 { // 删除空列表
					c.deleteKey(c.Argv[j])
				}
				c.addReplyArray(repy)
			}
			return
		}
	}
	// 所有输入列表键都不存在， 只能阻塞
	blockForKeys(c, c.Argv[1:c.Argc-1], time.Duration(timeout) * time.Second)
}