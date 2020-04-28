package server

import (
	"ddvideo/go_redis_server/consts"
	"ddvideo/go_redis_server/server2/object"
	"fmt"
	"github.com/tidwall/redcon"
	"log"
	"syscall"
)

func AcceptTcp( eventLoop *AeEventLoop, fd uint64, mask int, c *redisClient) {
	var (
		nfd int
		err error
	)
	for {
		nfd, _,  err = syscall.Accept(int(fd))
		if nfd == -1 {
			if err == syscall.EAGAIN {
				continue
			}
			panic(err)
		}
		break
	}
	acceptCommondHandler(nfd)
}

func acceptCommondHandler(fd int) {
	NewRedisClient(fd)
	// 这里增加如果服务端的连接大于连接数
	// 则报错的保护
}

// 处理请求
func ReadQueryFromClient( eventLoop *AeEventLoop, fd uint64, mask int, c *redisClient) {
	Server.CurrentClients = c
	nread, err := syscall.Read(c.Fd, c.QueryBuf)
	if err != nil { // 读入出错
		if err == syscall.EAGAIN {
			nread = 0
		} else {
			log.Printf("reding from client err: %s", err.Error())
			Server.FreeClient(c)
			return
		}
	} else if (nread == 0) {// 读到EOF 客户端关闭
		log.Printf("Client closed connection")
		Server.FreeClient(c)
		return
	}

	if nread == 0 {
		Server.CurrentClients = nil
		return
	}
	processInputBuffer(c, c.QueryBuf[:nread])
	Server.CurrentClients = nil
}

// 处理客户端输入的命令内容
func processInputBuffer(c *redisClient, in []byte)  {
	var (
		complete bool
		err error
		argv [][]byte
		leftover []byte
	)
	data := c.ProcessBuf.Begin(in)
	defer func(){
		c.ProcessBuf.End(leftover)
	}()
	fmt.Printf("读取到到内容%q\n",in)
	complete, argv, _, leftover, err = redcon.ReadNextCommand(data, argv)
	if err != nil {
		c.addReplyError(err.Error())
		return
	}
	if !complete {
		return
	}
	if len(argv) == 0 {
		return
	}
	c.Argc = len(argv)
	if cap(c.Argv) > 10 {
		c.Argv = nil
	}
	c.Argv = c.Argv[:0]
	for index, arg := range argv {
		argStr := string(arg)
		if index == 0 { // key
			c.Argv = append(c.Argv, object.CreateStringObject(argStr))
			continue
		}
		c.Argv = append(c.Argv, object.CreateObject(argStr))
	}
	processCommand(c)
}


// 负责传送命令回复的写处理器
func SendReplyToClient(eventLoop *AeEventLoop, fd uint64, mask int, c *redisClient){
	// 开始写事件到客户端
	for len(c.ReplyBuf) > 0 {
		nwrite, err := syscall.Write(int(fd),c.ReplyBuf)
		if err != nil {
			if err == syscall.EAGAIN {
				continue
			} else {
				Server.FreeClient(c)
				return
			}
		}
		c.ReplyBuf = c.ReplyBuf[nwrite:]
	}
	c.freeReplyBuf()
	eventLoop.DeleteFileEvent(int(fd), consts.AeWRITEABLE)
	if ( c.Flags & consts.RedisCloseAfterReply) != 0 {
		Server.FreeClient(c)
	}
}