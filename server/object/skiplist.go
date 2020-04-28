package object

import (
	"fmt"
	"math/rand"
)

const (
	ZSKIPLIST_MAXLEVEL = 32
	ZSKIPLIST_P = 0.25
)

type skiplist struct {
	header *skiplistNode 	// 头节点
	tail   *skiplistNode	// 尾节点
	len  uint64  			// 节点数量
	level int  				// 目前表内节点的最大层数
}

// 创建并返回一个新的跳跃表
func createSkiplist () *skiplist {
	skiplist := new(skiplist)
	skiplist.level = 1
	skiplist.len = 0
	skiplist.header = createSkiplistNode(ZSKIPLIST_MAXLEVEL, 0, nil)
	return skiplist
}


/*
创建一个层数为level的跳跃表节点
并将节点的成员对象设置为obj，分值设置为score
返回值为新创建的跳跃表节点
 */
func createSkiplistNode(level int, score float64, obj *Object) *skiplistNode {
	node := new(skiplistNode)
	node.level = make([]*skiplistLevel, level)
	for i :=0; i < level; i++ {
		node.level[i] = createSkiplistLevel()
	}
	node.score = score
	node.obj = obj
	return node
}

func createSkiplistLevel() *skiplistLevel {
	return &skiplistLevel{}
}

type skiplistLevel struct {
	forward *skiplistNode 	// 前进节点
	span uint64         		// 跨度
}

type skiplistNode struct {
	obj *Object 				// member成员
	score float64 				// 分值
	backward *skiplistNode 	// 回退指针
	level []*skiplistLevel  // 层
}

func (z *skiplist) Len() int {
	return 0
}

/*
创建一个成员为obj, 分值为score的新节点
 */
func (z *skiplist) insert(member *Object, score float64)  {
	// 存储经过的节点跨度
	rank := make([]uint64,ZSKIPLIST_MAXLEVEL)
	x := z.header
	// 存储搜索路径
	update := make([]*skiplistNode, ZSKIPLIST_MAXLEVEL)
	// 在各个层查找节点的插入位置
	for i := z.level - 1; i >= 0; i-- {
		if i == z.level - 1 {
			rank[i] = 0
 		} else {
 			rank[i] = rank[i+1]
		}
		for x.level[i].forward != nil /*前置节点存在*/&&
			( x.level[i].forward.score < score /*比对分值*/||
				(x.level[i].forward.score == score &&
					// 比对成员
					CompareStringObject(x.level[i].forward.obj, member) < 0)){
			// 记录沿途跨越了多少个节点
			rank[i] += x.level[i].span
			// 移动至下一个指针
			x = x.level[i].forward
		}
		// 记录将要和新节点相连接的节点
		update[i] = x
	}
	// 获取一个随机值作为新节点的层数
	level := z.randomLevel()
	// 如果新节点的层数比表中其他节点的层数都要大
	// 那么初始化表头节点中未使用的层，并将他们记录到update数组中
	// 将来也指向新节点
	if level > z.level {
		// 初始化未使用层
		for i := z.level; i < level; i++ {
			rank[i] = 0
			update[i] = z.header
			update[i].level[i].span = z.len
		}
		z.level = level
	}

	// 创建新节点
	x = createSkiplistNode(level, score, member)
	// 将前面记录的指针指向新节点， 并做相应的设置
	for i := 0; i < level; i++ {
		// 设置新节点的forward指针
		x.level[i].forward = update[i].level[i].forward
		// 将沿途记录的各个节点的forward指针指向新节点
		update[i].level[i].forward = x
		// 计算新节点跨越的节点数量
		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		// 更新新节点插入之后，沿途节点的span值
		// 其中的+1计算的是新节点
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}
	// 未接触的节点的span值也需要增一，这些节点直接从表头指向新节点
	for i := level; i < z.level; i++ {
		update[i].level[i].span++
	}
	// 设置新节点的后退指针
	if update[0] == z.header {
		x.backward = nil
	} else {
		x.backward = update[0]
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	}
	z.len++
}

// 从跳跃表中删除包含给定节点score并且带有
func (z *skiplist) Delete(member *Object, score float64) bool {
	update := make([]*skiplistNode, ZSKIPLIST_MAXLEVEL)
	x := z.header
	for i := z.level -1; i >= 0; i-- {
		for x.level[i].forward != nil /*前置节点存在*/&&
			( x.level[i].forward.score < score /*比对分值*/||
				(x.level[i].forward.score == score &&
					// 比对成员
					CompareStringObject(x.level[i].forward.obj, member) < 0)){
			// 移动至下一个指针
			x = x.level[i].forward
		}
		update[i] = x
	}
	// 检查找到的元素x，只有在它的分值跟对象都相同时，才将它删除
	x = x.level[0].forward
	if x != nil && score == x.score && CompareStringObject(x.obj, member) == 0 {
		z._deleteNode(x,update)
		return true
	}
	return false
}

// 内部删除函数
func (z *skiplist) _deleteNode(x *skiplistNode, update []*skiplistNode) {
	// 更新所有和被删除节点x有关的节点的指针，解除他们之间的关系
	for i := 0; i < z.level; i++ {
		if update[i].level[i].forward == x {
			update[i].level[i].span += x.level[i].span - 1
			update[i].level[i].forward = x.level[i].forward
		} else {
			update[i].level[i].span -= 1
		}
	}
	// 更新被删除节点x的前进和后退方针
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x.backward
	} else {
		z.tail = x.backward
	}
	// 更新跳跃表最大层数（只在有被删除节点是跳跃表中最高的节点时才执行）
	for z.level > 1 && z.header.level[z.level - 1].forward == nil {
		z.level--
	}
	z.len -= 1
}

func (z *skiplist) findMember(member *Object) *skiplistNode {
	// 遍历吗？
	return nil
}

func (z *skiplist) Zrange() {
	ln := z.header.level[0].forward
	for ln != nil {
		fmt.Println(ln.obj.Data.(string))
		ln = ln.level[0].forward
	}
}

func (z *skiplist) GetElementByRank(rank uint64) *skiplistNode {
	x := z.header
	traversed := uint64(0)
	for i := z.level -1; i >= 0; i-- {
		for x.level[i].forward != nil && ( traversed + x.level[i].span) <= rank {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		// 如果越过的节点数量已经等于rank
		// 那么说明已经到达要找的节点
		if traversed == rank {
			return x
		}
	}
	return nil
}

/*
返回一个随机值，用做新跳跃表节点的层数
返回值介乎1 和ZSKIPLIST_MAXLEVEL之间（包含ZSKIPLIST_MAXLEVEL）
根据随机算法所使用的幂次定律，越大的值生成的机率越小
由于 ZSKIPLIST_P = 0.25
ZSKIPLIST_P * 0xFFFF = 0xFFFF >> 2 = 0x3FFFF
假设rand比较均匀，进行0xFFFF高16位清零之后，底16位取值
就落在0x0000-0xFFFF之间，其中落在0x3FFFF内的概率为1/4
定量的分析如下：
* 节点层数恰好等于1的概率为1-p
* 节点层数大于等于2的概率为p, 而节点层数恰好等于2的概率为p(1-p)
* 节点层数大于等于3的概率为p^2,而节点层数恰好等于3的概率为p^2(1-p)
* 节点层数大于等于4的概率为p^3, 而节点层数恰好等于4的概率为p^3(1-p)
 */
func (z *skiplist) randomLevel() int {
	level := 1
	for (rand.Int() & 0xFFFF) < ( 0xFFFF >> 2 /*equal 0xFFFF * skiplist_P*/ ) {
		level++
	}
	if level >= ZSKIPLIST_MAXLEVEL {
		return ZSKIPLIST_MAXLEVEL
	}
	return level
}

func convertPr() int {
	return 1 / ZSKIPLIST_P
}