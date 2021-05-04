package main

import (
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/go-jakarta/slides/16-grpc/src"
)

// START OMIT
type myhandler struct{}

func (h *myhandler) Echo(ctx context.Context, msg *pb.EchoMessage) (*pb.EchoMessage, error) {
	return msg, nil
}

// END OMIT

func main() {
	lis, err := net.Listen("tcp", ":8079")
	if err != nil {
		log.Fatal(err)
	}

	srv := grpc.NewServer()
	pb.RegisterUtilServiceServer(srv, &myhandler{})
	srv.Serve(lis)
}
