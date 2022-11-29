package main

import (
	"bytes"
	"context"
	"flag"
	"log"
	"os/exec"
	"strings"

	"github.com/svenwiltink/reversegrpc"
	pb "github.com/svenwiltink/reversegrpc/examples/shell/protos"
	"google.golang.org/grpc"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func main() {
	flag.Parse()

	lis := reversegrpc.NewDialListener(*addr)

	s := grpc.NewServer()
	pb.RegisterExecutorServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type server struct {
	pb.UnimplementedExecutorServer
}

func (s *server) Exec(ctx context.Context, in *pb.ExecRequest) (*pb.ExecResponse, error) {
	log.Printf("running %s %s", in.Command, strings.Join(in.Args, " "))
	var stdout, stderr bytes.Buffer
	e := exec.Command(in.Command, in.Args...)
	e.Stdout = &stdout
	e.Stderr = &stderr

	err := e.Run()
	if err != nil {
		return nil, err
	}

	return &pb.ExecResponse{
		Exitcode: int64(e.ProcessState.ExitCode()),
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}, nil

}
