package server

import "time"

// 从磁盘加载数据
func loadDataFromDist() {
	// 记录开始时间
	start := time.Now()
	// 优先AOF持久化
	if Server.AofState {

	} else {

	}
}
