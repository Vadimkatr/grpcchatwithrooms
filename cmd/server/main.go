package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/Vadimkatr/grpcchatwithrooms/internal/apiserver/server"
	pb "github.com/Vadimkatr/grpcchatwithrooms/internal/proto"
)

func main() {

	srv, _ := server.InitServer()

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("error creating the server %v", err)
	}

	log.Println("Starting server at port :8080")
	pb.RegisterChatRoomsServer(grpcServer, srv)
	grpcServer.Serve(listener)
}
