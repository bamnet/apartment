package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bamnet/apartment/wemo"
	"github.com/cenk/backoff"
	"golang.org/x/net/context"

	apb "github.com/bamnet/apartment/proto/apartment"
)

// Number of scans in-a-row a device must be not found to be removed.
const missingThreshold = 5

// Server holds the internal device connections.
type Server struct {
	devices map[string]*wemo.Device
	missing map[string]int

	mutex *sync.Mutex
}

// NewServer builds a new Apartment server.
// It connects and maps the initial set of devices.
func NewServer() (*Server, error) {
	aSrv := &Server{
		devices: map[string]*wemo.Device{},
		missing: map[string]int{},
		mutex:   &sync.Mutex{},
	}
	if err := aSrv.mapDevices(); err != nil {
		return nil, err
	}
	aSrv.remapper(60 * time.Second)

	return aSrv, nil
}

func (s *Server) mapDevices() error {
	devices, err := wemo.DiscoverDevices()
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	found := map[string]bool{} // Keys of newly found devices.
	for _, d := range devices {
		key := rename(d.FriendlyName)
		s.devices[key] = d
		found[key] = true
		delete(s.missing, key) // Remove the device from the missing map.
	}

	// Loop through all the existing devices, see if we found them during the
	// latest scan. Increase the missing count of those not found.
	for key := range s.devices {
		if _, exists := found[key]; !exists {
			s.missing[key]++
		}
	}

	// Loop through the missing devices and remove them if missing for too long.
	for key, count := range s.missing {
		if count > missingThreshold {
			delete(s.devices, key)
			delete(s.missing, key)
		}
	}
	return nil
}

func (s *Server) remapper(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.mapDevices()
			}
		}
	}()
}

// ListDevices lists all the devices the server is aware of.
// It does not attempt to identify the state of the devices.
func (s *Server) ListDevices(ctx context.Context, _ *apb.ListDevicesRequest) (*apb.ListDevicesResponse, error) {
	resp := apb.ListDevicesResponse{}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for n, d := range s.devices {
		device := &apb.Device{
			Name:         n,
			FriendlyName: d.FriendlyName,
		}
		resp.Device = append(resp.Device, device)
	}
	return &resp, nil
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
	s.mutex.Lock()
	defer s.mutex.Unlock()
	d, ok := s.devices[name]
	if !ok {
		return nil, fmt.Errorf("no device found")
	}
	return d, nil
}

func rename(in string) string {
	return strings.ToLower(in)
}

// apiDevice converts a wemo.Device to an apartment protobuf Device.
func apiDevice(d *wemo.Device) (*apb.Device, error) {
	device := &apb.Device{
		Name:         rename(d.FriendlyName),
		FriendlyName: d.FriendlyName,
	}

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
