package object

import "ddvideo/go_redis_server/consts"

type Set struct {
	encoding int
	bottom seter
}

type seter interface {
	Add(o *Object) bool
	IsMember(o *Object) bool
	Remove(o *Object) bool
	Len() int
}

func createSet() *Set {
	return &Set{
		encoding: consts.RedisEncodingIntSet,
		bottom:   createIntSet(),
	}
}

func (s *Set) Len() int {
	return s.bottom.Len()
}

func (s *Set) Add(o *Object) bool {
	s.tryTypeConvert(o)
	return s.bottom.Add(o)
}

func (s *Set) IsMember(o *Object) bool {
	return s.bottom.IsMember(o)
}

func (s *Set) Remove(o *Object) bool {
	return s.bottom.Remove(o)
}

func (s *Set) tryTypeConvert(os ...*Object) {
	var (
		needConvert bool
	)
	if s.encoding != consts.RedisEncodingIntSet {
		return
	}
	for _, o := range os {
		if o.Type != consts.RedisInt {
			needConvert = true
			break
		}
	}
	if needConvert {
		mapset := createMapSet()
		iset := s.bottom.(*intSet)
		// 将压缩列表的数据拷贝到map
		for i := 0; i < iset.Len(); i++ {
			o := CreateIntObject( iset.Index(i) )
			mapset.Add(o)
		}
		s.bottom = mapset
		s.encoding = consts.RedisEncodingHt
	}
}