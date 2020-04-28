package server

type FileProc func( eventLoop *AeEventLoop, fd uint64, mask int, c *redisClient)

// 文件事件结构
type AeFileEvent struct {
	// 监听事件类型掩码
	Mask int /* one of AE_(READABLE|WRITABLE) */
	// 读事件处理器
	RFileProc FileProc
	// 写事件处理器
	WFileProc FileProc
	// 私有数据
	ClientData *redisClient
}

// 已就绪文件事件
type AeFiredEvent struct {
	// 已就绪文件描述符
	Fd uint64
	Mask int


}
