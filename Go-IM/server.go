// What you expect in the server.go?

package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int
	// 在线用户的列表
	OnlineMap map[string]*User // map是一个key-value结构，key是string类型，value是*User类型
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
		OnlineMap: make(map[string]*User), // 初始化map，make初始化map，make初始化channel
		Message:   make(chan string),
	}
	return server
}

// 监听Message广播消息channel的goroutine，一旦有消息就发送给全部的在线user .
func (s *Server) ListenMessager() {
	for {
		msg := <-s.Message // 监听管道中的数据，如果有数据，就读取出来，没有数据就阻塞
		// 将msg发送给全部的在线user
		// 需要做一些处理，用户下线后，不能给用户发消息了，不要给关闭的管道发消息。
		s.mapLock.Lock()                  // 加锁 为啥要加锁呢？因为Onlinemap是线程不安全的，所以要加锁。
		for _, cli := range s.OnlineMap { // cli value, user对象
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
	user := NewUser(conn, s)
	user.Online()
	//监听用户是否活跃的channel管道
	isLive := make(chan bool)
	// 接收客户端消息. 需要Read -> block of Socket. 用goroutine来处理，防止阻塞。go func(){}().
	go func() {
		defer conn.Close()           // 关闭连接
		buffer := make([]byte, 6666) // 如果大小写成宏，怎么写： const BUFFER_SIZE = 1024. buffer := make([]byte, BUFFER_SIZE)
		for {
			n, err := conn.Read(buffer) // 读取用户数据. a是读取到的字节数(长度)
			if n == 0 {                 // 用户下线
				user.Offline()
				return
			}
			if err != nil && err != io.EOF { // io.EOF表示读取到文件末尾
				fmt.Println("conn.Read err:", err)
				return
			}
			// 提取用户的消息(去除'\n')
			msg := string(buffer[:n-1]) // 从buffer中取出用户输入的消息. 注意这里如果用Windows系统，需要将n-1改成n-2，因为Windows系统的换行符是\r\n，而Linux系统的换行符是\n。拉跨难搞的Windows.
			// 将得到的消息进行广播
			user.DoMessage(msg)
			isLive <- true // 有消息，就将isLive管道中写入true
			fmt.Println("用户输入的消息:", msg)
		}
	}()

	// 当前handler阻塞
	for {
		select {
		case <-isLive:
		case <-time.After(time.Second * 600):
			user.SendMsg("你被踢了")
			conn.Close()
			user.Offline()
			return
		}
	}
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
