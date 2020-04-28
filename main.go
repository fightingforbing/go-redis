package main

import (
	"ddvideo/go_redis_server/consts"
	"ddvideo/go_redis_server/server"
	"fmt"
	"os"
)


func init() {
	server.InitServerConfig()
}

func main() {
	help()
	server.Run()
}


func help() {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "-v", "--version":
			version()
		case "-h", "--help":
			usage()
		}
	}
}

func version() {
	fmt.Printf("Redis server v=%s", consts.RedisVersion)
	os.Exit(0)
}

func usage() {
	fmt.Println("Usage: ./redis-server [/path/to/redis.conf] [options]")
	fmt.Println("\t./redis-server - (read config from stdin)")
	fmt.Println("\t./redis-server -v or --version")
	fmt.Println("\t./redis-server -h or --help")
	fmt.Println("Examples:")
	fmt.Println("\t./redis-server (run the server with default conf)")
	fmt.Println("\t./redis-server /etc/redis/6379.conf")
	fmt.Println("\t./redis-server --port 7777")
	fmt.Println("\t./redis-server /etc/myredis.conf --loglevel verbose")
	os.Exit(0)
}
