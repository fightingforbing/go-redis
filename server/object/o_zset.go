package object

import "ddvideo/go_redis_server/consts"

type zseter interface {
	Insert(member *Object, score float64) bool
	Delete(member *Object) bool
	Len() int
}

func createZset() *Zset {
	var bottom zseter
	if consts.RedisZsetMaxZiplistEntries == 0 {
		bottom = createZsetSkiplist()
	} else {
		bottom = createZsetZiplist()
	}
	return &Zset{
		encoding: consts.RedisEncodingZipList,
		bottom:   bottom,
	}
}

type Zset struct {
	encoding int
	bottom zseter
}

func (zs *Zset) Len() int {
	return zs.bottom.Len()
}

func (zs *Zset) Insert(member *Object, score float64) bool {
	return zs.bottom.Insert(member,score)
}

func (zs *Zset) Delete(member *Object) bool {
	return zs.bottom.Delete(member)
}

func (zs *Zset) tryConvert(o *Object) {
	switch zs.encoding {
	case consts.RedisEncodingSkipList:
		zs.tryConvertSkipToZip(o)
	case consts.RedisEncodingZipList:
		zs.tryConvertZipToSkip(o)
	}
}

func (zs *Zset) tryConvertSkipToZip(o *Object) {

}

func (zs *Zset) tryConvertZipToSkip(o *Object) {
	needConvert := false
	if s, ok := o.Data.(string); ok {
		if len(s) > consts.RedisListMaxZipListValue {
			needConvert = true
		}
	}
	if zs.Len() > consts.RedisListMaxZipListEntries {
		needConvert = true
	}
	if needConvert {

	}
}