package consts

// 过期循环
const (
	ActiveExpireCycleSlow = 0	// 时间充裕的用slow
	ActiveExpireCycleFast = 1	// 时间不够的用fast
)

// 事件类型掩码
const (
	AeNone = 0
	AeREADABLE = 1
	AeWRITEABLE = 2
)


// 事件处理器 flags
type AeFlags int
func (f AeFlags) NoneDo() bool {
	return (f & AeTimeEvents == 0) && (f & AeFileEvents == 0)
}

func (f AeFlags) NoBlockTimeEvnet() bool {
	return (f & AeTimeEvents != 0) && (f & AeDontWait == 0)
}

const (
	AeNoMore AeFlags =  -1
	// 文件事件
	AeFileEvents AeFlags = 1
	// 时间事件
	AeTimeEvents AeFlags = 2
	// 所有事件
	AeAllEvents AeFlags = AeFileEvents | AeTimeEvents
	// 不阻塞， 也不进行等待
	AeDontWait AeFlags = 4
)

// 请求类型
const (
	RedisReqInline = 1
	RedisReqMultiBulk = 2
)

// Object type
// 对象类型
const (
	RedisString = 0 // 字符串
	RedisList = 1	// 链表
	RedisSet = 2	// 集合
	RedisZset = 3	// 有序集合
	RedisHash = 4	// 哈希
	RedisInt = 5    // 整数
	RedisFloat = 6  // 浮点数
)

const (
	RedisHead = 0
	RedisTail = 1
)

// 对象编码
const (
	RedisEncodingRaw = 0 // RedisString 动态字符串
	RedisEncodingInt = 1 // RedisString 使用整数值实现的字符串对象
	RedisEncodingHt = 2 // RedisHash 使用字典实现的哈希对象
	RedisEncodingZipmap = 3 //
	RedisEncodingLinkedList = 4 // RedisList 使用双端链表实现的列表对象
	RedisEncodingZipList = 5  // RedisList 使用压缩列表实现的列表对象
	RedisEncodingIntSet = 6 // 使用整数集合实现的集合对象
	RedisEncodingSkipList = 7 // 使用跳跃表和字典实现的有序集合对象
)

// 命令类型
const(
	// set
	RedisSetNoFlag = 0
	RedisSetNx = 1
	RedisSetXx = 2
	RedisUnitSeconds = 3
	RedisUnitMillisencods = 4
	// ttl
	RedisOutputSeconds = 1
	RedisOutputMilliseconds = 0
	// ziplist
	RedisZipPushHead = 0
	RedisZipPushTail = 1
)

// redisClient flags
const (
	RedisMulti = 1 << 3 // this client is in a multi context
	RedisBlocked = 1 << 4 // this client is waiting in a blocking operation
	RedisDirtyCas = 1 << 5 // Watched keys modified. EXEC will fail.
	RedisCloseAfterReply = 1 << 6 // close after writing entrie reply
	RedisUnBlocked = 1 << 7 // this client was unblocked and is stored in server.unblocled_clients
	RedisDitryExec = 1 << 12 // Exec will failed for error with queueing
)

// client block type
const (
	RedisBlockedNone = 0 	// not blocked, no RedisBlocked flag set
	RedisBlockedList = 1	// blpop
	RedisBockedWait = 2		// wait for synchronous replication
)

// int类型范围
const (
	RangeInt64Max int64 = 1 << 63 - 1
	RangeInt64Min int64 = 0 - (1 << 63)

	RangeInt8Max = 1 << 7 -1
	RangeInt8Min = 0 - (1 << 7)

	RangeInt16Max = 1 << 15 - 1
	RangeInt16Min = 0 - (1 << 15 )

	RangeInt24Max = 1 << 24 - 1
	RangeInt24Min = 0 - (1 << 24)

	RangeInt32Max = 1 << 32 - 1
	RangeInt32Min = 0 - (1 << 32)
)

// RDB
const (
	RedisRdbVersion = 6
	// 对象类型在RDB文件中的类型
	RedisRdbTypeString = 0
	RedisRdbTypeList = 1
	RedisRdbTypeSet = 2
	RedisRdbTypeZset = 3
	RedisRdbTypeHash = 4

	// 对象的编码方式
	RedisRdbTypeHashZipmap = 9
	RedisRdbTypeListZiplist = 10
	RedisRdbTypeSetIntset = 11
	RedisRdbTypeZsetZiplist = 12
	RedisRdbTypeHashZiplist = 13

	// 数据库特殊操作标识符
	RedisRdbOpcodeExpiretimeEs = 252 // 以MS算的过期时间
	RedisRdbOpcodeExpireMe = 253     // 以秒计算的过期时间
	RedisRdbOpcodeSelectdb = 254     // 选择数据库
	RedisRdbOpcodeEof  = 255         // 数据库的结尾（但不是RDB文件的结尾)
)

// 检查给定类型是否对象
func RdbIsObjectType(t int) bool {
	return (t >= 0 && t <= 4) || (t >= 9 && t <= 13)
}