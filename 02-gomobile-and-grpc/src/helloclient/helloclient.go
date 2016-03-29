package helloclient

import (
	pb "github.com/kenshaw/go-jakarta/02-gomobile-and-grpc/src"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// A wrapper type to expose via gomobile.
type HelloClient struct {
	conn   *grpc.ClientConn
	client pb.HelloServiceClient
}

// New creates a new HelloClient with the endpoint addr.
func New(addr string) (*HelloClient, error) {
	var err error

	// create connection
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	// return our "object"
	return &HelloClient{
		conn:   conn,
		client: pb.NewHelloServiceClient(conn),
	}, nil
}

// SayHello calls HelloClient.
func (hc *HelloClient) SayHello(s string) (string, error) {
	// create request
	req := &pb.HelloRequest{Greeting: s}

	// call method
	res, err := hc.client.SayHello(context.Background(), req)
	if err != nil {
		return "", err
	}

	return res.Reply, nil
}

// Shutdown closes connections.
func (hc *HelloClient) Shutdown() error {
	return hc.conn.Close()
}
