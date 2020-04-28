package object

import (
	"bytes"
	"ddvideo/go_redis_server/consts"
	"encoding/binary"
)

const (
	ZipBigLen uint8 = 254  // 前置节点长度占用字节数 5字节长度标识符

	/*
	编码													编码长度	connect部分内容的值
	00bbbbbb											1byte	长度小于等于63
	01bbbbbb cccccc										2byte	长度小于等于 16383 字节的字符数组
	10______ aaaaaaaa bbbbbbbb cccccccc dddddddd 		5byte 	长度小于等于 4294967295 的字符数组 (_表示留空)
	11000000											1byte	int16_t类型的整数
	11010000											1byte	int32_t类型的整数
	11100000											1byte   int64_t类型的整数
	11110000											1byte   24bit有符号整数
	11111110											1byte   8bit有符号整数
	1111xxxx											1byte	4bit无符号整数，介于0至12之间
	*/

	ZipStrMask uint8 = 0xc0  // 字符串编码的掩码
	ZipIntMask uint8 = 0x30  // 整数编码的掩码

	// 整数编码类型
	ZipInt16B  = ZipStrMask | (0 << 4)
	ZipInt32B  = ZipStrMask | (1 << 4)
	ZipInt64B  = ZipStrMask | (2 << 4)
	ZipInt24B  = ZipStrMask | (3 << 4)
	ZipInt8B   = 0xfe

	ZipIntImmMin uint8 = 0xf1

	// 字符串
	ZipStr06B uint8 = 0 << 6
	ZipStr14B uint8 = 1 << 6
	ZipStr32B uint8 = 2 << 6
)

/*
非空 ziplist 示例图

area        <----------- entries ------------->|

size            ?        ?        ?        ?
            +--------+--------+--------+--------+
component     entry1 | entry2 |  ...   | entryN |
            +--------+--------+--------+--------+
            ^                          ^
address     |                          |
      ZIPLIST_ENTRY_HEAD               |
                                       |
                               ZIPLIST_ENTRY_TAIL
*/

type zipList struct {
	tail int // 最后节点起始位置
	entrysNum  uint16 // 节点数量
	buf []byte // 链表存储内容
}

func createZipList() *zipList {
	return &zipList{
		tail: 0,
		entrysNum: 0,
		buf: nil,
	}
}

func (zl *zipList) Encoding() int {
	return consts.RedisEncodingZipList
}

func (zl *zipList) isEmpty() bool {
	if zl.entrysNum == 0 {
		return true
	}
	return false
}

func (zl *zipList) isEnd(pos int) bool {
	return pos != 0 && len(zl.buf) == pos
}



func (zl *zipList) Len() int {
	return int(zl.entrysNum)
}

// pos 为-1不存在
func (zl *zipList) Find(o *Object) (pos int) {
	if zl.Len() == 0 {
		return -1
	}

	for !zl.isEnd(pos) {
		e := zl.GetCurEntry(pos)
		ev := e.getV()
		if *ev == *o {
			return pos
		}
		pos += e.getBufLen()
	}
	return -1
}

func (zl *zipList) GetTailEntry() *zipEntry {
	return zl.GetCurEntry(zl.tail)
}

func (zl *zipList) GetPrevEntry(pos int) *zipEntry {
	cur := zl.GetCurEntry(pos)
	if cur.getPrevlen() == 0 { // 前驱节点不存在
		return nil
	}
	pos -= cur.getPrevlen()
	return zl.GetCurEntry(pos)
}

func (zl *zipList) GetNextEntry(pos int) *zipEntry {
	cur := zl.GetCurEntry(pos)
	if cur == nil {
		return nil
	}
	pos += cur.getBufLen()
	return zl.GetCurEntry(pos)
}

func (zl *zipList) GetCurEntry(pos int) *zipEntry {
	if zl.isEnd(pos) {
		return nil
	}
	if zl.isEmpty() {
		return nil
	}
	if len(zl.buf) < pos {
		return nil
	}
	return createZipEntryByDecode(zl.buf[pos:])
}

func (zl *zipList) Pop(where int) *Object {
  var (
  	index int
  	pos int
  )
  if where == consts.RedisHead {
  	index = 0
  	pos = 0
  } else {
  	index = -1
  	pos = zl.tail
  }
  o, _  := zl.Index(index)
  if o != nil {
  	zl.delete(pos,1)
  }
  return o
}

func (zl *zipList) Push(o *Object, where int) {
	var pos int
	if where == consts.RedisHead {
		pos = 0
	}else {
		pos = len(zl.buf)
	}
	zl.Insert(o, pos)
}

func (zl *zipList) insertEntry( e *zipEntry, pos int) {
	buf := new(bytes.Buffer)
	buf.Write(zl.buf[:pos])
	buf.Write(e.Bytes())
	buf.Write(zl.buf[pos:])
	zl.buf = buf.Bytes()
}

// note: 只负责替换节点，不负责检查更新后驱节点是否有足够空间存储节点长度
func (zl *zipList) updateEntry( e *zipEntry, pos int) {
	curEntry := zl.GetCurEntry(pos)
	if curEntry == nil {
		return
	}
	buf := new(bytes.Buffer)
	buf.Write(zl.buf[:pos])
	buf.Write(e.Bytes())
	buf.Write(zl.buf[pos + curEntry.getBufLen():])
	zl.buf = buf.Bytes()
}

func (zl *zipList) InsertFront(o *Object) {
	zl.Insert(o, 0)
}

func (zl *zipList) InsertBack(o *Object) {
	zl.Insert(o, zl.tail)
}

func (zl *zipList) InsertAt(pos int, objs ...*Object) {
	for _ ,obj := range objs {
		zl.Insert(obj, pos)
		pos = zl.GetCurEntry(pos).getBufLen() + pos
	}
}

func (zl *zipList) Insert(o *Object, pos int) {
	var (
		prevEntry *zipEntry // 插入位置前一个节点的信息
	)
	if zl.isEnd(pos) { // 在链表未追加数据
		prevEntry = zl.GetTailEntry()
	} else {
		prevEntry = zl.GetPrevEntry(pos)
	}
	reqEntry := createZipEntryByEncodeObject(prevEntry.getBufLen(), o)
	if !zl.isEmpty() {
		if zl.isEnd(pos) {
			zl.tail += prevEntry.getBufLen()
		} else {
			zl.tail += reqEntry.getBufLen()
		}
	}

	// 插入数据到链表
	zl.insertEntry(reqEntry, pos)
	// 更新后续节点
	zl.cascadeUpdate(pos)
	zl.entrysNum++
}

func (zl *zipList) delete(pos, num int) {
	var (
		firstPos int = pos
		deleted uint16
	)

	for i :=0; i < num; i++ {
		if zl.isEnd(pos) {
			break
		}
		e := zl.GetCurEntry(pos)
		pos += e.getBufLen()
	}
	if pos > firstPos { // 有节点被删除
		zl.buf = append(zl.buf[0:firstPos], zl.buf[pos:]...)
		zl.tail -= pos - firstPos
		prevPos := firstPos - zl.GetPrevEntry(firstPos).getBufLen()
		zl.cascadeUpdate( prevPos )
		zl.entrysNum -= deleted
	}
}

/*
+----------+----------+----------+----------+----------+----------+
|          |          |          |          |          |          |
|   prev   |   new    |   next   | next + 1 | next + 2 |   ...    |
|          |          |          |          |          |          |
+----------+----------+----------+----------+----------+----------+

当插入new节点时，新的 new 节点取代原来的 prev 节点， 成为了 next 节点的新前驱节点，
不过， 因为这时 next 节点的 pre_entry_length 域编码的仍然是 prev 节点的长度，
所以程序需要将 new 节点的长度编码进 next 节点的 pre_entry_length 域里，
这里会出现三种可能：

1.next 的 pre_entry_length 域的长度正好能够编码 new 的长度（都是 1 字节或者都是 5 字节）
2.next 的 pre_entry_length 只有 1 字节长，但编码 new 的长度需要 5 字节
3.next 的 pre_entry_length 有 5 字节长，但编码 new 的长度只需要 1 字节
对于情况 1 和 3 ， 程序直接更新 next 的 pre_entry_length 域。

如果是第二种情况， 那么程序必须对 ziplist 进行内存重分配，
从而扩展 next 的空间。 然而，因为 next 的空间长度改变了，
所以程序又必须检查 next 的后继节点 —— next+1 ，
看它的 pre_entry_length 能否编码 next 的新长度，
如果不能的话，程序又需要继续对 next+1 进行扩容。。。
 */
func (zl *zipList) cascadeUpdate(pos int) {
	for {
		curEntry := zl.GetCurEntry(pos)
		if curEntry == nil {
			return
		}
		nextEntry := zl.GetNextEntry(pos)
		if nextEntry == nil {
			return
		}
		pos += curEntry.getBufLen() // pos指向next节点
		offset := lookupNextEntrySavePrevSizeOffset(curEntry, nextEntry)
		if offset > 0 { // 空间不足，需要扩容
			resizeEntry := createZipEntryByUpdatePrevlen(nextEntry, curEntry.getBufLen())
			 zl.updateEntry(resizeEntry, pos)
			// 更新tail的位置
			zl.tail += offset
		}
	}
}

func (z *zipList) Index(index int) (*Object, int) {
	e, pos := z._index(index)
	if e == nil {
		return nil, 0
	}
	return e.getV(), pos
}

func (z *zipList) _index(index int) (*zipEntry, int) {
	var (
		pos int
		entry *zipEntry
	)
	if index < 0 { // 负索引
		index = (-index) - 1
		pos = z.tail
		entry = z.GetCurEntry(pos)
		for  entry.getPrevlen() > 0 && index > 0 {
			index--
			pos -= entry.getPrevlen()
			entry = z.GetCurEntry(pos)
		}
	} else { // 正索引
		pos = 0
		for !z.isEnd(pos) && index >= 0 {
			index--
			entry = z.GetCurEntry(pos)
			pos += entry.getBufLen()
		}
	}
	if index >= 0 {
		return nil, 0
	}
	return entry, pos
}

// 查询当前节点的后驱节点是否用足够空间存储当前节点的长度
// 返回正数表示空间不足
// 返回值可能为 0,4,-4
func lookupNextEntrySavePrevSizeOffset(cur, next *zipEntry) int {
	lenNeedSize, _ := encodePrevlen(cur.getBufLen(),false)
	nextHavePrevSize := next.getPrevlenSize()
	return int( lenNeedSize - nextHavePrevSize )
}

type zipEntry struct {
	prevlen int // 前一个节点字节长度
	prevlenSize uint8 // 存储前置节点字节长度占用度字节数
	encoding uint8 // 存储当前节点内容长度的编码
	lenSize uint8 // 存储当前节点内容的长度需要的字节空间数
	vlen int // 当前节点内容的长度
	v *Object // 当前节点的内容
	encodeBuf []byte // 当前节点总的编码内容
}

func (z *zipEntry) getBufLen() int {
	if z == nil {
		return 0
	}
	return len(z.encodeBuf)
}

func (z *zipEntry) getV() *Object {
	return z.v
}

func (z *zipEntry) getLensize() uint8 {
	return z.lenSize
}

func (z *zipEntry) getVlen() int {
	return z.vlen
}

func (z *zipEntry) getEncoding() uint8 {
	return z.encoding
}

func (z *zipEntry) getPrevlenSize() uint8 {
	return z.prevlenSize
}

func (z *zipEntry) getPrevlen() int {
	if z == nil {
		return 0
	}
	return z.prevlen
}

func (z *zipEntry) Bytes() []byte {
	return z.encodeBuf
}

func (z *zipEntry) prevlenBytes() []byte {
	return z.encodeBuf[:z.prevlenSize]
}

func (z *zipEntry) encodingAndContentBytes() []byte {
	return z.encodeBuf[z.prevlenSize:]
}

func createZipEntryByUpdatePrevlen(z *zipEntry, prevlen int) *zipEntry {
	prevlenSize, prevlenEncodeData := encodePrevlen(prevlen, true)
	buf := append(prevlenEncodeData, z.encodingAndContentBytes()...)
	return &zipEntry{
		prevlen:     prevlen,
		prevlenSize: prevlenSize,
		encoding:    z.getEncoding(),
		lenSize:     z.getLensize(),
		vlen:        z.getVlen(),
		v:           z.getV(),
		encodeBuf:   buf,
	}
}

func createZipEntryByDecode(bs []byte) *zipEntry {
	if len(bs) == 0 {
		return nil
	}
	prevlen, prevlenSize := decodePrelen(bs)
	encoding, lensize, strlen, v := decodeEncodingAndContent(bs[prevlenSize:])
	return &zipEntry{
		prevlen: prevlen,
		prevlenSize: prevlenSize,
		encoding: encoding,
		lenSize: lensize,
		v: v,
		vlen: strlen,
		encodeBuf: bs[: int(prevlenSize) + 1/*encoding占用一个字节*/ + int(lensize) + strlen],
	}
}

// prevlen 前置节点的总占用的字节数
// prevlenSize 保存前置节点占用的字节数
func decodePrelen(bs []byte) (prevlen int, prevlenSize uint8) {
	if bs[0] < ZipBigLen {
		prevlenSize = 1
		prevlen = int(bs[0])
	} else {
		prevlenSize = 5
		prevlen = int (binary.BigEndian.Uint32(bs[1:]) )
	}
	return
}
// encoding 数据类型
// lensize 存储数据长度占用的字节数
// slen   字符串数据的长度
// v  存储的内容
func decodeEncodingAndContent(bs []byte) (encoding uint8, lensize uint8, slen int, obj *Object){
	encoding = bs[0]
	if encoding < ZipStrMask { // 字符串
		encoding = encoding & 0xc0
		switch encoding {
		case ZipStr06B:
			lensize = 0
			slen = int(bs[0] & 0x3f)
		case ZipStr14B:
			lensize = 1
			slen = int(  uint16( (bs[0] & 0x3f) << 8 ) | uint16(bs[1]) )
		case ZipStr32B:
			lensize = 4
			slen = int( binary.BigEndian.Uint32(bs[1 + lensize:]) )
		}
		vs := string( bs[1 + lensize: (1 + int(lensize) + slen)])
		obj = CreateStringObject(vs)
	} else { // 整数
		var vi int64
		switch encoding {
		case ZipInt8B:
			lensize = 1
			slen = 0
			vi = int64(bs[1])
		case ZipInt16B:
			lensize = 2
			slen = 0
			vi = int64(binary.BigEndian.Uint16(bs[1:]))
		case ZipInt24B:
			lensize = 3
			slen = 0
			vi = int64( (bs[1] << 16) | (bs[2] << 8) | bs[3] )
		case ZipInt32B:
			lensize = 4
			slen = 0
			vi = int64(binary.BigEndian.Uint32(bs[1:]))
		case ZipInt64B:
			lensize = 8
			slen = 0
			vi = int64(binary.BigEndian.Uint64(bs[1:]))
		}
		obj = CreateIntObject(vi)
	}
	return
}


func createZipEntryByEncodeObject(prevlen int, obj *Object) *zipEntry {
	prevlenSize, prevlenEncodeData := encodePrevlen(prevlen, true)
	encoding, lensize, strlen, contentEncodeData := encodeContent(obj)
	return &zipEntry{
		prevlen:     prevlen,
		prevlenSize: prevlenSize,
		encoding:    encoding,
		lenSize:     lensize,
		vlen:        strlen,
		v:           obj,
		encodeBuf:   append(prevlenEncodeData, contentEncodeData...),
	}
}

func encodePrevlen(prevlen int, needEncodeData bool) (prevlenSize uint8, encodeData []byte) {
	buf := new(bytes.Buffer)
	if prevlen < int(ZipBigLen) {
		prevlenSize = 1
		if needEncodeData {
			binary.Write(buf, binary.BigEndian, uint8(prevlen))
		}
	} else {
		prevlenSize = 5
		if needEncodeData {
			binary.Write(buf,binary.BigEndian, uint8(ZipBigLen)) // 存储254
			binary.Write(buf,binary.BigEndian, int32(prevlen)) // 存储实际长度
		}
	}
	if needEncodeData {
		encodeData = buf.Bytes()
	}
	return
}

func encodeContent(obj *Object) (encoding uint8, lensize uint8, strlen int, encodeData []byte) {
	var (
		buf = new(bytes.Buffer)
	)
	if obj.Type != consts.RedisString {
		panic("ziplist entry encode type err")
	}

	switch v := obj.Data.(type) {
	case string:
		strlen = len(v)
		if strlen <= 0x3f {
			encoding = ZipStr06B
			lensize = 0
			binary.Write(buf, binary.BigEndian, int8(ZipStr06B | uint8(strlen) ) )
		} else if strlen <= 0x3fff {
			encoding = ZipStr14B
			lensize = 1
			binary.Write(buf, binary.BigEndian, ZipStr14B | uint8(strlen >> 8))
			binary.Write(buf, binary.BigEndian, int8(strlen))
		} else {
			encoding = ZipStr32B
			lensize = 4
			binary.Write(buf, binary.BigEndian, ZipStr32B)
			binary.Write(buf, binary.BigEndian, int32(strlen))
		}
		buf.WriteString(v)
	case int64:
		vInteger := v
		if vInteger >= 0 && vInteger <= 12 {
			encoding = ZipIntImmMin + uint8(vInteger)
			lensize = 0
			binary.Write(buf, binary.BigEndian, encoding)
		} else if vInteger >= consts.RangeInt8Min && vInteger <= consts.RangeInt8Max {
			encoding = ZipInt8B
			lensize = 1
			binary.Write(buf, binary.BigEndian, encoding)
			binary.Write(buf, binary.BigEndian, int8(vInteger))
		} else if vInteger >= consts.RangeInt16Min && vInteger <= consts.RangeInt16Max {
			encoding = ZipInt16B
			lensize = 2
			binary.Write(buf, binary.BigEndian, encoding)
			binary.Write(buf, binary.BigEndian, int16(vInteger))
		} else if vInteger >= consts.RangeInt24Min && vInteger <= consts.RangeInt24Max {
			encoding = ZipInt24B
			lensize = 3
			binary.Write(buf, binary.BigEndian, encoding)
			binary.Write(buf, binary.BigEndian, int8(vInteger >> 16))
			binary.Write(buf, binary.BigEndian, int8(vInteger >> 8))
			binary.Write(buf, binary.BigEndian, int8(vInteger))
		} else if vInteger >= consts.RangeInt32Min && vInteger <= consts.RangeInt32Max {
			encoding = ZipInt32B
			lensize = 4
			binary.Write(buf, binary.BigEndian, encoding)
			binary.Write(buf, binary.BigEndian, int32(vInteger))
		} else {
			encoding = ZipInt64B
			lensize = 8
			binary.Write(buf, binary.BigEndian, encoding)
			binary.Write(buf, binary.BigEndian, int64(vInteger))
		}
	}
	encodeData = buf.Bytes()
	return
}