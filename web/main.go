package main

import (
	"context"
	"flag"
	"html/template"
	"log"
	"net/http"
	"sort"

	"google.golang.org/grpc"

	apb "github.com/bamnet/apartment/proto/apartment"
)

var (
	srvAddr = flag.String("apt_server", "localhost:10000", "Host of the apartment server.")

	templates = template.Must(template.ParseFiles("index.html"))
)

var client apb.ApartmentClient

// ByFriendlyName sorts devices by their friendly name.
type ByFriendlyName []*apb.Device

func (s ByFriendlyName) Len() int {
	return len(s)
}
func (s ByFriendlyName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByFriendlyName) Less(i, j int) bool {
	return s[i].FriendlyName < s[j].FriendlyName
}

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
	sort.Sort(ByFriendlyName(p.Devices))

	if err := templates.ExecuteTemplate(w, "index.html", p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func init() {
	flag.Parse()
}

func main() {
	conn, err := grpc.Dial(*srvAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()
	client = apb.NewApartmentClient(conn)

	http.Handle("/node_modules/", http.StripPrefix("/node_modules/", http.FileServer(http.Dir("./node_modules"))))
	http.HandleFunc("/toggle", toggleHandler)
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(":8080", nil)
}
