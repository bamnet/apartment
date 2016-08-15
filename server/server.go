package main

import (
	"fmt"
	"log"

	"github.com/bamnet/apartment/wemo"
	"github.com/cenk/backoff"
	"golang.org/x/net/context"

	apb "github.com/bamnet/apartment/proto/apartment"
)

// Server holds the internal device connections.
type Server struct {
	devices map[string]*wemo.Device
}

// NewServer builds a new Apartment server.
// It connects and maps the initial set of devices.
func NewServer() (*Server, error) {
	aSrv := &Server{devices: map[string]*wemo.Device{}}
	hosts := []string{"192.168.1.187:49153"}
	for _, h := range hosts {
		d, err := wemo.NewDevice(h)
		if err != nil {
			log.Printf("unable to connect to %s: %v", h, err)
			continue
		}
		aSrv.devices[d.FriendlyName] = d
	}
	return aSrv, nil
}

// GetDevice gets the latest information about a Device.
func (s *Server) GetDevice(ctx context.Context, in *apb.GetDeviceRequest) (*apb.Device, error) {
	d, err := s.lookupDevice(in.Name)
	if err != nil {
		return nil, err
	}

	return apiDevice(d)
}

// UpdateDevice sets the state of a Device.
func (s *Server) UpdateDevice(ctx context.Context, in *apb.UpdateDeviceRequest) (*apb.Device, error) {
	d, err := s.lookupDevice(in.Device.Name)
	if err != nil {
		return nil, err
	}

	if err := backoff.Retry(func() error {
		return d.SetState(in.Device.State)
	}, backoff.NewExponentialBackOff()); err != nil {
		return nil, err
	}

	return apiDevice(d)
}

// lookupDevice is a shortcut function to try and find a device in
// the internal device map.
func (s *Server) lookupDevice(name string) (*wemo.Device, error) {
	d, ok := s.devices[name]
	if !ok {
		return nil, fmt.Errorf("no device found")
	}
	return d, nil
}

// apiDevice converts a wemo.Device to an apartment protobuf Device.
func apiDevice(d *wemo.Device) (*apb.Device, error) {
	device := &apb.Device{Name: d.FriendlyName}

	if err := backoff.Retry(func() error {
		state, err := d.State()
		if err != nil {
			return err
		}
		device.State = state
		return nil
	}, backoff.NewExponentialBackOff()); err != nil {
		return nil, err
	}
	return device, nil
}
