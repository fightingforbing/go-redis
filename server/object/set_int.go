package object

import "ddvideo/go_redis_server/consts"

const(
	intSet_ENC_INT16 = 1
	intSet_ENC_INT32 = 2
	intSet_ENC_INT64 = 3

	// add where
	intSet_ADD_UNKNOW = 1
	intSet_ADD_HEAD = 2
	intSet_ADD_TAIL = 3
)


type intSet struct {
	data interface{} // may be []int16、 []int32、 []int64
	len int
}

func createIntSet() *intSet{
	return &intSet{data: make([]int16,0)}
}


func (s *intSet) Len() int {
	return s.len
}

func (s *intSet) IsMember(o *Object) bool {
	v := o.Data.(int64)
	_, find := s.search(v)
	return find
}

func (s *intSet) Add(o *Object) (success bool) {
	v := o.Data.(int64)
	s._convertEncoding(v) // 判断是否扩容
	lastPos, find := s.search(v)
	if find {
		return false
	}
	s.insert(lastPos, v)
	return true
}

func (s *intSet) insert(pos int, v int64) {
	switch decodeData := s.data.(type) {
	case []int16:
		tmp := decodeData[:pos]
		tmp = append(tmp, int16(v))
		tmp = append(tmp, decodeData[pos:]...)
		s.data = tmp
	case []int32:
		tmp := decodeData[:pos]
		tmp = append(tmp, int32(v))
		tmp = append(tmp, decodeData[pos:]...)
		s.data = tmp
	case []int64:
		tmp := decodeData[:pos]
		tmp = append(tmp, v)
		tmp = append(tmp, decodeData[pos:]...)
		s.data = tmp
	}
	s.len += 1
}

func (s *intSet) Remove(o *Object) bool {
	v := o.Data.(int64)
	lastpost, find := s.search(v)
	if !find { // 没有找到
		return false
	}
	s.delete(lastpost)
	return true
}


func (s *intSet) delete(pos int) {
	switch decodeData := s.data.(type) {
	case []int16:
		s.data = append(decodeData[:pos], decodeData[pos+1:]...)
	case []int32:
		s.data = append(decodeData[:pos], decodeData[pos+1:]...)
	case []int64:
		s.data = append(decodeData[:pos], decodeData[pos+1:]...)
	}
	s.len -= 1
}

func (s *intSet) Find(v int64) bool {
	_, find := s.search(v)
	return find
}


func (s *intSet) Index(pos int) int64 {
	switch decodeData := s.data.(type) {
	case []int16:
		return int64(decodeData[pos])
	case []int32:
		return int64(decodeData[pos])
	case []int64:
		return decodeData[pos]
	}
	panic("intSet bottom type error")
}

// 如果找到就返回true
// 如果没找到就返回比value小但是离value最近的位置
func (s *intSet) search(v int64) (lastPos int, find bool) {
	if s.len == 0 {
		return 0, false
	}

	// 因为底层数组是有序的，如果value比数组中最后一个值都要大
	// 那么value肯定不存在于集合中
	// 并且应该将value增加到底层数组到最末端
	if v > s.Index(s.len - 1) {
		return  s.len - 1, false
	}
	// 因为底层数组是有序到，如果value比数组中最前一个值都小
	// 那么value肯定不存在于集合中
	// 并且应该将它增加到底层数组到最前端
	if v < s.Index(0) {
		return 0, false
	}

	// 二分查找
	min, max, mid, cur := 0, s.len - 1, -1, int64(-1)
	for max >= min {
		mid = (min + max) / 2
		cur := s.Index(mid)
		if  v > cur {
			min = mid + 1
		} else if v < cur {
			max = mid - 1
		} else {
			break
		}
	}
	if v == cur { // 查找到了
		return mid, true
	}
	return min, false
}


// 返回是否扩容
func (s *intSet) _convertEncoding(v int64) {
	vEncoding := _valueEncoding(v)
	switch decodeData := s.data.(type) {
	case []int16:
		if vEncoding == intSet_ENC_INT32 { // []int16 => []int32
			resizeData := make([]int32, len(decodeData))
			for i := 0; i < len(decodeData); i++ {
				resizeData[i] = int32(decodeData[i])
			}
		} else if vEncoding == intSet_ENC_INT64 { // []int16 => []int64
			resizeData := make([]int64, len(decodeData))
			for i := 0; i < len(decodeData); i++ {
				resizeData[i] = int64(decodeData[i])
			}
		}
	case []int32:
		if vEncoding == intSet_ENC_INT64 {
			resizeData := make([]int64, len(decodeData))
			for i := 0; i < len(decodeData); i++ {
				resizeData[i] = int64(decodeData[i])
			}
		}
	}
}

func _valueEncoding(v int64) int {
	if v < consts.RangeInt32Min || v > consts.RangeInt32Max {
		return intSet_ENC_INT64
	} else if v < consts.RangeInt16Min || v > consts.RangeInt16Max {
		return intSet_ENC_INT32
	}
	return intSet_ENC_INT16
}