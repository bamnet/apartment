package main

import (
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	apb "github.com/bamnet/apartment/proto/apartment"
)

type server struct{}

func (s *server) GetDevice(ctx context.Context, in *apb.GetDeviceRequest) (*apb.Device, error) {
	return &apb.Device{Name: "water"}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":10000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	apb.RegisterApartmentServer(s, &server{})
	s.Serve(lis)
}
