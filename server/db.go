package server

import (
	"ddvideo/go_redis_server/consts"
	"ddvideo/go_redis_server/server/object"
	"time"
)

// map key 由Object更新成string类型，减少内存
type redisDb  struct {
	Dict map[string]*object.Object // 所有的键值对
	Expires map[string]time.Time	// 键对过期时间
	BlockIngKeys map[string][]*redisClient // 阻塞的keys
	WatchKeys    map[string][]*redisClient // 监视的键
}

func createRedisDb() *redisDb {
	return &redisDb{
		Dict:         make(map[string]*object.Object),
		Expires:      make(map[string]time.Time),
		BlockIngKeys: make(map[string][]*redisClient),
		WatchKeys:    make(map[string][]*redisClient),
	}
}


func (db *redisDb) lookupKey(key *object.Object) *object.Object {
	v,ok := db.Dict[key.Data.(string)]
	if !ok {
		return nil
	}
	return v
}

func (db *redisDb) deleteKey(key *object.Object)  {
	keyStr := key.Data.(string)
	// 删除过期键
	delete(db.Expires, keyStr)
	// 删除键值
	delete(db.Dict, keyStr)
}

func (db *redisDb) deleteWatchClient(key *object.Object, c *redisClient) {
	keyStr := key.Data.(string)
	clients,ok := db.WatchKeys[keyStr]
	if !ok {
		return
	}
	pos := -1
	for index, client := range clients {
		if client == c {
			pos  = index
			break
		}
	}
	if pos >= 0 {
		db.WatchKeys[keyStr] = append(clients[:pos], clients[pos+1:]...)
	}
}

func (db *redisDb) appendWatchClient(key *object.Object, c *redisClient) {
	keyStr := key.Data.(string)
	clients, ok := db.WatchKeys[keyStr]
	if ok {
		db.WatchKeys[keyStr] = append(clients, c)
	}
	db.WatchKeys[keyStr] = []*redisClient{c}
}

func (db *redisDb) setExpire(key *object.Object, d time.Duration) {
	db.Expires[key.Data.(string)] = time.Now().Add(d)
}

func (db *redisDb) ttl(key *object.Object) (ttl time.Duration, isPerpet bool) {
	expire, ok := db.Expires[key.Data.(string)]
	if !ok {
		return 0, false
	}
	d := expire.Sub(time.Now())
	if d < 0 {
		return 0, false
	}
	return d, false
}

func (db *redisDb) isExpire(key *object.Object) bool {
	ttl, isPerpet := db.ttl(key)
	if isPerpet || ttl > 0 {
		return false
	}
	return true
}

// 检查key 释放过期key
// true  表示存在键过期并删除
// false 表示键未过期或永久存在
func (db *redisDb) expireIfNeed(key *object.Object) bool {
	obj := db.lookupKey(key)
	if obj == nil {
		return false
	}
	if obj.Type != consts.RedisString {
		return false
	}
	if !db.isExpire(key) {
		return false
	}
	Server.StatExpiredKeys++
	db.deleteKey(key)
	return true
}

func (db *redisDb) addVal(key, val *object.Object) {
	if db.lookupKey(key) != nil {
		return
	}
	db.Dict[key.Data.(string)] = val
}

func (db *redisDb) overWriteVal( key, val *object.Object) {
	if db.lookupKey(key) == nil {
		return
	}
	db.Dict[key.Data.(string)] = val
}

func (db *redisDb) setkey(key, val *object.Object) {
	if db.lookupKey(key) == nil {
		db.addVal(key,val)
	} else {
		db.overWriteVal(key,val)
	}
}
