package main

// import the hello.proto grpc service definitions
import (
	pb "github.com/kenshaw/go-jakarta/02-gomobile-and-grpc/src"
)

import (
	"fmt"
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type HelloServer struct{}

// SayHello says 'hi' to the user.
func (hs *HelloServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	// create response
	res := &pb.HelloResponse{
		Reply: fmt.Sprintf("hello %s from go", req.Greeting),
	}

	return res, nil
}

func main() {
	var err error

	// create socket listener
	l, err := net.Listen("tcp", ":8833")
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	// create server
	helloServer := &HelloServer{}

	// register server with grpc
	s := grpc.NewServer()
	pb.RegisterHelloServiceServer(s, helloServer)

	// run
	s.Serve(l)
}
