# Simple Load Balancer
This Go package implements a simple load balancer with round-robin server selection and basic health checks.

## Features
- Round-robin load balancing
- Health checks for backend servers
- Reverse proxy to forward requests to backend servers

## Installation
- Ensure you have Go installed on your system. You can download it from golang.org.
- Clone the repository or download the source code.
```sh
git clone https://github.com/arnab-afk/loadbalancer.git
cd loadbalancer
```

## Run
```sh
go run .
```

## Configuration
To customize the server list or the port, modify the main function in main.go:

```go
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
```


## Code Overview
`simpleServer`
This type represents a backend server.

- `Address() string:` Returns the server address.
- `IsAlive() bool:` Checks if the server is alive by sending an HTTP GET request and checking the response status.
- `Serve(rw http.ResponseWriter, req *http.Request):` Forwards the request to the backend server using a reverse proxy.
___

`LoadBalancer` This type represents the load balancer.
- `getNextAvailableServer() Server`: Implements the round-robin algorithm to select the next available server.
- `serveProxy(rw http.ResponseWriter, req *http.Request):` Forwards the request to the selected backend server.

___

## Example
Here is an example of how the load balancer works:
```go
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
```

In this example, the load balancer will start on port `8009` and forward incoming requests to one of the servers in the list (`https://www.facebook.com/`, `https://www.bing.com/`, `https://www.duckduckgo.com/`) using a round-robin algorithm. If a server is not alive, it will be skipped, and the next available server will be selected.

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.
