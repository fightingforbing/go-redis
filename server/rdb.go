package server

import (
	"bufio"
	"ddvideo/go_redis_server/consts"
	"log"
	"os"
	"strconv"
	"time"
)



// 将给定rdb中保存的数据载入到数据库中
func rdbLoad(filename string) bool {
	f, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer f.Close()
	rio := bufio.NewReader(f)
	buf := make([]byte, 9)
	n, err := rio.Read(buf)
	if err != nil || n < 9 {
		log.Print("Wrong signature trying to load DB from file")
		return false
	}
	// 检查版本号
	if string(buf[:5]) != "REDIS" {
		log.Printf("Wrong sign trying to load DB from file")
	}
	rdbver, _ := strconv.Atoi(string(buf[5:9]))
	if rdbver < 1 || rdbver > consts.RedisRdbVersion {
		log.Printf("Can't handle RDB format version %d", rdbver)
	}

	// 将服务器状态调整到开始载入状态
	Server.StartLoading(f)
	for {
		/*
		 * Read type
		 *
		 * 读入类型指示，决定该如何读入之后跟着的数据。
		 * 这个指示可以是consts包中定义的所有以
		 * RedisRdbType* 为前缀的常量的其中一个
		 * 或者说有以RedisRdbOpcode* 为前缀的常量的其中一个
		 */
		t, err := loadType(rio)
		if err != nil {
			goto eoferr
		}
		// 读入过期时间值
		switch t {
		case consts.RedisRdbOpcodeExpireMe: // 以秒计算的过期时间
			expiretime, err := loadTime(rio)
			if err != nil {
				goto eoferr
			}
			// 在过期时间之后会跟着一个键值对，我们要读入这个键值对的类型
			t, err := loadType(rio)
			if err != nil {
				goto eoferr
			}
			// 将格式转化为毫秒
			expiretime *= 1000
		case consts.RedisRdbOpcodeExpiretimeEs:
			// 以毫秒计算的过期时间
			expiretime, err := loadMillisecondTime(rio)
		}
	}
eoferr:
		log.Printf("Short read or OOM loading DB. Unrecoverable error, aborting now.")
		os.Exit(1)
		return false
}

func (s *server) StartLoading(f *os.File) {
	// 正在载入
	s.Loading = true
	// 开始进行载入的时间
	s.LoadingStartTime = time.Now()
	// 文件大小
	fileInfo, _ := f.Stat()
	s.LoadingTotalBytes = fileInfo.Size()
}

func loadType( io *bufio.Reader) (int, error) {
	b, err := io.ReadByte()
	if err != nil {
		return 0, err
	}
	return int(b), nil
}

func loadTime(io *bufio.Reader) (int64, error)  {
	buf := make([]byte, 4)
	if _, err := io.Read(buf); err != nil {
		return 0,  err
	}
	return strconv.ParseInt(string(buf), 10, 32)
}

func loadMillisecondTime(io *bufio.Reader) (int64, error) {
	buf := make([]byte, 8)
	if _, err := io.Read(buf); err != nil {
		return 0, err
	}
	return strconv.ParseInt(string(buf),10,64)
}