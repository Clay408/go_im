package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name    string
	Addr    string
	C       chan string
	Conn    net.Conn
	isAlive chan int

	server *Server //当前用户是属于哪个服务器
}

func NewUser(server *Server, conn net.Conn) *User {
	user := &User{
		Name:    conn.RemoteAddr().String(),
		Addr:    conn.RemoteAddr().String(),
		C:       make(chan string),
		Conn:    conn,
		isAlive: make(chan int),
		server:  server,
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
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式：rename|张三
		newName := strings.Split(msg, "|")[1]
		//判断name是否存在
		_, ok := u.server.OnlineMap[newName]
		if ok {
			//用户已经存在
			u.SendMsg(newName + "已存在")
		} else {
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name) //删除旧的
			u.server.OnlineMap[newName] = u    //添加新的
			u.server.mapLock.Unlock()
			u.Name = newName
			u.SendMsg("修改名称成功：" + u.Name)
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//消息格式:  to|张三|消息内容
		toUserName := strings.Split(msg, "|")[1]
		if toUserName == "" {
			u.SendMsg("消息格式不正确，请使用\"to|张三|你好啊\"格式。\n")
			return
		}

		toUser, ok := u.server.OnlineMap[toUserName]
		if !ok {
			u.SendMsg("用户不存在\n")
			return
		}

		content := strings.Split(msg, "|")[2]
		if content == "" {
			u.SendMsg("无消息内容，请重发\n")
			return
		}
		toUser.SendMsg(u.Name + "对您说：" + content)
	} else {
		u.server.BroadCast(u, msg)
	}
}
