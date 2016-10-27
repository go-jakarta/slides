package main

import (
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/kenshaw/go-jakarta/05-rest-to-grpc/src"
)

type myhandler struct {
}

func (h *myhandler) Echo(ctx context.Context, msg *pb.EchoMessage) (*pb.EchoMessage, error) {
	return msg, nil
}

func main() {
	lis, err := net.Listen("tcp", ":8079")
	if err != nil {
		log.Fatal(err)
	}

	srv := grpc.NewServer()
	pb.RegisterUtilServiceServer(srv, &myhandler{})
	srv.Serve(lis)
} // END
