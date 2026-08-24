// 最小可运行的 gRPC Demo —— 服务端 + 客户端共用一个 proto 生成包。
// 运行方式（在 gRPC 目录下）：
//   终端1：go run . server   # 启动 gRPC 服务，监听 :50051
//   终端2：go run . client   # 调用 SayHello，打印 "Hello, world"
//
// 依赖：google.golang.org/grpc + google.golang.org/protobuf
//   go get google.golang.org/grpc
//   go mod tidy
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"gRPC-demo/hellopb"
)

type greeterServer struct {
	hellopb.UnimplementedGreeterServer
}

// SayHello 实现 proto 中定义的 Unary 接口
func (s *greeterServer) SayHello(ctx context.Context, req *hellopb.HelloRequest) (*hellopb.HelloReply, error) {
	// 服务端感知客户端超时：ctx 到期就返回 DeadlineExceeded
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.DeadlineExceeded, "deadline exceeded")
	case <-time.After(100 * time.Millisecond):
		return &hellopb.HelloReply{Message: "Hello, " + req.GetName()}, nil
	}
}

func runServer(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	hellopb.RegisterGreeterServer(s, &greeterServer{})

	log.Printf("gRPC server listening on %s", addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func runClient(addr string) {
	// 本地演示用不安全连接；生产环境务必用 TLS（见知识库正文第七章）
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := hellopb.NewGreeterClient(conn)

	// 务必设置超时，否则可能永久阻塞
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	reply, err := c.SayHello(ctx, &hellopb.HelloRequest{Name: "world"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	fmt.Println(reply.GetMessage())
}

func main() {
	mode := flag.String("mode", "", "server 或 client")
	addr := flag.String("addr", ":50051", "监听地址")
	flag.Parse()

	switch *mode {
	case "server":
		runServer(*addr)
	case "client":
		runClient(*addr)
	default:
		log.Fatal("用法: go run . -mode server | go run . -mode client")
	}
}
