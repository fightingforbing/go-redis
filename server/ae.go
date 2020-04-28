package server

import (
	"ddvideo/go_redis_server/consts"
	"fmt"
	"time"
)


type BeforeSleepFunc func(ae *AeEventLoop)

type AeEventLoop struct {
	// 目前已注册的最大描述符
	MaxFd int
	// 目前以追踪到的最大描述符
	SetSize int
	// 用于生成时间事件id
	TimeEventNextId uint64
	// 最后一次执行时间事件的时间
	LastTime time.Time
	// 已注册的时间事件
	Events []*AeFileEvent
	// 已就绪的文件事件
	Fired  []*AeFiredEvent
	// 时间事件
	TimeEventHead *AeTimeEvent
	// 事件处理器的开关
	Stop bool
	// 多路复有库的私有数据
	ApiData *AeApiState
	// 在处理事件前要执行的函数
	BeforeSleep BeforeSleepFunc
}

func (ae *AeEventLoop) Main() {
	ae.Stop = false
	for !ae.Stop { // 事件处理器的主循环
		if ae.BeforeSleep != nil {
			ae.BeforeSleep(ae)
		}
		ae.ProcessEvents(consts.AeAllEvents)
	}
}

func (ae *AeEventLoop) ProcessEvents(flags consts.AeFlags) {
	if flags.NoneDo() {
		return
	}
	if ae.MaxFd != -1 {
		fmt.Println("-----------------------")
		numevents, _ := ae.ApiData.Wait(nil)
		for j := 0; j < numevents; j++ {
			// 从已就绪数组中获取事件
			fe := ae.Events[ae.Fired[j].Fd]
			mask := ae.Fired[j].Mask
			fd := ae.Fired[j].Fd
			rfired := false
			// 读事件
			if fe.Mask & mask & consts.AeREADABLE != 0 {
				rfired = true
				fe.RFileProc(ae,fd,mask,fe.ClientData)
			}
			// 写事件
			if fe.Mask & mask & consts.AeWRITEABLE != 0 {
				if !rfired  {
					fe.WFileProc(ae,fd,mask,fe.ClientData)
				}
			}
		}
	}
}

func (ae *AeEventLoop) CreateFileEvent( fd, mask int, proc FileProc, data *redisClient) {
	ae.GtSetPizePanic(fd)
	fe := ae.Events[fd]
	ae.ApiData.Add(fd, mask)
	fe.Mask |= mask
	if mask & consts.AeREADABLE != 0 {
		fe.RFileProc = proc
	}
	if mask & consts.AeWRITEABLE != 0 {
		fe.WFileProc = proc
	}
	if fd > ae.MaxFd {
		ae.MaxFd = fd
	}
	fe.ClientData = data
}

func (ae *AeEventLoop) GtSetPizePanic(fd int) {
	if fd >= ae.SetSize {
		panic("")
	}
}

func AeCreateEventLoop(setSize int) (ae *AeEventLoop) {
	ae = &AeEventLoop{}
	ae.Events = make([]*AeFileEvent, 0, setSize)
	ae.Fired =  make([]*AeFiredEvent, 0, setSize)
	ae.SetSize = setSize
	ae.LastTime = time.Now()
	ae.TimeEventNextId = 0
	ae.Stop = false
	ae.MaxFd = -1
	ae.BeforeSleep = DefaultEventBeforeSleep
	// 保证先初始化SetSize
	ae.ApiData = AeApiCreate(ae)

	// 初始化监听事件
	for i := 0; i < setSize; i++ {
		ae.Events = append(ae.Events, &AeFileEvent{Mask: consts.AeNone})
	}

	// 初始化已就绪事件
	for i := 0; i < setSize; i++ {
		ae.Fired = append(ae.Fired, &AeFiredEvent{})
	}
	return ae
}

func (ae *AeEventLoop) DeleteFileEvent( fd, mask int) {
	ae.GtSetPizePanic(fd)

	fe := ae.Events[fd]
	if fe.Mask == consts.AeNone {
		return
	}
	fe.Mask = fe.Mask & (^mask)
	if fd == ae.MaxFd && fe.Mask == consts.AeNone {
		var j int
		for j = ae.MaxFd - 1; j >= 0; j-- {
			if ae.Events[j].Mask != consts.AeNone {
				break
			}
		}
		ae.MaxFd = j
	}
	ae.ApiData.Delete(fd, mask)
}

func (ae *AeEventLoop) SetBeforeSleepProc( beforeSleep BeforeSleepFunc ) {
	ae.BeforeSleep = beforeSleep
}


func activeExpireCycle(cycleType int) {
	if cycleType == consts.ActiveExpireCycleFast {
		// 如果上次函数没有触发timelimit_exit  那么不执行处理
		// 如果距离上次执行未够一定时间，那么不执行处理
		// 运行到这里，说明执行快速处理，记录当前时间
	}
}

// 每次处理事件执行
func DefaultEventBeforeSleep(ae *AeEventLoop) {
	// 清除模式为 CYCLE_SLOW ，这个模式会尽量多清除过期键
	activeExpireCycle(consts.ActiveExpireCycleSlow)
}
