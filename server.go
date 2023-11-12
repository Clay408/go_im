package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	Message   chan string
	mapLock   sync.RWMutex
}

// NewServer 创建服务端实例
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// Start 开启服务端
func (s *Server) Start() {
	//开启一个监听服务
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err: ", err)
		return
	}
	//关闭监听
	defer listener.Close()

	//监听广播消息
	go s.listenBroadCastMessage()

	for {
		//accept
		//等待客户端连接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listen accept err: ", err)
			continue
		}
		//do handler
		go s.handler(conn)
	}
}

// 处理客户端连接
func (s *Server) handler(conn net.Conn) {
	user := NewUser(s, conn)
	//用户上线
	user.Online()
	//客户端消息监听
	go s.listenPubMsgFromClient(user)
	for {
		select {
		case <-user.isAlive:
			//啥都不用做
		case <-time.After(time.Second * 30): //30秒不说话自动踢
			user.SendMsg("你被踢了")
			close(user.C)
			//关闭连接
			conn.Close()
			return
		}
	}
}

// 客户端消息监听
func (s *Server) listenPubMsgFromClient(u *User) {
	buf := make([]byte, 4096)
	for {
		n, err := u.Conn.Read(buf) //没有消息在这里阻塞
		if n == 0 {
			u.Offline()
			return
		}

		if err != nil && err != io.EOF {
			fmt.Println("Conn Read err:", err)
			return
		}

		//提取出用户消息
		msg := string(buf[:n-1])
		u.DoMessage(msg)
		u.isAlive <- 1
	}
}

// BroadCast 广播消息
func (s *Server) BroadCast(user *User, msg string) {
	msg = "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- msg
}

// ListenBroadCastMessage 服务端监听广播消息
func (s *Server) listenBroadCastMessage() {
	for {
		msg := <-s.Message // 没有消息时当前goroutine会阻塞在这里
		//将msg发送给全部在线的User
		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.C <- msg
		}
		s.mapLock.Unlock()
	}
}
