package main

import (
	"fmt"
	"net"
	"sync"
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

	//监听服务器广播消息
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
	go user.Online()
	//接收客户端发送的广播消息
	go user.HandleClientPubMessage()
	//当前Handler阻塞
	select {}
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
