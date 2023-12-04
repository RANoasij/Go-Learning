package main

import (
	"net"
	"strings"
)

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

func (this *User) Online() {
	// 用户上线，将用户加入到onlinemap中
	this.server.mapLock.Lock()              // 加锁
	this.server.OnlineMap[this.Name] = this // 将用户加入到onlinemap中
	this.server.mapLock.Unlock()            // 解锁
	// 广播当前用户上线消息
	this.server.BroadCast(this, "已上线")
}
func (this *User) Offline() {
	// User goes offline, remove from OnlineMap 	// 用户下线，将用户从onlinemap中删除
	this.server.mapLock.Lock() // 加锁
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock() // 解锁
	// Broadcast user offline message
	this.server.BroadCast(this, "已下线")
	// Safely close the user's message channel
	if this.C != nil {
		this.CloseMessageChannel()
	}
}

func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户有哪些
		for _, user := range this.server.OnlineMap { //注意这里this是User结构体的实例。this.server就是Server的实例，也就有OnlineMap属性记着用户列表。
			onlineMsd := "[" + user.Addr + "]" + user.Name + ":" + "在线\n"
			this.SendMsg(onlineMsd) //注意这里只发送给当前用户，不广播。所以是conn.Write([]byte(msg))写给当前用户User结构体的conn属性。
		}
	} else if len(msg) >= 7 && msg[:7] == "rename " { //这里还要判断下len(msg) > 6，不然会报错的。
		newName := strings.Split(msg, " ")[1]     // 通过空格分割，取第二个元素，也就是新名字。
		_, flag := this.server.OnlineMap[newName] // Go语言里，访问map会返回两个值，第一个是对应的值，第二个是这个key是否存在。这里只要判断第二个值就可以了。
		if flag {                                 // flag == true
			this.SendMsg("当前用户名已经被使用\n")
		} else {
			this.server.mapLock.Lock()                      // 改之前先上锁。OnlineMap不是线程安全的。
			delete(this.server.OnlineMap, this.Name)        // 删除原来的名字
			this.server.OnlineMap[newName] = this           // 加入新名字 //回忆OnlineMap的定义：map[string]*User，也就是 ['名字']: User结构体的实例
			this.server.mapLock.Unlock()                    // 解锁
			this.Name = newName                             // 改名字
			this.SendMsg("老铁，已经改好了，现在你叫：" + newName + "\n") // 发送给当前用户 //注意不是server.BroadCast(this, "xxx")，而是this.SendMsg("xxx")，因为只发给当前用户。
		}
	} else if len(msg) >= 4 && msg[:3] == "to " {
		// 消息格式： to username content
		// 1. 获取对方的用户名
		remoteName := strings.Split(msg, " ")[1]
		if remoteName == "" {
			this.SendMsg("消息格式不正确，请使用 \"to 用户名 消息内容\"格式\n")
			return
		}
		// 2. 根据用户名得到对方User对象
		remoteUser, ok := this.server.OnlineMap[remoteName] // OnlineMap里存着对面的User结构体的实例
		if !ok {
			this.SendMsg("该用户名不存在\n")
			return
		}
		// 3. 获取消息内容，通过对方的User对象将消息内容发送过去
		content := strings.Split(msg, " ")[2]
		if content == "" {
			this.SendMsg("啥都没写，请重发\n")
			return
		}
		remoteUser.SendMsg("[悄悄话]" + this.Name + "对你说：" + content + " :比心: " + "\n")
	} else {
		this.server.BroadCast(this, msg)
	}
}

// 这里改了个[BUG]，防止用户发送空消息，导致服务器死循环
// 监听当前user channel的方法，一旦有消息，就直接发送给对端客户端。
func (a *User) ListenMessage() { //一般最好叫一个别名。不过a和this是一样的。
	for msg := range a.C { // 监听管道中的数据，如果有数据，就读取出来，没有数据就阻塞
		a.conn.Write([]byte(msg + "\n")) // 转成byte二进制类型，发送给客户端
	}
}

// Proper way to close a channel, the key to take away is to set the channel to nil to indicate it's closed
func (this *User) CloseMessageChannel() {
	this.server.mapLock.Lock()
	defer this.server.mapLock.Unlock()
	close(this.C)
	this.C = nil // Set the channel to nil to indicate it's closed
}
