package server

import (
	"ddvideo/go_redis_server/consts"
	"ddvideo/go_redis_server/server/object"
)

func saddCommand(c *redisClient) {
	var (
		added int64
	)
	o := c.lookupKeyWrite(c.Argv[1])
	if  o == nil {
		o = object.CreateSetObject()
		c.addKeyValue(c.Argv[1],o)
	}
	if c.checkTypeOrReply(o, consts.RedisSet) {
		return
	}
	set := o.Data.(*object.Set)
	// 将所有输入元素增加到集合中
	for j := 2; j < c.Argc; j++ {
		if set.Add(c.Argv[j]) {
			added++
		}
	}
	c.addReplyInt64(added)
}

func sismemberCommand(c *redisClient) {
	o := c.lookupKeyRead(c.Argv[1])
	if o == nil {
		c.addReplyInt64(0)
		return
	}
	set := o.Data.(*object.Set)
	find := set.IsMember(c.Argv[2])
	if find {
		c.addReplyInt64(1)
		return
	}
	c.addReplyInt64(0)
}

func sremCommand(c *redisClient) {
	var (
		deleted int64
	)
	o := c.lookupKeyWrite(c.Argv[1])
	if o == nil {
		c.addReplyInt64(0)
		return
	}
	if o.Type != consts.RedisSet {
		c.addReplyTypeError()
		return
	}
	set := o.Data.(*object.Set)
	for j := 2; j < c.Argc; j++ {
		if set.Remove(c.Argv[j]) {
			deleted++
			if set.Len() == 0 { // 已经是空
				c.deleteKey(c.Argv[1])
			}
		}
	}
	c.addReplyInt64(deleted)
}
