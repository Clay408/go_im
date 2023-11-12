package main

import (
	"fmt"
	"io"
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
	//启动监听Message的goroutine
	go s.ListenMessage()

	for {
		//accept
		//等待获取客户端连接的套接字(Socket)
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
	user := NewUser(conn)
	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user //记录在线用户
	s.mapLock.Unlock()
	//广播上线消息
	s.BroadCast(user, "已上线")

	//接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				delete(s.OnlineMap, user.Name)
				s.BroadCast(user, "下线")
				fmt.Println("当前在线人数：", len(s.OnlineMap))
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err : ", err)
				return
			}

			//提取出用户消息
			msg := string(buf[:n-1])
			//广播消息
			s.BroadCast(user, msg)
		}
	}()

	//当前Handler阻塞
	select {}
}

// ListenMessage 监听服务端 Message广播消息channel的goroutine，一旦有消息就发送给全部的在线User
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message // 没有消息时当前goroutine会阻塞在这里
		//将msg发送给全部在线的User
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// BroadCast 广播消息
func (s *Server) BroadCast(user *User, msg string) {
	msg = "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- msg
}
