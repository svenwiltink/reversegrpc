package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"sync"

	"github.com/svenwiltink/reversegrpc"
	pb "github.com/svenwiltink/reversegrpc/examples/shell/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type CNC struct {
	Workers    map[int]pb.ExecutorClient
	mut        *sync.Mutex
	Controller *reversegrpc.Controller
}

func (c *CNC) Run() {
	err := c.Controller.Listen(*addr)
	if err != nil {
		panic(err)
	}

	clientID := 0

	for {
		clientConn, err := c.Controller.Accept(grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			log.Fatalf("did not 'connect': %v", err)
		}

		clientID++

		c.handleConn(clientConn, clientID)
	}
}

func (c *CNC) Exec(cmd string) {
	c.mut.Lock()
	defer c.mut.Unlock()

	for clientID, worker := range c.Workers {
		response, err := worker.Exec(context.Background(), &pb.ExecRequest{Command: "sh", Args: []string{"-c", cmd}})
		if err != nil {
			stat, _ := status.FromError(err)
			if stat.Code() == codes.Canceled {
				log.Println("worker disconnected, removing connection")
				delete(c.Workers, clientID)
				return
			}
			log.Fatalln(err)
		}

		log.Printf("client %d response: ExitCode %d Output %#v StdErr %#v", clientID, response.Exitcode, response.Stdout, response.Stderr)
	}
}

func (c *CNC) handleConn(clientConn *grpc.ClientConn, clientID int) {
	log.Println("client connected")

	c.mut.Lock()
	defer c.mut.Unlock()

	client := pb.NewExecutorClient(clientConn)
	c.Workers[clientID] = client
}

var (
	addr = flag.String("addr", "localhost:50051", "the address to listen on")
)

// this is a TCP SERVER listening on port 50051 that will behave like a grpc CLIENT afterwards
func main() {
	flag.Parse()

	controller := reversegrpc.Controller{}
	cnc := CNC{
		Workers:    make(map[int]pb.ExecutorClient),
		Controller: &controller,
		mut:        &sync.Mutex{},
	}

	go cnc.Run()

	reader := bufio.NewScanner(os.Stdin)
	for reader.Scan() {
		cmd := reader.Text()
		cnc.Exec(cmd)
	}

}
