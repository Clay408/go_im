# Go语言实现的一个IM即时通信Demo

## 服务端编译
```shell
go build -o server main.go user.go server.go
```

## 客户端编译
```shell
go build -o client client.go
```


## 服务端启动
```shell
./server
```

## 客户端启动
```shell
# 客户端默认连接 127.0.0.1:8888 服务端
./client -ip 127.0.0.1 -port 8888
```