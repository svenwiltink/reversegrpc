package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	pb "reverse-grpc/helloworld"
	"time"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

// this is a TCP SERVER listening on port 50051 that will behave like a grpc CLIENT afterwards
func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatal(err)
	}

	clientID := 0

	for {
		tcpconn, err := listener.Accept()
		if err != nil {
			log.Fatal("error accepting connection", err)
		}

		clientID++

		// completely disregard context timeouts
		conn, err := grpc.Dial(*addr, grpc.WithInsecure(), grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			log.Println("client is 'dialing'")
			return tcpconn, nil
		}))

		go func(clientID int) {
			if err != nil {
				log.Fatalf("did not connect: %v", err)
			}
			defer conn.Close()
			c := pb.NewGreeterClient(conn)

			for range time.NewTicker(5 * time.Second).C {
				r, err := c.SayHello(context.TODO(), &pb.HelloRequest{Name: fmt.Sprintf("client %d", clientID)})
				if err != nil {
					log.Fatalf("could not greet: %v", err)
				}
				log.Printf("Greeting: %s", r.GetMessage())
			}
		}(clientID)
	}
}
