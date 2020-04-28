package server

// 执行AOF文件中的命令
// 出错时返回true
// 出现非执行错误(比如文件长度为0)时返回false
// 出现致命错误时打印信息到日志，并且程序退出
func loadAppendOnlyFile(filename string) {
	var (
		fakeClient *redisClient
	)
	
}