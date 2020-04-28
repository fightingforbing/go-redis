package server

import (
	"ddvideo/go_redis_server/consts"
	"ddvideo/go_redis_server/server/object"
)



func multiCommand(c *redisClient) {
	if (c.Flags & consts.RedisMulti) != 0 {
		c.addReplyError("MULTI calss can not be nested")
		return
	}
	c.Flags |= consts.RedisMulti
	c.addReplyOK()
}

// 将一个新命令增加到事务队列中
func queueMultiCommand(c *redisClient) {
	argv := make([]*object.Object, len(c.Argv))
	copy(argv, c.Argv)
	c.Mstate = append(c.Mstate, &multiState{
		Argv: argv,
		Argc: c.Argc,
		Cmd:  c.Cmd,
	})
}

func execCommand(c *redisClient) {
	if (c.Flags & consts.RedisMulti) == 0 {
		c.addReplyError("EXEC withour MULTI")
	}
	// 检查是否需要阻止事务执行，因为：
	// 1.有被监视的键已经被修改了
	// 2.命令在入队时发生错误
	//  第一种情况返回多个批量回复的空对象
	// 第二种情况返回一个EXECABORT错误
	if c.Flags & (consts.RedisDirtyCas | consts.RedisDitryExec) != 0 {
		if c.Flags & consts.RedisDitryExec != 0  {
			c.addReplyError("EXECABORT Transaction discarded because of previous error")
		} else {
			c.addReplyNulMultilBulk()
		}
		return
	}
	// 已经可以保证安全性了， 取消客户端对所有键对监视
	unwatchAllKeys(c)

	// 因为事务中的命令在执行时可能会修改命令和命令的参数
	// 所以为了正确地传播命令， 需要先备份这些命令和参数
	origArgv := c.Argv
	origArgc := c.Argc
	origCmd := c.Cmd
	// 执行事务中的命令
	for j := 0; j < len(c.Mstate); j++ {
		c.Argc = c.Mstate[j].Argc
		c.Argv = c.Mstate[j].Argv
		c.Cmd = c.Mstate[j].Cmd
		call(c)

		// 因为执行后命令、命令参数可能会被改变
		// 比如SPOP会被该写为SREM
		// 所以这里需要更新事务队列中的命令和参数
		// 确保附属节点和AOF的数据一致性
		c.Mstate[j].Argc = c.Argc
		c.Mstate[j].Argv = c.Argv
		c.Mstate[j].Cmd = c.Cmd
	}

	// 还原命令、命令参数
	c.Argv = origArgv
	c.Argc = origArgc
	c.Cmd = origCmd

	// 清理事务状态
	discardTransaction(c)

}

func discardCommand(c *redisClient) {
	// 不能在客户端未进行事务状态之前使用
	if c.Flags & consts.RedisMulti == 0 {
		c.addReplyError("DISCRAD without MULTI")
		return
	}
	discardTransaction(c)
	c.addReplyOK()
}

func discardTransaction(c *redisClient) {
	c.Mstate = nil
	// 屏蔽事务状态
	c.Flags &= ^(consts.RedisMulti | consts.RedisDitryExec | consts.RedisDirtyCas)
	// 取消对所有键对监视
	unwatchAllKeys(c)
}


func watchCommand(c *redisClient) {
	// 不能在事务开始后执行
	if c.Flags & consts.RedisMulti != 0 {
		c.addReplyError("WATCH inside MULTI is not allowed")
	}
	for j := 1; j < c.Argc; j++ {

	}
	c.addReplyOK()
}

func unwatchCommand(c *redisClient) {

}

type watchedKey struct {
	key *object.Object // 被监视的键
	db *redisDb // 键所在的数据库
}

// 让客户端c监视给定的键key
func watchForKey(c *redisClient, key *object.Object) {
	for _, wk := range c.WatchedKeys {
		// 检查key是否已经保存在watched_keys链表中
		// 如果是的话，直接返回
		if wk.db == c.Db && object.CompareStringObject(key, wk.key) == 0 {
			return
		}
	}
	// 键不存在于watched_keys， 增加它
	// 以下是一个 key 不存在于字典的例子：
	// before :
	// {
	//  'key-1' : [c1, c2, c3],
	//  'key-2' : [c1, c2],
	// }
	// after c-10086 WATCH key-1 and key-3:
	// {
	//  'key-1' : [c1, c2, c3, c-10086],
	//  'key-2' : [c1, c2],
	//  'key-3' : [c-10086]
	// }


	// 检查key是否存在于数据库的watched_keys字典中
	c.Db.appendWatchClient(key, c)
	wk := &watchedKey{key:key,db: c.Db}
	c.WatchedKeys = append(c.WatchedKeys, wk)
}


// 取消客户端对所有键对监视
// 清除客户端事务状态的任务由调用者执行
func unwatchAllKeys(c *redisClient) {
	// 没有键被监视， 直接返回
	if len(c.WatchedKeys) == 0 {
		return
	}
	for _, wk := range c.WatchedKeys {
		wk.db.deleteWatchClient(wk.key, c)
	}
	c.clearWatchKey()
}