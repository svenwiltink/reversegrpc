package main

import (
	"context"
	"flag"
	"log"

	"github.com/svenwiltink/reversegrpc"
	pb "github.com/svenwiltink/reversegrpc/examples/echoer/protos"
	"google.golang.org/grpc"
)

// this is a TCP CLIENT that will behave like a grpc SERVER one it has dialed out
func main() {
	flag.Parse()

	lis := reversegrpc.NewDialListener("localhost:50051")

	s := grpc.NewServer()
	pb.RegisterEchoerServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedEchoerServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) Echo(ctx context.Context, in *pb.EchoRequest) (*pb.EchoReply, error) {
	log.Printf("Received: %v", in.GetMessage())
	return &pb.EchoReply{Message: in.GetMessage()}, nil
}
