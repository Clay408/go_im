package main

import (
	"fmt"
	"io"
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

// HandleClientPubMessage 用户广播消息处理业务
func (u *User) HandleClientPubMessage() {
	buf := make([]byte, 4096)
	for {
		n, err := u.Conn.Read(buf)
		if n == 0 {
			u.Offline()
			return
		}
		if err != nil && err != io.EOF {
			fmt.Println("Conn Read err : ", err)
			return
		}

		//提取出用户消息
		msg := string(buf[:n-1])
		//广播消息
		u.server.BroadCast(u, msg)
	}
}

// HandlerClientPriMessage 处理用户私聊的消息
func (u *User) HandlerClientPriMessage() {

}

// ListenMessage 监听当前User channel 的方法，一旦有消息，就直接发送给客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.Conn.Write([]byte(msg + "\n"))
	}
}
