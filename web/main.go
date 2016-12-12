package main

import (
	"context"
	"flag"
	"html/template"
	"log"
	"net/http"

	"google.golang.org/grpc"

	apb "github.com/bamnet/apartment/proto/apartment"
)

var (
	srvAddr = flag.String("apt_server", "localhost:10000", "Host of the apartment server.")

	templates = template.Must(template.ParseFiles("index.html"))
)

var client apb.ApartmentClient

func toggleHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "missing name param", http.StatusBadRequest)
		return
	}
	log.Printf("toggling state for: %s", name)
	d, err := client.GetDevice(context.Background(), &apb.GetDeviceRequest{Name: name})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	d.State = !d.State
	d, err = client.UpdateDevice(context.Background(), &apb.UpdateDeviceRequest{
		Device: d,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	p := struct {
		Devices []*apb.Device
	}{}

	resp, _ := client.ListDevices(context.Background(), &apb.ListDevicesRequest{})
	p.Devices = resp.Device

	if err := templates.ExecuteTemplate(w, "index.html", p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	conn, err := grpc.Dial(*srvAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()
	client = apb.NewApartmentClient(conn)

	http.HandleFunc("/toggle", toggleHandler)
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(":8080", nil)
}
