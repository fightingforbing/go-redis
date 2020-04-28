package object

type mapSet struct {
	data map[Object]interface{}
}

func createMapSet() *mapSet {
	return &mapSet{
		data: make(map[Object]interface{}),
	}
}

func (ms *mapSet) Len() int {
	return len(ms.data)
}

func (ms *mapSet) Add(o *Object) bool {
	if ms.IsMember(o) {
		return false
	}
	ms.data[*o] = nil
	return true
}


func (ms *mapSet) IsMember(o *Object) bool {
	return ms.IsMember(o)
}

func (ms *mapSet) Remove(o *Object) bool {
	if !ms.IsMember(o) {
		return false
	}
	delete(ms.data, *o)
	return true
}