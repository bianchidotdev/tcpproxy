# TCPProxy

A configurable raw TCP proxy.

This is a simple TCP proxy that handles multiple app configs, each with multiple listeners and multiple ports.

## Features
* Will listen on multiple ports, forwarding TCP connections for multiple apps with multiple targets
* If the app has multiple targets, it will load balance among them
* If the proxy is unable to connect to the target, it will try the other targets until all are exhausted

## Installation

Install it locally:

``` sh
go install github.com/michaeldbianchi/tcpproxy@latest
```

Or clone it and run it

``` sh
go build ./...
```

## Usage

Run the proxy directly or with the `go run`` command:

``` sh
$ tcpproxy
2023/04/02 14:43:04 Starting a proxy with the following config:
2023/04/02 14:43:04 - Name: five-thousand - Ports: [5001 5200 5300 5400] - Targets: [tcp-echo.fly.dev:5001 tcp-echo.fly.dev:5002]
2023/04/02 14:43:04 - Name: six-thousand - Ports: [6001 6200 6300 6400] - Targets: [tcp-echo.fly.dev:6001 tcp-echo.fly.dev:6002 bad.target.for.testing:6003]
2023/04/02 14:43:04 - Name: seven-thousand - Ports: [7001 7200 7300 7400] - Targets: [tcp-echo.fly.dev:7001 tcp-echo.fly.dev:7002]
2023/04/02 14:43:04 - Name: broken - Ports: [8001] - Targets: [bad.target.for.testing:6003 bad.target.for.testing:6003 bad.target.for.testing:6003 bad.target.for.testing:6003]
 
$ tcpproxy -c ./other_config.yaml

# OR

$ go run ./...
```


Test it out with a simple netcat command:

``` sh
echo hello | timeout 1 nc localhost 5001
```


## Appendix

### How this works

There are several parts of the application:

- Config - Reads in a JSON config file at a specified location (default to `./config.json`) and parses into a `ProxyConfig` struct
- Proxy - Contains a lightweight factory to create a proxy object based on the config struct and implementation for functions that enable the proxy to listen on a given port and to proxy connections (more details to follow)
- Main - Parses cli args, reads config, spins up proxies, captures ctrl-c input to close listeners

More details on the proxies functioning:

- There's a top-level `WaitGroup` to ensure we wait for all the goroutines running proxies to close before we stop the program

- A new proxy is spun up in a new Goroutine for each port (not app) since each has to listen on a given port

- Another Goroutine is spun up for each individual TCP client connection to facilitate proxying the TCP streams back and forth while leaving the listener open to accept new connections

- For a given TCP client connection, the following happens:

  - We try to open a TCP connection with the application's target host (the current behavior is to pick a random starting target, this could just as easily have been round-robin any other strategy)

  - If we successfully open a connection to the target host

    - We spin up 2 more Goroutines with a `WaitGroup` to copy bytes from the client to the server and back asynchronously
    - Once either the client or server closes it's connection, we then ensure both connections are closed
  - If we cannot open a connection to the target host
    - We try all other hosts in the target list, one by one and sequentially, looking for a healthy target
    - If we find one, we continue with that host, if not, we close the client connection and log an error


I used goroutines for the proxies, connections, and stream copying because they were easy and sufficient for this local proxy. It might not scale for production.

Health-checking can be an art in itself to get done right, so in this example, I didn't cut implement any form of target eliminations (ie. cutting out targets because of n failed connections). Instead, I just always iterate through all targets until we've tried them all. I also didn't do retrying of a particular target. This is definitely a brute-force strategy that adds latency as we have more bad targets in the target pool, but the alternative would require a lot more state management (keeping track of successes and failures on a rolling basis) than is strictly necessary for a lightweight proxy.
