package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
	}
	return server
}

func (this *Server) doHandler(conn net.Conn) {
	fmt.Println("建立连接成功.......")
}

func (this *Server) Start() {
	//开启一个监听服务
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err: ", err)
		return
	}
	//关闭监听
	defer listener.Close()

	for {
		//accept
		//等待获取客户端连接的套接字(Socket)
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listen accept err: ", err)
			continue
		}
		//do handler
		go this.doHandler(conn)
	}
}
