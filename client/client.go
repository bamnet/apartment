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

	for i := 0; i < 50; i++ {
		device, err := c.GetDevice(context.Background(), &apb.GetDeviceRequest{Name: "Water"})
		if err != nil {
			log.Fatalf("get device error: %v", err)
		}
		log.Printf("response: %v", device)

		device.State = !device.State
		device, err = c.UpdateDevice(context.Background(), &apb.UpdateDeviceRequest{
			Device: device,
		})
		if err != nil {
			log.Fatalf("unable to update device: %v", err)
		}
	}
}
