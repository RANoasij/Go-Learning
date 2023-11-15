// What you expect in the server.go?

package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int
	// 在线用户的列表
	Onlinemap map[string]*User // map是一个key-value结构，key是string类型，value是*User类型
	mapLock   sync.RWMutex     // 读写锁 (在sync里有很多锁的实现)
	// 消息广播的channel
	Message chan string
}

// N 大写，表示public，此方法对外可用
func NewServer(ip string, port int) *Server { //* and &
	server := &Server{ // 创建一个server对象，返回一个指针，指向Server的地址? 指向地址可以修改值，而不是拷贝值。
		Ip:   ip,
		Port: port,
		// 初始化
		Onlinemap: make(map[string]*User), // 初始化map，make初始化map，make初始化channel
		Message:   make(chan string),
	}
	return server
}

// 监听Message广播消息channel的goroutine，一旦有消息就发送给全部的在线user .
func (s *Server) ListenMessager() {
	for {
		msg := <-s.Message // 监听管道中的数据，如果有数据，就读取出来，没有数据就阻塞
		// 将msg发送给全部的在线user
		s.mapLock.Lock()                  // 加锁 为啥要加锁呢？因为Onlinemap是线程不安全的，所以要加锁。
		for _, cli := range s.Onlinemap { // cli value, user对象
			cli.C <- msg // 将msg发送给User(cli)用户的管道发过去，随后User(cli)用户的ListenMessage()方法就会读取到msg。详见user.go
		}
		s.mapLock.Unlock() // 解锁
	}
}

// 广播消息的方法
func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg // 组织消息
	s.Message <- sendMsg                                     // 将消息发送到channel中
}

func (s *Server) Handler(conn net.Conn) { // conn is a socket 套接字作为形参
	fmt.Println("连接建立成功")
	// 发回给客户端
	// 用户上线，将用户加入到onlinemap中
	s.mapLock.Lock() // 加锁
	user := NewUser(conn, s)
	s.Onlinemap[user.Name] = user // 将用户加入到onlinemap中
	s.mapLock.Unlock()            // 解锁
	// 广播当前用户上线消息
	s.BroadCast(user, "已上线")
	// 接收客户端消息. 需要Read -> block of Socket. 用goroutine来处理，防止阻塞。go func(){}().
	go func()  {	
		buffer := make([]byte, 1700) // 如果大小写成宏，怎么写： const BUFFER_SIZE = 1024. buffer := make([]byte, BUFFER_SIZE)
		for {
			a, err := conn.Read(buffer) // 读取用户数据. a是读取到的字节数(长度)
			if a == 0 { // 用户下线
				s.BroadCast(user, "下线")
				return
			}
			if err != nil && err != io.EOF { // io.EOF表示读取到文件末尾
				fmt.Println("conn.Read err:", err)
				return
			}
			// 提取用户的消息(去除'\n')
			msg := string(buffer[:a-1]) // 从buffer中取出用户输入的消息
			
		
	}

	// 当前handler阻塞
	select {} // 阻塞当前handler
}

// 启动服务器的接口
func (s *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port)) // 监听协议udp/tcp, 定义一个格式化的字符串，s.Ip和s.Port分别替换%s和%d
	if err != nil {                                                        // go语言中没有异常机制，使用error来原地处理错误
		fmt.Println("net.Listen err:", err)
		return
	}
	// close listen socket
	defer listener.Close()

	// 启动监听Message的goroutine
	go s.ListenMessager()

	// 在循环中。服务器，需要阻塞监听用户连接请求。如果有客户端连接，创建一个连接。
	for {
		// accept
		conn, err := listener.Accept() // 接收连接 //basic socket programming
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			continue //如果出错，继续下一个循环
		}
		// do handler 业务回调处理
		go s.Handler(conn)

	}
}
