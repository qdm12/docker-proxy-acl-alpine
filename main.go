package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/kyokomi/emoji"
	"github.com/namsral/flag"
)

type upStream struct {
	name   string
	handle *http.Client
}

func (r upStream) pass() func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			http.Error(res, "400 Bad request ; only GET allowed.", 400)
			return
		}
		param := ""
		if len(req.URL.RawQuery) > 0 {
			param = "?" + req.URL.RawQuery
		}
		body, _ := r.get("http://docker"+req.URL.Path+param, res)
		fmt.Fprintf(res, "%s", body)
	}
}

func (r upStream) get(url string, res http.ResponseWriter) ([]byte, error) {
	req, err := r.handle.Get(url)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	contentType := req.Header.Get("Content-type")
	if contentType != "" {
		res.Header().Set("Content-type", contentType)
	}
	return ioutil.ReadAll(req.Body)
}

func newProxySocket(socket string) upStream {
	stream := upStream{name: socket}
	stream.handle = &http.Client{
		Transport: &http.Transport{
			Dial: func(proto, addr string) (net.Conn, error) {
				conn, err := net.Dial("unix", socket)
				return conn, err
			},
		},
	}
	return stream
}

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%d", *s)
}

func (s *stringSlice) Set(value string) error {
	log.Println(emoji.Sprint(":heavy_check_mark:") + "Allowing endpoint " + value)
	*s = append(*s, value)
	return nil
}

var allowedOptions = []string{"containers", "images", "volumes", "services", "tasks", "events", "version", "info", "ping"}

func main() {
	fmt.Println("############################################")
	fmt.Println("############# Docker Proxy ACL #############")
	fmt.Println("############# by Quentin McGaw #############")
	fmt.Println("############### Give some " + emoji.Sprint(":heart:") + "at ###############")
	fmt.Println("# github.com/qdm12/docker-proxy-acl-alpine #")
	fmt.Print("##############################################\n\n")
	fs := flag.NewFlagSetWithEnvPrefix(os.Args[0], "GO", 0)
	var (
		allowed  stringSlice
		filename = fs.String("filename", "/tmp/docker-proxy-acl/docker.sock", "Location of socket file")
	)
	fs.Var(&allowed, "a", "Allowed location pattern prefix")
	fs.Parse(os.Args[1:])
	if len(allowed) < 1 {
		log.Fatal(emoji.Sprint(":x:") + "Need at least 1 argument for -a: [" + strings.Join(allowedOptions, ",") + "]")
	}
	var ok bool
	for i := range allowed {
		ok = false
		for j := range allowedOptions {
			if allowed[i] == allowedOptions[j] {
				ok = true
				break
			}
		}
		if !ok {
			log.Fatal(emoji.Sprint(":x:") + "Argument " + allowed[i] + " not recognized")
		}
	}
	var routers [2]*mux.Router
	routers[0] = mux.NewRouter()
	routers[1] = routers[0].PathPrefix("/{version:[v][0-9]+[.][0-9]+}").Subrouter()
	upstream := newProxySocket("/var/run/docker.sock")
	for i := range allowed {
		log.Println(emoji.Sprint(":registered:") + "Registering " + allowed[i] + " handlers...")
		for _, m := range routers {
			switch allowed[i] {
			case "containers":
				containers := m.PathPrefix("/containers").Subrouter()
				containers.HandleFunc("/json", upstream.pass())
				containers.HandleFunc("/{name}/json", upstream.pass())
			case "images":
				images := m.PathPrefix("/images").Subrouter()
				images.HandleFunc("/json", upstream.pass())
				images.HandleFunc("/{name}/json", upstream.pass())
				images.HandleFunc("/{name}/history", upstream.pass())
			case "volumes":
				m.HandleFunc("/volumes", upstream.pass())
				m.HandleFunc("/volumes/{name}", upstream.pass())
			case "networks":
				m.HandleFunc("/networks", upstream.pass())
				m.HandleFunc("/networks/{name}", upstream.pass())
			case "services":
				m.HandleFunc("/services", upstream.pass())
				m.HandleFunc("/services/{name}", upstream.pass())
			case "tasks":
				m.HandleFunc("/tasks", upstream.pass())
				m.HandleFunc("/tasks/{name}", upstream.pass())
			case "events":
				m.HandleFunc("/events", upstream.pass())
			case "version":
				m.HandleFunc("/version", upstream.pass())
			case "info":
				m.HandleFunc("/info", upstream.pass())
			case "ping":
				m.HandleFunc("/_ping", upstream.pass())
			}
		}
	}
	http.Handle("/", routers[0])
	listener, err := net.Listen("unix", *filename)
	if err != nil {
		log.Fatal(err)
	}
	os.Chmod(*filename, 0666)
	// Looking up group ids coming up for Go 1.7
	// https://github.com/golang/go/issues/2617
	log.Println("Listening on " + *filename + emoji.Sprint(" :ear:"))
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
	go waitForSignal(sigc, listener)
	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func waitForSignal(c chan os.Signal, listener net.Listener) {
	sig := <-c
	log.Println(emoji.Sprint(":heavy_exclamation_mark:") + "Caught signal '" + sig.String() + "', shutting down")
	listener.Close()
	os.Exit(0)
}