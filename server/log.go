package server

import(
	"fmt"
)

const (
	RedisDebug = iota
	RedisVerBose
	RedisNotice
	RedisWarning
	RedisLogRaw = 1 << 10 // modifier to log without timestamp
	RedisDefaultVerbosity = RedisNotice
)

func redisLog(level int, fmt string) {
	fmt.Println("")
}
