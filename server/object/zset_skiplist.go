package object

type zsetSkiplist struct {
	m map[string]float64
	skip *skiplist
}

func createZsetSkiplist() *zsetSkiplist {
	return &zsetSkiplist{
		m: make(map[string]float64),
		skip: createSkiplist(),
	}
}

func (z *zsetSkiplist) Len() int {
	return int(z.skip.len)
}

func (z *zsetSkiplist) Insert(member *Object, score float64) bool {
	curscore, ok := z.m[member.Data.(string)]
	if ok { // 成员存在
		if score != curscore {
			// 删除原有元素
			z.Delete(member)
		}
	}
	z.Insert(member,score)
	z.m[member.Data.(string)] = score
	return true
}

func (z *zsetSkiplist) Delete(member *Object) bool {
	return z.skip.Delete(member, z.m[member.Data.(string)])
}