package consts

const(
	RedisVersion = "1.0.0"
	RedisServerPort = 5000
	RedisPidFile = "/var/run/redis.pid"
	RedisMaxClients = 10000
	RedisMinReservedFDs = 32
	RedisEventLoopFdSetIncr = RedisMinReservedFDs + 96
	RedisIoBufLen = 1024 * 16
	RedisCommandParamsCount = 10
	RedisListMaxZipListEntries = 512 // 压缩列表最大节点数
	RedisListMaxZipListValue = 64 // 压缩列表值的最大长度
	RedisDbNum  = 16  // 数据库数量
	RedisZsetMaxZiplistEntries = 512
	RedisZsetMaxZiplistValue = 64
	// lru
	RedisLruCLockResolution = 1000 // lru clock resolution in ms
	RedisLruBits = 24
	RedisLruClockMax = (1 << RedisLruBits) - 1  // max value of obj.Lru
)
