package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	apb "github.com/bamnet/apartment/proto/apartment"
)

func main() {
	lis, err := net.Listen("tcp", ":10000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	aSrv, err := NewServer()
	if err != nil {
		log.Fatalf("unable to setup apartment server: %v", err)
	}
	apb.RegisterApartmentServer(srv, aSrv)
	srv.Serve(lis)
}
