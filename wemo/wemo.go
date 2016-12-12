// Package wemo controls Belkin WeMo devices using their SOAP API.
package wemo

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/huin/goupnp"
)

const (
	setupURL = "http://%s/setup.xml"
	stateURL = "http://%s/upnp/control/basicevent1"
)

// Device models a WeMo device.
type Device struct {
	Host         string
	FriendlyName string
}

type deviceData struct {
	FriendlyName string `xml:"friendlyName"`
}

// NewDevice sets up a new Device instance.
// A connection is made to the device to lookup basic properties.
func NewDevice(host string) (*Device, error) {
	url := fmt.Sprintf(setupURL, host)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	data := struct {
		Device deviceData `xml:"device"`
	}{}
	err = xml.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return &Device{
		Host:         host,
		FriendlyName: data.Device.FriendlyName,
	}, nil
}

// DiscoverDevices finds all the Wemo Switch or Insight Switches on the network.
func DiscoverDevices() ([]*Device, error) {
	devices := []*Device{}

	types := []string{
		"urn:Belkin:device:insight:1",
		"urn:Belkin:device:controllee:1",
	}
	for _, t := range types {
		hosts, err := goupnp.DiscoverDevices(t)
		if err != nil {
			return nil, err
		}
		for _, host := range hosts {
			d, err := NewDevice(host.Location.Host)
			if err != nil {
				log.Printf("unable to connect to %s: %v", host.Location.Host, err)
				continue
			}
			devices = append(devices, d)
		}
	}
	return devices, nil
}

const getBinaryStateMsg = `
<?xml version="1.0" encoding="utf-8"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
  <s:Body>
    <u:GetBinaryState xmlns:u="urn:Belkin:service:basicevent:1"></u:GetBinaryState>
  </s:Body>
</s:Envelope>
`

var getStateRe = regexp.MustCompile(`.*<BinaryState>(\d+)</BinaryState>.*`)

// State gets the state of the Device.
// An error is returned if the state cannot be looked up.
func (d *Device) State() (bool, error) {
	msg := bytes.NewBuffer([]byte(getBinaryStateMsg))
	url := fmt.Sprintf(stateURL, d.Host)
	req, err := http.NewRequest("POST", url, msg)
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("SOAPACTION", `"urn:Belkin:service:basicevent:1#GetBinaryState"`)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	matches := getStateRe.FindStringSubmatch(string(rbody))
	return strconv.ParseBool(matches[1])
}

const setBinaryStateMsg = `
<?xml version="1.0" encoding="utf-8"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
  <s:Body>
    <u:SetBinaryState xmlns:u="urn:Belkin:service:basicevent:1">
      <BinaryState>%d</BinaryState>
    </u:SetBinaryState>
  </s:Body>
</s:Envelope>
`

// SetState sets the state of the device.
// An error is returned if it fails to do so. In this author's experience,
// errors are fairly common so retry logic should be used.
func (d *Device) SetState(state bool) error {
	i := 0
	if state {
		i = 1
	}
	msg := bytes.NewBuffer([]byte(fmt.Sprintf(setBinaryStateMsg, i)))
	url := fmt.Sprintf(stateURL, d.Host)
	req, err := http.NewRequest("POST", url, msg)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("SOAPACTION", `"urn:Belkin:service:basicevent:1#SetBinaryState"`)

	client := &http.Client{}
	_, err = client.Do(req)
	return err
}
