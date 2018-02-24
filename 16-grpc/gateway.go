package main

import (
	"log"
	"net/http"

	"golang.org/x/net/context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	pb "gophers.id/slides/16-grpc/src"
)

func main() {
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := pb.RegisterUtilServiceHandlerFromEndpoint(ctxt, mux, "localhost:8079", opts)
	if err != nil {
		log.Fatal(err)
	}

	http.ListenAndServe(":8080", mux)
}
