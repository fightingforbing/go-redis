package object

import "ddvideo/go_redis_server/consts"


type lister interface {
	Len() int
	Index(int) *Object
	Push(object *Object, where int) 
	Pop(where int) *Object
}

type List struct {
	encoding int
	bottom  lister
}

func createList() *List {
	return &List{
		encoding: consts.RedisEncodingZipList,
		bottom: createZipList(),
	}
}

func (l *List) Len() int {
	return l.bottom.Len()
}


func (l *List) Index(index int) *Object {
	return l.bottom.Index(index)
}

func (l *List) Push(o *Object, where int) {
	l.tryConversion(o)
	l.bottom.Push(o, where)
}

func (l *List) Pop(where int) *Object {
	return l.bottom.Pop(where)
}

func (l *List) tryConversion(o *Object) {
	if l.encoding != consts.RedisEncodingZipList {
		return
	}
	// 字符串过长
	if s, ok := o.Data.(string); ok {
		if len(s) > consts.RedisListMaxZipListValue {
			l._convert()
			return
		}
	}

	// 压缩列表的长度达到上限
	if l.bottom.Len() >= consts.RedisListMaxZipListEntries {
		l._convert()
		return
	}
}

// 将链表的底层编码从ziplist转化成双端链表
func (l *List) _convert() {
	linkedList := createLinkedList()
	zList := l.bottom.(*zipList)
	for i :=0; i < zList.Len(); i++ {
		o := zList.Index(i)
		linkedList.Push(o,consts.RedisTail)
	}
	l.bottom = linkedList
	l.encoding = consts.RedisEncodingLinkedList
}
