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

type UpStream struct {
	Name   string
	handle *http.Client
}

func (r UpStream) Pass() func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			http.Error(res, "400 Bad request ; only GET allowed.", 400)
			return
		}
		param := ""
		if len(req.URL.RawQuery) > 0 {
			param = "?" + req.URL.RawQuery
		}
		body, _ := r.Get("http://docker"+req.URL.Path+param, res)
		fmt.Fprintf(res, "%s", body)
	}
}

func (r UpStream) Get(url string, res http.ResponseWriter) ([]byte, error) {
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

func newProxySocket(socket string) UpStream {
	stream := UpStream{Name: socket}
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
		log.Println(emoji.Sprint(":x:") + "Need at least 1 argument for -a: [" + strings.Join(allowedOptions, ",") + "]")
		os.Exit(0)
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
			log.Println(emoji.Sprint(":x:") + "Argument " + allowed[i] + " not recognized!")
			os.Exit(0)
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
				containers.HandleFunc("/json", upstream.Pass())
				containers.HandleFunc("/{name}/json", upstream.Pass())
			case "images":
				containers := m.PathPrefix("/images").Subrouter()
				containers.HandleFunc("/json", upstream.Pass())
				containers.HandleFunc("/{name}/json", upstream.Pass())
				containers.HandleFunc("/{name}/history", upstream.Pass())
			case "volumes":
				m.HandleFunc("/volumes", upstream.Pass())
				m.HandleFunc("/volumes/{name}", upstream.Pass())
			case "networks":
				m.HandleFunc("/networks", upstream.Pass())
				m.HandleFunc("/networks/{name}", upstream.Pass())
			case "services":
				m.HandleFunc("/services", upstream.Pass())
				m.HandleFunc("/services/{name}", upstream.Pass())
			case "tasks":
				m.HandleFunc("/tasks", upstream.Pass())
				m.HandleFunc("/tasks/{name}", upstream.Pass())
			case "events":
				m.HandleFunc("/events", upstream.Pass())
			case "version":
				m.HandleFunc("/version", upstream.Pass())
			case "info":
				m.HandleFunc("/info", upstream.Pass())
			case "ping":
				m.HandleFunc("/_ping", upstream.Pass())
			}
		}
	}
	http.Handle("/", routers[0])
	listener, err := net.Listen("unix", *filename)
	os.Chmod(*filename, 0666)
	// Looking up group ids coming up for Go 1.7
	// https://github.com/golang/go/issues/2617
	log.Println("Listening on " + *filename + emoji.Sprint(" :ear:"))
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func(c chan os.Signal) {
		sig := <-c
		log.Println(emoji.Sprint(":heavy_exclamation_mark:")+"Caught signal %s, shutting down", sig)
		listener.Close()
		os.Exit(0)
	}(sigc)
	if err != nil {
		panic(err)
	}
	err = http.Serve(listener, nil)
	if err != nil {
		panic(err)
	}
}
