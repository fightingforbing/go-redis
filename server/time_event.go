package server

type TimeProc func (eventLoop *AeEventLoop, fd uint64, mask int)

// 时间事件结构
type AeTimeEvent struct {

}