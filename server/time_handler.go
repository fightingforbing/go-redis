package server

import "time"

// Redis的时间中断器，每秒调用server.hz次
// 以下是需要异步执行的操作
// 主动清除过期键
// 更新软件watchdo的信息
// 更新统计信息
// 对数据库进行渐增式rehash
// 触发BGSAVE或者AOF重写，并处理之后由BGSAVE和AOF重写引发的子进程停止
// 处理客户端超时
// 复制重连
// 等等
func serverCron(eventLoop *AeEventLoop, fd uint64, mask int, c *redisClient) {
	// update the time cache
	updateCachedTime()

	// 记录服务器执行命令的次数

	// 即使服务器的时间最终比1.5年长也无所谓
	// 对象系统仍会正常运作，不过一些对象可能会比服务器本身的时钟更年轻
	// 不过这要这个对象在1.5年内都没有被访问过，才会出现这中现象
	// LRU时间的精度可以通过修改 REDIS_LRU_CLOCK_RESOLUTION常量来改变

}


func updateCachedTime() {
	Server.UnixTime = time.Now()
}

func getLRUClock() {

}