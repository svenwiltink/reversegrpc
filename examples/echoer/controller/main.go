package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/svenwiltink/reversegrpc"
	pb "github.com/svenwiltink/reversegrpc/examples/echoer/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to listen on")
)

// this is a TCP SERVER listening on port 50051 that will behave like a grpc CLIENT afterwards
func main() {
	flag.Parse()

	controller := reversegrpc.Controller{}
	controller.Listen(*addr)
	clientID := 0

	for {
		clientConn, err := controller.Accept(grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			log.Fatalf("did not 'connect': %v", err)
		}

		clientID++

		go handleConn(clientConn, clientID)
	}
}

func handleConn(clientConn *grpc.ClientConn, clientID int) {
	defer clientConn.Close()
	c := pb.NewEchoerClient(clientConn)

	for range time.NewTicker(5 * time.Second).C {
		r, err := c.Echo(context.TODO(), &pb.EchoRequest{Message: fmt.Sprintf("sup client %d", clientID)})
		if err != nil {
			stat, _ := status.FromError(err)
			if stat.Code() == codes.Canceled {
				log.Println("worker disconnected, removing connection")
				return
			}
			log.Fatalln(err)
		}

		if err != nil {
			log.Fatalf("could not echo: %v", err)
		}

		log.Printf("Echo response: %s", r.GetMessage())
	}
}
