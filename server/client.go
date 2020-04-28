package server

import (
	"ddvideo/go_redis_server/consts"
	"ddvideo/go_redis_server/server/object"
	"errors"
	"fmt"
	"github.com/tidwall/evio"
	"github.com/tidwall/redcon"
	"syscall"
	"time"
)


type redisClient struct {
	Fd int 							// 套接字描述符
	Name string 					// 客户端的名字
	Argc int  						// 参数数量
	Argv []*object.Object 			// 参数内容
	Cmd  *redisCommand
	LastCmd *redisCommand
	QueryBuf []byte 				// 客户端发送数据
	ProcessBuf evio.InputStream  	// 还需要处理的数据
	Flags  int
	ReqType int 					// 请求类型
	Db *redisDb 					// 当前正在使用的数据库
	Bpop *blockingState             // 阻塞状态
	BType int                       // 阻塞类型
	Mstate []*multiState              // 事务
	// redis的回复缓存区，是一个16M的固定缓冲区，和一个回复链表
	// 每个回复链表有一个16M的缓冲区，
	// 当前缓冲区加满， 在创建下一个缓冲区
	ReplyBuf []byte

	// 记录了客户端所有订阅的频道
	PubsubChannels map[object.Object]interface{}
	WatchedKeys []*watchedKey 		// 被监视的键
}

type multiState struct {
	Argv []*object.Object
	Argc int
	Cmd *redisCommand
}

// 阻塞状态
type blockingState struct {
	Timeout time.Duration // 阻塞时限
	Keys map[object.Object]bool // 造成阻塞到键
}

func (c *redisClient) block(t int) {
	c.Flags |= consts.RedisBlocked
	c.BType = t
	Server.BpopBlockedClients++
}

func (c *redisClient) lookupKey(key *object.Object) *object.Object {
	return c.Db.lookupKey(key)
}

func (c *redisClient) clearWatchKey() {
	c.WatchedKeys = c.WatchedKeys[:0]
}

func (c *redisClient) lookupKeyRead(key *object.Object) *object.Object {
	c.Db.expireIfNeed(key)
	obj := c.lookupKey(key)

	// 更新命中/不命中信息
	if obj == nil {
		Server.StatKeySpaceMiss++
	} else {
		Server.StatKeySpaceHits++
	}
	return obj
}

func (c *redisClient) lookupKeyWrite( key *object.Object) *object.Object {
	c.Db.expireIfNeed(key)
	return c.lookupKey(key)
}

func (c *redisClient) setKey(key, val *object.Object) {
	if c.lookupKey(key) == nil {
		c.Db.addVal(key, val)
	} else {
		c.Db.overWriteVal(key, val)
	}
}


func (c *redisClient) addKeyValue(key, val *object.Object) {
	c.Db.addVal(key, val)
}

func (c *redisClient) ttl(key *object.Object) (d time.Duration, isPerpet bool) {
	return c.Db.ttl(key)
}

func (c *redisClient) setExpire( key  *object.Object, milliseconds int64) {
	v := c.lookupKey(key)
	if v != nil {
		c.Db.setExpire(key, time.Millisecond * time.Duration(milliseconds))
	}
}

func (c *redisClient) existKey(key *object.Object) bool {
	c.Db.expireIfNeed(key)
	if c.lookupKey(key) == nil {
		return false
	}
	return true
}

// true 删除成功
// false 键不存在
func (c *redisClient) deleteKey( key *object.Object) bool {
	if c.lookupKey(key) == nil { // 不存在
		return false
	}

	if c.Db.expireIfNeed(key) { // 已过期
		return false
	}
	c.Db.deleteKey(key)
	return true
}

func (c *redisClient) freeReplyBuf() {
	if cap(c.ReplyBuf) > consts.RedisIoBufLen {
		c.ReplyBuf = make([]byte, 0, consts.RedisIoBufLen)
	}
	c.ReplyBuf = c.ReplyBuf[:0]
}

func (c *redisClient) addReplyOK() {
	replyWarp(c, "", func(c *redisClient, s string) {
		c.ReplyBuf = redcon.AppendOK(c.ReplyBuf)
	})
}

func (c *redisClient) addReplyNull() {
	c.prepareToWrite()
	c.ReplyBuf = redcon.AppendNull(c.ReplyBuf)
}

func (c *redisClient) addReplyNulMultilBulk() {
	c.prepareToWrite()
	c.ReplyBuf = redcon.AppendArray(c.ReplyBuf, -1)
}

func (c *redisClient) addReplyArray(objs []*object.Object) {
	c.prepareToWrite()
	c.ReplyBuf = redcon.AppendArray(c.ReplyBuf, len(objs))
	for _, o := range objs {
		c.addReplyBulk(o)
	}
}

func (c *redisClient) addReplyBulk( o *object.Object) {
	c.prepareToWrite()
	switch v := o.Data.(type) {
	case string:
		c.ReplyBuf = redcon.AppendBulkString(c.ReplyBuf, v)
	case int64:
		c.ReplyBuf = redcon.AppendBulkInt(c.ReplyBuf, v)
	}
}


func (c *redisClient) addReply( obj *object.Object ) {
	if err := c.prepareToWrite(); err != nil {
		return
	}
	s := object.ConvertStringTypeObjectToString(obj)
	c.ReplyBuf = redcon.AppendString(c.ReplyBuf, s)
}

func (c *redisClient) addReplySyntaxError() {
	c.prepareToWrite()
	c.ReplyBuf = redcon.AppendString(c.ReplyBuf,"ERR syntax error")
}

func (c *redisClient) addReplyTypeError() {
	c.prepareToWrite()
	c.ReplyBuf = redcon.AppendError(c.ReplyBuf, "WRONGTYPE Operation against a key holding the wrong kind of value")
}

func (c *redisClient) addReplyError(s string) {
	replyWarp(c,s, func(c *redisClient, s string) {
		c.ReplyBuf = redcon.AppendError(c.ReplyBuf, s)
	})
}

func (c *redisClient) addReplyTimeoutError() {
	c.addReply("timeout is negative")
}


func (c *redisClient) checkTypeOrReply(o *object.Object, t int) bool {
	if o.Type() != t {
		c.addReplyError("WRONGTYPE Operation against a key holding the wrong kind of value")
		return true
	}
	return false
}


func (c *redisClient) addReplyInt64 (v int64) {
	c.prepareToWrite()
	c.ReplyBuf = redcon.AppendInt(c.ReplyBuf, v)
}

func replyWarp(c *redisClient, s string, f func(*redisClient, string)) {
	c.prepareToWrite()
	f(c,s)
}

// 保证还没reply的时候 设置读事件
func (c *redisClient) prepareToWrite() error {
	if len(c.ReplyBuf) > 0 {
		return errors.New("已经有回复信息")
	}
	Server.EventLoop.CreateFileEvent(c.Fd,consts.AeWRITEABLE,SendReplyToClient,c)
	return nil
}

func (c *redisClient) Reset() {

}

func (c *redisClient) Free() {

	if c.Fd > -1 {
		Server.EventLoop.DeleteFileEvent(c.Fd, consts.AeREADABLE)
		Server.EventLoop.DeleteFileEvent(c.Fd, consts.AeREADABLE)
		syscall.Close(c.Fd)
	}
}

func NewRedisClient(fd int) {
	// 设置非阻塞
	if err := syscall.SetNonblock(fd, true); err != nil {
		fmt.Printf("%d 设置非阻塞失败 %v\n",fd, err)
	}

	// 禁用Nagle
	if err := syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1); err != nil {
		fmt.Printf("%d 设置 no delay 失败 %v\n", fd, err)
	}
	// 设置keepalive
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1); err != nil {
		fmt.Printf("%d 设置 keep alive 失败%v\n", fd, err)
	}
	c := &redisClient{}
	// 绑定读事件
	Server.EventLoop.CreateFileEvent(fd, consts.AeREADABLE, ReadQueryFromClient, c)

	c.Fd = fd
	c.Name = ""
	c.Db = selectDb(0)
	c.Argv = make([]*object.Object,0,10)
	c.QueryBuf = make([]byte, consts.RedisIoBufLen)
	c.ReplyBuf = make([]byte,0, consts.RedisIoBufLen)
	c.Bpop = &blockingState{}
}
