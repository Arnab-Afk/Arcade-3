package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

type Server interface {
	Address() string
	IsAlive() bool
	Serve(rw http.ResponseWriter, r *http.Request)
}

func newSimpleServer(addr string) *simpleServer {
	serverURL, err := url.Parse(addr)
	handleError(err)

	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverURL),
	}
}

type LoadBalancer struct {
	port            string
	roundRobinIndex int
	serverList      []Server
}

func newLoadBalancer(port string, serverList []Server) *LoadBalancer {
	return &LoadBalancer{
		port:            port,
		roundRobinIndex: 0,
		serverList:      serverList,
	}
}

func handleError(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

func (s *simpleServer) Address() string {
	return s.addr
}

func (s *simpleServer) IsAlive() bool {
	// Implement a real health check here.
	resp, err := http.Get(s.addr)
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func (s *simpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("Forwarding request to %s\n", s.addr)
	s.proxy.ServeHTTP(rw, req)
}

func (lb *LoadBalancer) getNextAvailableServer() Server {
	for {
		server := lb.serverList[lb.roundRobinIndex%len(lb.serverList)]
		lb.roundRobinIndex++
		if server.IsAlive() {
			fmt.Printf("Selected server: %s\n", server.Address())
			return server
		}
	}
}

func (lb *LoadBalancer) serveProxy(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("Received request: %s\n", req.URL.Path)
	targetServer := lb.getNextAvailableServer()
	targetServer.Serve(rw, req)
}

func main() {
	serverList := []Server{
		newSimpleServer("https://www.facebook.com/"),
		newSimpleServer("https://www.bing.com/"),
		newSimpleServer("https://www.duckduckgo.com/"),
	}

	lb := newLoadBalancer("8009", serverList)
	handleRedirect := func(rw http.ResponseWriter, req *http.Request) {
		lb.serveProxy(rw, req)
	}
	http.HandleFunc("/", handleRedirect)
	fmt.Printf("Load Balancer started at :%s\n", lb.port)
	err := http.ListenAndServe(":"+lb.port, nil)
	handleError(err)
}