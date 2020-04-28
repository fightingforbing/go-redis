package server

import (
	"ddvideo/go_redis_server/consts"
	"ddvideo/go_redis_server/server/object"
	"fmt"
	"time"
)


type commandProc func(*redisClient)

type redisCommand struct {
	Name string // 命令名字
	Proc commandProc // 实现函数
	Arity int // 参数个数
}
var (
	redisCommandTable = []redisCommand{
		{"get", getCommand, 2},
		{"set", setCommand, -3},
		{"ttl", ttlCommand, 2},
		{"del", delCommand, -2},
		{"exists", existCommand, 2},
		{"append", appendCommand, 3},
		{"expire", expireCommand, 3},
		{"expireat", expireatCommand, 3},
		{"pexpire", pexpireCommand, 3},
		{"pexpireat", pexpireatCommand, 3},
		{"incr", incrCommand, 2},
		{"decr", decrCommand, 2},
		{"incrby", incrByCommand, 3},
		{"decrby", decrByCommand, 3},
		{"scan", scanCommand, -2},
		// 列表
		{"lpush", lpushCommand, -3},
		{"rpush", rpushCommand, -3},
		{"lpop", lpopCommand, 2},
		{"rpop", rpopCommmand, 2},
		{"blpop", blpopCommand, -3},
		{"brpop", brpopCommand, -3},
		{"lindex", lindexCommand, 3},
		{"hset", hsetCommand, 4},
		// 集合
		{"sadd", saddCommand, -3},
		{"sismember", sismemberCommand, 3},
		{"srem", sremCommand, -3},
		// 有序集合
		{"zadd", zaddCommand,-4},
		{"zrem", zremCommand, -3},
		{"zrange", zrangCommand, -4},
		{"zrevrange", zrevrangeCommand, -4},
		// 事务
		{"multi", multiCommand, 1},
		{"exec", execCommand, 1},
		{"watch", watchCommand, -2},
		{"unwatch", unwatchCommand, 1},
		{"discard", discardCommand,1},
		// 订阅发布
		{"subscribe", subscribeCommand, -2},
		{"publish", publishCommand, 3},
		{"unsubscribe", unsubscribeCommand, -1}
	}
)

func populateCommandTable() map[string]*redisCommand {
	result := make(map[string]*redisCommand)
	for i := 0; i < len(redisCommandTable); i++ {
		command := redisCommandTable[i]
		result[command.Name] = &command
	}
	return result
}

func processCommand(c *redisClient) {
	if StrcmpObjectAndStringIngoreCase(c.Argv[0], "quit") { // 客户端退出
		c.addReplyOK()
		c.Flags |= consts.RedisCloseAfterReply
		return
	}
	argv0 := object.ConvertStringTypeObjectToString(c.Argv[0])
	cmd, ok := Server.Commands[argv0]
	if !ok {
		msg := fmt.Sprintf("unknown command %s", argv0)
		c.addReplyError(msg)
		return
	}
	c.Cmd = cmd
	c.LastCmd = cmd
	// 检查参数数量
	if ( cmd.Arity > 0 && cmd.Arity != c.Argc ) ||
		( c.Argc < -cmd.Arity ) {
		c.addReplyError(fmt.Sprintf("wrong number of arguments for '%s' command", argv0))
	}

	// exec the command
	if c.Flags & consts.RedisMulti != 0 && c.Cmd.Name != "exec" &&
		c.Cmd.Name != "discard" && c.Cmd.Name != "multi" && c.Cmd.Name != "watch"{
		// 在事务上下文中
		// 除EXEC、DISCARD、MULTI和WATCH命令外
		// 其他所有命令都会被入队到事务队列中

	}
	// 调用
	call(c)
}

func call(c *redisClient) {
	c.Cmd.Proc(c)
}

func incrCommand(c *redisClient) {
	fmt.Println("执行incr命令")
	incrDecrCommand(c, 1)
}

func decrCommand(c *redisClient) {
	fmt.Println("执行decr命令")
	incrDecrCommand(c, -1)
}

func incrByCommand(c *redisClient) {
	incr ,ok := c.Argv[2].Data.(int64)
	if !ok {
		c.addReplyTypeError()
		return
	}
	incrDecrCommand(c, incr)
}

/*
	阻塞POP操作的运作方法， 以BLPOP作为例子

	-如果用户调用BLPOP, 并且列表非空， 那么程序执行LPOP
	因此， 当列表非空时， 调用BLPOP等于调用LPOP

	-当BLPOP对一个空键执行时，客户端才会被阻塞:
	服务端不再对这个客户端发送任何数据，
	对这个客户端的状态设为"被阻塞"，直到解除阻塞为止。
	并且客户端会被加入到一个以阻塞键为key
	以被阻塞客户端为value的字典 db.BlockingKeys中

	-当有PUSH命令作用于一个造成客户端阻塞的键时，
	程序将这个键标记为"就绪"，并且在执行这个命令、事务、或脚本之后，
	程序会按"先阻塞先服务"的顺序(链表)，将列表的元素返回给那么被阻塞的客户端
	被解除阻塞的客户端数量取决于PUSH命令推入的元素数量
 */
func blockForKeys(c *redisClient, blockKeys []*object.Object, timeout time.Duration) {
	c.Bpop.Timeout = timeout
	for _, o := range blockKeys { // 所有阻塞键
		c.Bpop.Keys[*o] = true
		// 将客户端填接到被阻塞客户端到链表中
		de, ok := c.Db.BlockIngKeys[o.Data.(string)]
		if !ok {
			c.Db.BlockIngKeys[o.Data.(string)] = make([]*redisClient, 1)
			c.Db.BlockIngKeys[o.Data.(string)][0] = c
		} else {
			c.Db.BlockIngKeys[o.Data.(string)] = append(de, c)
		}
	}
	c.block(consts.RedisBlockedList)
}




func hsetCommand(c *redisClient) {
}

func decrByCommand(c *redisClient) {
	if c.Argv[2].Type != consts.RedisEncodingInt {
		c.addReplyTypeError()
	}
	incr := c.Argv[2].Data.(int64)
	incrDecrCommand(c, incr)
}

func incrDecrCommand(c *redisClient, incr int64) {
	obj := c.lookupKey(c.Argv[1])
	if obj != nil && c.checkTypeOrReply(obj,consts.RedisString) {
		return
	}
	if obj.Type != consts.RedisEncodingInt {
		c.addReplyTypeError()
		return
	}
	dataInteger := obj.Data.(int64)
	if ( incr < 0 && dataInteger < 0 && incr < (consts.RangeInt64Min - dataInteger) ) ||
		( incr > 0 && dataInteger > 0 && incr > (consts.RangeInt64Min - dataInteger)) {
		c.addReplyError("increment or decrement would overflow")
		return
	}
	dataInteger += incr
	c.addReplyInt64(dataInteger)
}

func getCommand(c *redisClient) {
	fmt.Println("执行get命令")
	obj := c.lookupKeyRead(c.Argv[1])
	if  obj == nil {
		c.addReplyNull()
		return
	}
	if c.checkTypeOrReply(obj, consts.RedisString) {
		return
	}
	c.addReply(obj)
}

func appendCommand(c *redisClient) {
	fmt.Println("执行append命令")
	var (
		curStr string
		appendStr string
	)
	curObj := c.lookupKey(c.Argv[1])
	appendStr = object.ConvertStringTypeObjectToString(c.Argv[2])
	if curObj != nil {
		if c.checkTypeOrReply(curObj, consts.RedisString) {
			return
		}
		curStr = object.ConvertStringTypeObjectToString(curObj)
	}
	catStr := curStr + appendStr
	c.setKey(c.Argv[1], object.CreateObject(catStr))
	c.addReplyInt64(int64(len(catStr)))
	return
}

func setCommand(c *redisClient) {
	fmt.Println("执行set命令")
	var (
		flag  = consts.RedisSetNoFlag
		unit  = consts.RedisUnitSeconds
		next *object.Object
		expire *object.Object
	)
	for j := 3; j < c.Argc; j++ {
		next = nil
		if j < c.Argc - 1 {
			next = c.Argv[j+1]
		}
		obj := c.Argv[j]
		if StrcmpObjectAndStringIngoreCase(obj, "nx") {
			flag = consts.RedisSetNx
		} else if StrcmpObjectAndStringIngoreCase(obj, "xx") {
			flag = consts.RedisSetXx
		} else if StrcmpObjectAndStringIngoreCase(obj, "ex") {
			unit = consts.RedisUnitSeconds
			expire = next
		} else if StrcmpObjectAndStringIngoreCase(obj, "px") {
			unit = consts.RedisUnitMillisencods
			expire = next
		} else {
			c.addReplySyntaxError()
		}
	}
	if expire != nil && expire.Type != consts.RedisEncodingInt {
		c.addReplyTypeError()
		return
	}
	setGenericCommand(c, flag, c.Argv[1], c.Argv[2], expire, unit)
}

func setGenericCommand(c *redisClient, flags int,  key, val, expire *object.Object, unit int) {
	var (
		millisecond int64
	)
	if expire != nil { // 存在过期设置
		millisecond = expire.Data.(int64)
		if millisecond < 0 {
			c.addReplyError("invalid expire time in SETEX")
			return
		}
		if unit == consts.RedisUnitSeconds { // 转化成毫秒
			millisecond = millisecond * 1000
		}
	}
	if ( flags == consts.RedisSetNx  && c.lookupKey(key) != nil) ||
		( flags == consts.RedisSetXx && c.lookupKey(key) == nil ) {
		c.addReplyNull()
		return
	}
	// 存储数据
	c.setKey(key, val)
	// 设置过期时间
	fmt.Printf("设置过期时间%d\n", millisecond)
	if millisecond > 0 {
		c.setExpire(key, millisecond)
	}
	c.addReplyOK()
}

func ttlCommand(c *redisClient) {
	fmt.Println("执行ttl命令")
	ttlGenericCommand(c, consts.RedisOutputSeconds)
}

func ttlGenericCommand(c *redisClient, output int) {
	var ttlD time.Duration
	if obj := c.lookupKeyRead(c.Argv[1]); obj == nil {
		c.addReplyInt64(-2)
		return
	}

	ttlD, isPerpet := c.ttl(c.Argv[1])
	if isPerpet { // 未设置过期时间
		c.addReplyInt64(-1)
		return
	}
	ttl := ttlD.Nanoseconds() / 1000000 // 毫秒
	if output == consts.RedisOutputSeconds { // 秒
		ttl = int64(ttlD.Seconds())
	}
	c.addReplyInt64(ttl)
}

func delCommand(c *redisClient) {
	var delete int64
	for j := 1; j < c.Argc; j++ {
		if c.deleteKey(c.Argv[j]) {
			// 删除键成功 发送通知
			delete++
		}
	}
	c.addReplyInt64(delete)
}

func existCommand(c *redisClient) {
	var result int64
	if c.existKey(c.Argv[1]) {
		result = 1
	}
	c.addReplyInt64(result)
}

func expireCommand(c *redisClient) {
	expireGenericCommand(c, time.Now().Unix() * 1000, consts.RedisUnitSeconds)
}

func expireatCommand(c *redisClient) {
	expireGenericCommand(c, 0, consts.RedisUnitSeconds)
}

func pexpireCommand(c *redisClient) {
	expireGenericCommand(c, time.Now().Unix() * 1000, consts.RedisUnitMillisencods)
}

func pexpireatCommand(c *redisClient) {
	expireGenericCommand(c, 0, consts.RedisUnitMillisencods)
}

/*
 这个函数是EXPIRE、PEXPIRE、EXPIREAT 和PEXPIREAT命令的底层实现函数
 命令的第二个参数可能是绝对值，也可能是相对值

 当执行 *AT 命令时， basttime为0， 在其他情况下， 它保存的是当前的绝对时间
 uint 用于执行argv[2] （传入过期时间）的格式
 它可以是UNIT_SECONDS 或UNIT_MILLISECONDS

 basetime 参数则总是毫秒格式的
 */
func expireGenericCommand(c *redisClient, basetime int64, unit int) {
	key := c.Argv[1]
	param := c.Argv[2]
	if param.Type != consts.RedisInt {
		c.addReplyError("value is not an integer or out of range")
		return
	}
	when := param.Data.(int64)
	if unit == consts.RedisUnitSeconds {
		when *= 1000
	}
	when += basetime
	if c.lookupKeyRead(key) == nil {
		c.addReplyInt64(0)
		return
	}
	c.setExpire(key, when)
}

func scanCommand(c *redisClient) {
	cursor := c.Argv[1].Data.(int64)

}

// 这是scan， hscan， sscan命令的实现函数
// 如果给定了对象o， 那么它必须是一个哈希对象或者集合对象
// 如果o为null的话， 函数将使用当前数据库作为迭代对象
// 如果参数o不为nil, 那么说明它是一个键对象， 函数将跳过这些键对象
// 对给定的命名选项进行分析

// 如果被迭代的是哈希对象，那么函数返回的是键值对
func scanGenericCommand(c *redisClient,o *object.Object, cursor int64) {
	// 设置第一个选项参数的索引位置
	// 0 1 2 3
	// SCAN OPTION <op_arg> 命令的选项值从索引2开始
	// HSCAN <>HSCAN 而其他*SCAN命令的选项值从索引3开始
	i := 3
	if o == nil {
		i = 2
	}
	// 分析选项参数
	for i < c.Argc {
		j := c.Argc - i
		// COUNT <number>
		if StrcmpObjectAndStringIngoreCase(c.Argv[i],"count") && j >= 2 {
			if c.Argv[i+1].Type != consts.RedisInt {
				c.addReplyError("value is out of range")
				return
			}
			count := c.Argv[i+1].Data.(int64)
			if count < 1 {
				c.addReplySyntaxError()
				return
			}
			i += 2
		} else {
			c.addReplySyntaxError()
			return
		}
	}
	// 如果这些对象的底层实现为ziplist、intset 而不是哈希表
	// 那么这些对象应该只包含了少量元素
	// 为了保持不让服务器记录迭代状态的设计
	// 我们将ziplist或者intset里面的所有元素有一次返回给调用者
	// 并向调用者返回游标 0
}