package object

// 这里定义一个ziplist
// 用来做zlist的存储
// 这个应该是key value
// 两个连续的这样
// 就是不知道他是存储的key
// 还是value


type zsetZiplist struct {
	ziplist *zipList
}
// member -> score -> member -> score
func createZsetZiplist() *zsetZiplist {
	return &zsetZiplist{ziplist:createZipList()}
}

// 先推入元素，后推入分值
func (zz *zsetZiplist) Push(member, score *Object) {
	// 这里取出来分值
	// 然后一个循环
	// 找到合适的位置
	//
}

func (zz *zsetZiplist) Len() int {
	return 0
}

func (zz *zsetZiplist) Insert(member *Object, score float64) bool {
	if zz.ziplist.Len() == 0 {
		zz.ziplist.InsertAt(0,member, CreateFloat64Object(score))
	}
	for i := 0; i < zz.ziplist.Len(); i += 2 {
		scoreObj, _ := zz.ziplist.Index(i)
		memberObj, pos  := zz.ziplist.Index(i+1)
		curScore,_ := ConvertObjectToFloat64(scoreObj)
		if curScore > score {
			zz.ziplist.InsertAt(pos, member, CreateFloat64Object(score))
			break
		} else if curScore == score { // 分值相等
			// 根据member的字符串位置来决定新节点的插入位置
			if CompareStringObject(memberObj, member) > 0 {
				zz.ziplist.InsertAt(pos, member, CreateFloat64Object(score))
			}
			break
		}
	}
	return true
}

// 从ziplist 编码的有序结合中查找ele成员， 并将它的分值保存到score
// 寻找成功返回指向成员ele的指针，查找失败返回null
func (zz *zsetZiplist) find(member *Object) (score float64, memberPos int) {
	pos := 0
	for !zz.ziplist.isEnd(pos) {
		scoreEle := zz.ziplist.GetCurEntry(pos)
		pos += scoreEle.getBufLen()
		memberEle := zz.ziplist.GetCurEntry(pos)
		if *memberEle.v == *member {
			score, _ := ConvertObjectToFloat64(scoreEle.v)
			return score, pos
		}
		pos += memberEle.getBufLen()
	}
	return 0, -1
}

func (zz *zsetZiplist) Delete(member *Object) bool {
	pos := zz.ziplist.Find(member)
	if pos == -1 { // 不存在
		return false
	}
	// num 2 删除成员和分值
	zz.ziplist.delete(pos, 2)
	return true
}