package server

import (
	"ddvideo/go_redis_server/consts"
	"ddvideo/go_redis_server/server/object"
	"container/list"
	"time"
)

var Server *server

type server struct {
	ConfigFile string         					// 配置文件路径
	Port       int          					// 默认端口
	PidFile    string        					// 进程id 文件
	MaxClients  int            					// 最大能连接客户端
	Ipfd       []int           					// 描述符数量
	AllClients    map[int]*redisClient  		// 保存了所有客户端状态结构
	CloseClients []*redisClient 				// 保存了所有待关闭的客户端
	CurrentClients *redisClient 				// 当前客户端
	EventLoop         *AeEventLoop  			// 事件循环
	DbNum         int                           // 数据库数量
	Db            []*redisDb                    // 数据库
	UnixTime      time.Time                     // 缓存时间
	// 字典，键是频道，值为链表
	// 链表中保存链所有订阅某个频道的客户端
	// 新客户端总是被增加到链表到表尾
	PubsubChannels map[object.Object]list.List

	Commands map[string]*redisCommand 			// redis的命令集合
	BpopBlockedClients uint64                   // 阻塞状态的客户端
	StatNumCommands uint64   					// 已处理命令的数量
	StatNumConnections uint64 					// 服务器接到的连接请求数量
	StatExpiredKeys  uint64 					// 已过期的键数量
	StatEvictedKeys uint64 						// 回收内存而被释放的过期键
	StatKeySpaceHits uint64  					// 成功查找键的次数
	StatKeySpaceMiss uint64 					// 查找键失败的次数
	StatRejectedConn uint64 					// 服务器因为客户端数量过多而拒绝客户端连接的次数
	StatSyncFull uint64 						// 执行 full sync的次数

	// 持久化
	AofState bool								// 持久化开关
	RdbFilename string							// RDB文件路径
	Loading  bool                               // 值为真时，表示服务器正在进行载入
	LoadingTotalBytes int64                    // 正在载入的数据的大小
	LoadingLoadedBytes int64                   // 已载入数据的大小
	LoadingStartTime time.Time                  // 开始进行载入的时间
}


func (s *server) FreeClient(c *redisClient) {
	c.Free()
	if Server.CurrentClients == c {
		Server.CurrentClients = nil
	}
	delete(s.AllClients, c.Fd)

}

func InitServerConfig() {
	Server = &server{}
	Server.Port = consts.RedisServerPort
	Server.PidFile = consts.RedisPidFile
	Server.MaxClients = consts.RedisMaxClients
	Server.Ipfd = make([]int, 0, 16)
	Server.Commands = populateCommandTable()
	Server.DbNum = consts.RedisDbNum

	// 初始化数据库
	for i := 0; i < Server.DbNum; i++ {
		Server.Db = append(Server.Db, createRedisDb())
	}
}
func Run() {
	Server.EventLoop = AeCreateEventLoop(Server.MaxClients + consts.RedisEventLoopFdSetIncr)

	// 打开TCP监听端口，用于等待客户端的命令请求
	listenToPort(Server.Port)

	// 关联连接应答处理器
	for _, fd := range Server.Ipfd {
		Server.EventLoop.CreateFileEvent(fd, consts.AeREADABLE, AcceptTcp, nil)
	}

	// 开始事件循环
	Server.EventLoop.Main()
}

func selectDb(id int) *redisDb {
	// 确保id在正确范围内
	if id < 0 || id >= Server.DbNum {
		return nil
	}
	return Server.Db[id]
}