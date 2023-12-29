package main

import (
	"zinx/znet"
)

/*
	j基于Zinx框架来开发的 服务器应用程序
*/

func main() {
	// 1创建一个Server句柄，使用Zinx的api
	s := znet.NewServer("[zinx V0.1]")
	// 2启动server
	s.Serve()
}
