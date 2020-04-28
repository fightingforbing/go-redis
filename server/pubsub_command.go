package server

import (
	"container/list"
	"ddvideo/go_redis_server/server/object"
)

func subscribeCommand(c *redisClient) {
	for j := 1; j < c.Argc; j++ {
		pubsubScribeChannel(c, c.Argv[j])
	}
}

// 设置客户端c订阅频道channel
// 订阅成功返回1， 如果可会淡已经订阅了该频道， 那么返回0
func pubsubScribeChannel(c *redisClient, channel *object.Object) bool {
	 if _, ok := c.PubsubChannels[*channel]; !ok {
		c.PubsubChannels[*channel] = nil
		clients, ok := Server.PubsubChannels[*channel]
		if !ok {
			clients = list.List{}
		}
		clients.PushBack(c)
		Server.PubsubChannels[*channel]  = clients
	 }
	 subscribeConst := object.CreateStringObject("subscribe")
	 count := object.CreateIntObject(int64(len(c.PubsubChannels)))
	 reply := []*object.Object{subscribeConst, channel, count}
	 c.addReplyArray(reply)
	 return true
}

func unsubscribeCommand(c *redisClient) {
	if c.Argc == 1 { // 全部删除
		pubsubUnsubscribeAllChannels(c, true)
	} else {
		for j := 1; j < c.Argc; j++ {
			pubsubScribeChannel(c,c.Argv[j])
		}
	}
}

// 返回被退订的数量
func pubsubUnsubscribeAllChannels(c *redisClient, notify bool) int {
	count := 0
	// 从当前的client的 channel
	for channel , _ := range c.PubsubChannels {
		if pubsubUnscribeChannel(c, &channel, notify) {
			count++
		}
	}
	if notify && count == 0 {
		// 这里回复逻辑还有点问题
		// 需要重构下代码
	}
	return count
}

// 客户端c 退订频道channel
// 如果取消成功返回1， 如果因为客户端未订阅频道，而造成取消失败，返回0
func pubsubUnscribeChannel(c *redisClient, channel *object.Object, notify bool) bool {
	// 将频道channel从clients.PubsubChannels字典中删除
	// 示意图：
	// before:
	// {
	//  'channel-x': nil,
	//  'channel-y': nil,
	//  'channel-z': nil,
	// }
	// after unsubscribe channel-y ：
	// {
	//  'channel-x': nil,
	//  'channel-z': nil,
	// }
	if _, ok := c.PubsubChannels[*channel]; !ok { // 未订阅该频道
		return false
	}
	delete(c.PubsubChannels, *channel)
	me := Server.PubsubChannels[*channel]
	for e := me.Front(); e != nil; e = e.Next() {
		if e.Value.(*redisClient) == c {
			me.Remove(e)
		}
	}
	return true
}

func publishCommand(c *redisClient) {
	receivers := pubsubPublishMessage(c.Argv[1], c.Argv[2])
	c.addReplyInt64(int64(receivers))
}

// 将message发送到所有订阅频道 channel的客户端
// 以及所有订阅了和channel频道匹配的模式的客户端
func pubsubPublishMessage( channel, message *object.Object) int {
	var receivers int
	// 取出包含所有订阅频道channel的客户端的链表
	if clients, ok := Server.PubsubChannels[*channel]; ok {
		for e := clients.Front(); e !=nil; e = e.Next() {
			c := e.Value.(*redisClient)
			messageConst := object.CreateStringObject("message")
			reply := []*object.Object{messageConst, channel, message}
			c.addReplyArray(reply)
			receivers++
		}
	}
	return receivers
}
