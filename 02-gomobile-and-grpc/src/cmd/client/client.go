package main

// import the grpc service definitions
import (
	"fmt"
	"log"

	pb "github.com/go-jakarta/slides/02-gomobile-and-grpc/src"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {
	var err error

	// connect to server
	conn, err := grpc.Dial("localhost:8833", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
	defer conn.Close()

	// create client
	client := pb.NewHelloServiceClient(conn)

	// create request
	req := &pb.HelloRequest{Greeting: "ken"}

	// call method
	res, err := client.SayHello(context.Background(), req)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	// handle response
	fmt.Printf("Received: \"%s\"\n", res.Reply)
}
