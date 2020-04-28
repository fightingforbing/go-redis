package server

import (
	"ddvideo/go_redis_server/consts"
	"fmt"
	"syscall"
)

/*
struct Kevent_t {
	uinptr_t ident;  	// 事件ID
	short filter;   	// 事件过滤器
	u_short flags;  	// 行为标识
	u_int fflags;       // 过滤器标识值
	intptr_t data;      // 过滤器数据
	void *udata;		// 应用透传数据
}
在一个kevent_t数据中 {ident, filter}确定一个唯一的事件
*/

type AeApiState struct {
	KqFd int
	EventLoop *AeEventLoop
}

func (api *AeApiState) Wait( timeout *syscall.Timespec) (numEvents int,err error) {
	events := make([]syscall.Kevent_t, api.EventLoop.SetSize)
	n, err := syscall.Kevent(api.KqFd, nil, events, nil)
	if err != nil && err != syscall.EINTR {
		return 0, err
	}
	for i := 0; i < n; i++ {
		mask := 0
		e := events[i]
		if e.Filter == syscall.EVFILT_READ {
			mask |= consts.AeREADABLE
		}
		if e.Filter == syscall.EVFILT_WRITE {
			mask |= consts.AeWRITEABLE
		}
		api.EventLoop.Fired[i].Fd = e.Ident
		api.EventLoop.Fired[i].Mask = mask
	}
	return n,nil
}

func (api *AeApiState) Add(fd, mask int) error {
	if mask & consts.AeREADABLE != 0 {
		api.addRead(fd)
	}
	if mask & consts.AeWRITEABLE != 0 {
		api.addWrite(fd)
	}
	return nil
}

func (api *AeApiState) addWrite(fd int) (err error) {
	ke := syscall.Kevent_t{}
	evSET(&ke, uint64(fd), syscall.EVFILT_WRITE, syscall.EV_ADD, 0,0, nil)
	_ , err = syscall.Kevent(api.KqFd, []syscall.Kevent_t{ke},nil,nil)
	if err != nil {
		fmt.Printf("write%v\n", err)
	}
	return err
}

func (api *AeApiState) addRead(fd int) (err error) {
	ke := syscall.Kevent_t{}
	evSET(&ke, uint64(fd), syscall.EVFILT_READ, syscall.EV_ADD, 0,0, nil)
	_ , err = syscall.Kevent(api.KqFd, []syscall.Kevent_t{ke},nil,nil)
	if err != nil {
		fmt.Printf("增加read事件失败%v\n", err)
	}
	return
}

func (api *AeApiState) Delete(fd, mask int) {
	ke := syscall.Kevent_t{}
	if mask & consts.AeREADABLE  != 0 {
		evSET(&ke, uint64(fd), syscall.EVFILT_READ, syscall.EV_DELETE, 0,0, nil)
		syscall.Kevent(api.KqFd, []syscall.Kevent_t{ke},nil,nil)
	}
	if mask & consts.AeWRITEABLE != 0 {
		evSET(&ke, uint64(fd), syscall.EVFILT_WRITE, syscall.EV_DELETE,0,0,nil)
		syscall.Kevent(api.KqFd, []syscall.Kevent_t{ke}, nil, nil)
	}
}

func AeApiCreate(aeEventLoop *AeEventLoop) (api *AeApiState) {
	var err error
	api = &AeApiState{}
	api.KqFd, err = syscall.Kqueue()
	api.EventLoop = aeEventLoop
	if err != nil {
		panic(err)
	}
	return api
}


func evSET(ke *syscall.Kevent_t,ident uint64, filter int16, flags uint16, fflags uint32, data int64, udata *byte ) {
	ke.Ident = ident
	ke.Filter = filter
	ke.Flags = flags
	ke.Fflags = fflags
	ke.Data = data
	ke.Udata = udata
}
