package main

import (
	"fmt"
	"net"
)

type User struct {
	Name string
	Addr string
	C    chan string
	Conn net.Conn

	server *Server //当前用户是属于哪个服务器
}

func NewUser(server *Server, conn net.Conn) *User {
	user := &User{
		Name:   conn.RemoteAddr().String(),
		Addr:   conn.RemoteAddr().String(),
		C:      make(chan string),
		Conn:   conn,
		server: server,
	}
	//启动监听当前user channel消息的goroutine
	go user.ListenMessage()
	return user
}

// SendMsg 给当前用户发送消息
func (u *User) SendMsg(msg string) {
	u.Conn.Write([]byte(msg + "\n"))
}

// Online 用户上线功能
func (u *User) Online() {
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u //服务端保存在线用户
	u.server.mapLock.Unlock()
	//广播上线消息
	u.server.BroadCast(u, "已上线")
}

// Offline 用户下线功能
func (u *User) Offline() {
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()
	u.server.BroadCast(u, "下线")
	fmt.Println("当前在线人数：", len(u.server.OnlineMap))
}

// ListenMessage 监听当前User channel 的方法，一旦有消息，就直接发送给客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.SendMsg(msg)
	}
}

// DoMessage 客户端消息处理
func (u *User) DoMessage(msg string) {
	if "who" == msg {
		//查询在线用户
		for _, user := range u.server.OnlineMap {
			res := "[" + user.Addr + "]" + user.Name + "-> 在线"
			u.SendMsg(res)
		}
	} else {
		u.server.BroadCast(u, msg)
	}
}
