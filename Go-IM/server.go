// What you expect in the server.go?
package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

func NewServer(ip string, port int) *Server { //* and &
	server := &Server{ // 创建一个server对象，返回一个指针，指向Server的地址? 指向地址可以修改值，而不是拷贝值。
		Ip:   ip,
		Port: port,
	}
	return server
}

// Explain Handler function

func (s *Server) Handler(conn net.Conn) { // 传递一个conn对象
	fmt.Println("连接建立成功")
}

// 启动服务器的接口
func (s *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port)) // 监听端口
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	// close listen socket
	defer listener.Close()

	for {
		// accept
		conn, err := listener.Accept() // 接收连接
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			continue
		}
		// do handler
		go s.Handler(conn) // 开启一个goroutine处理连接
		// I could not understand Handler and its relation with Struct Server
		// Let me explain: Handler is a method of Server struct, so it can access Server's field and method. And it can be called by Server object. So, it is a method of Server.
		// Continue. what's more about Handler? It is a function, and it has a parameter conn. conn is a net.Conn object, which is a connection between client and server.
	}
}
