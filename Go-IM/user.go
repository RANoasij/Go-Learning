package main

import "net"

type User struct {
	Name   string
	Addr   string
	C      chan string // channel 管道
	conn   net.Conn    // socket 套接字
	server *Server
}

// 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String() // 获取用户的网络地址
	// 常规初始化一个用户实例
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string), // make初始化channel
		conn:   conn,
		server: server,
	}
	// 启动监听当前user channel消息的goroutine
	go user.ListenMessage()
	return user
}

// 监听当前user channel的方法，一旦有消息，就直接发送给对端客户端。
func (u *User) ListenMessage() {
	for {
		msg := <-u.C                     // 监听管道中的数据，如果有数据，就读取出来，没有数据就阻塞
		u.conn.Write([]byte(msg + "\n")) // 转成byte二进制类型，发送给客户端
	}
}
