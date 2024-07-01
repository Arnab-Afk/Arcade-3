package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"time"
)

type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
	alive bool
	mu    sync.RWMutex
}

type Server interface {
	Address() string
	IsAlive() bool
	Serve(rw http.ResponseWriter, r *http.Request)
	CheckHealth()
}

func newSimpleServer(addr string) *simpleServer {
	serverURL, err := url.Parse(addr)
	handleError(err)

	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverURL),
		alive: true,
	}
}

type LoadBalancer struct {
	port       string
	serverList []Server
	mu         sync.Mutex
}

func newLoadBalancer(port string, serverList []Server) *LoadBalancer {
	lb := &LoadBalancer{
		port:       port,
		serverList: serverList,
	}

	go lb.healthCheck() // Start health checks in a separate goroutine

	return lb
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
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.alive
}

func (s *simpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("Forwarding request to %s\n", s.addr)
	s.proxy.ServeHTTP(rw, req)
}

func (s *simpleServer) CheckHealth() {
	for {
		resp, err := http.Get(s.addr)
		alive := err == nil && resp.StatusCode == http.StatusOK
		s.mu.Lock()
		s.alive = alive
		s.mu.Unlock()
		time.Sleep(10 * time.Second) // Check every 10 seconds
	}
}

func (lb *LoadBalancer) healthCheck() {
	var wg sync.WaitGroup
	for _, server := range lb.serverList {
		wg.Add(1)
		go func(s Server) {
			defer wg.Done()
			s.CheckHealth()
		}(server)
	}
	wg.Wait()
}

func (lb *LoadBalancer) getNextAvailableServer() Server {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	for _, server := range lb.serverList {
		if server.IsAlive() {
			return server
		}
	}
	return nil
}

func (lb *LoadBalancer) serveProxy(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("Received request: %s\n", req.URL.Path)
	for {
		targetServer := lb.getNextAvailableServer()
		if targetServer != nil {
			targetServer.Serve(rw, req)
			return
		}
		fmt.Println("No available servers, retrying...")
		time.Sleep(1 * time.Second) // Wait before retrying
	}
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
