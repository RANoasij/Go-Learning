package main

// This project is an IM based on Go language.
// It has ten version, first version is a simple server and client.
// 1. base server
// 2. user online and broadcast
// 3. user message broadcast
// 4. user abstract
// 5. online user list search
// 6. modify user name
// 7. private chat
// 8. offline message
// 9. timeout force kick offline
// 10. client interface

func main() {
	// run server.go
	server := NewServer("127.0.0.1", 8888)
	server.Start()
}

// how to go run
// go run main.go server.go user.go
/*
// I will explain how to do 2. user online and broadcast
// 1. user online.
给Server增加一个Onlinemap，用于记录在线用户。以及Message channel，用于传递用户发送的消息。
OnlineMap 记录在线用户
user.Name(key), user 用户对象
name1 , user1 当user1上线时，将user1加入到OnlineMap中。
name2 , user2
name3 , user3

客户端如何表示呢？
用user1, user2, user3类型表示，因为它包含了用户的信息。包含conn.
每个用户对象绑定一个channel. (是一个go routine不断监听这个channel管道) 用来给客户端发送消息。
type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}
其他用户上线时，需要通知其他用户。所以需要一个全局的channel，用于广播用户上线消息。
type Message struct {
	Content string
	User    *User
}

// 2. broadcast
*/
