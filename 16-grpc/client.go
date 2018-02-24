package main

import (
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "gophers.id/slides/16-grpc/src"
)

func main() {
	conn, err := grpc.Dial("localhost:8079", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	client := pb.NewUtilServiceClient(conn)
	res, err := client.Echo(context.Background(), &pb.EchoMessage{
		Msg: "hello!",
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("response: %+v", res)
}
