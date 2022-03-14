/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"flag"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc"
	"log"
	"net"
	pb "reverse-grpc/helloworld"
	"time"
)

// this is a TCP CLIENT that will behave like a grpc SERVER one it has dialed out
func main() {
	flag.Parse()

	lis := &DialListener{sema: semaphore.NewWeighted(1)}

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

type callbackCloser struct {
	net.Conn
	callback func()
}

func (c callbackCloser) Close() error {
	c.callback()
	return c.Conn.Close()
}

// who needs to listen anyways. just dial :svekw:
type DialListener struct {
	sema *semaphore.Weighted
}

func (d *DialListener) Accept() (net.Conn, error) {
	d.sema.Acquire(context.Background(), 1)

	for {
		log.Println("server is dialing out, what a weird world")
		conn, err := net.Dial("tcp", "localhost:50051")
		if err == nil {
			return callbackCloser{
				Conn: conn,
				callback: func() {
					log.Println("conn closed, releasing semaphore")
					d.sema.Release(1)
				},
			}, nil
		}

		log.Println("error dialing out", err)

		time.Sleep(time.Second)
	}
}

// nothing to close when you aren't listening
func (d *DialListener) Close() error {
	return nil
}

func (d DialListener) Addr() net.Addr {
	return &net.IPAddr{
		IP: net.IPv4(127, 0, 0, 1),
	}
}
