package main

import (
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	apb "github.com/bamnet/apartment/proto/apartment"
)

const (
	address = "localhost:10000"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()
	c := apb.NewApartmentClient(conn)

	resp, err := c.GetDevice(context.Background(), &apb.GetDeviceRequest{})
	if err != nil {
		log.Fatalf("get device error: %v", err)
	}
	log.Printf("response: %v", resp)
}
