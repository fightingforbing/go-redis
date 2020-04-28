package object

import (
	"container/list"
	"ddvideo/go_redis_server/consts"
)

type linkedList struct {
	list.List
}

func createLinkedList() *linkedList{
	return &linkedList{}
}


func (l *linkedList) Push(o *Object, where int) {
	if where == consts.RedisHead {
		l.PushFront(o)
	} else {
		l.PushBack(o)
	}
}

func (l *linkedList) Find(o *Object) bool {
	// 所以这里还是要自己实现去写的
	for e := l.Front(); e != nil ; e = e.Next() {
		eo := e.Value.(*Object)
		if *eo == *o {
			return true
		}
	}
	return false
}

func (l *linkedList) Pop(where int) *Object {
	var e *list.Element
	if where == consts.RedisHead {
		e = l.Front()
	} else {
		e = l.Back()
	}
	if e == nil {
		return nil
	}
	return l.Remove(e).(*Object)
}

func (l *linkedList) Index(index int) (o *Object) {
	if index < 0 { // 负索引
		index = (-index) - 1
		for e,i := l.Back(), (-index) - 1; e != nil && i >= 0; e,i = e.Prev(),i-1 {
			index = i
			o = e.Value.(*Object)
		}
	} else { // 正索引
		for e,i := l.Front(), index; e != nil && i >= 0; e,i = e.Next(), i-1 {
			index = i
			o = e.Value.(*Object)
		}
	}
	if index < 0 {
		return o
	}
	return nil
}